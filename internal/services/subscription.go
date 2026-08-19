package services

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/ofstudio/dancegobot/internal/config"
	"github.com/ofstudio/dancegobot/internal/models"
	"github.com/ofstudio/dancegobot/internal/store"
	"github.com/ofstudio/dancegobot/pkg/noplog"
	"github.com/ofstudio/dancegobot/pkg/trace"
)

var ErrSubscriptionUnavailable = errors.New("subscription is unavailable")

// MembershipFunc gets membership facts for a Telegram user in a chat.
type MembershipFunc func(chatID, userID int64) (models.Membership, error)

// SubscriptionStatus is the current subscription control state for an event.
type SubscriptionStatus struct {
	Available  bool
	Subscribed bool
}

// SubscriptionService manages subscriptions to new event posts.
type SubscriptionService struct {
	store      store.Store
	notifier   *NotifierService
	membership MembershipFunc
	log        *slog.Logger
}

func NewSubscriptionService(
	store store.Store,
	notifier *NotifierService,
	membership MembershipFunc,
) *SubscriptionService {
	return &SubscriptionService{
		store:      store,
		notifier:   notifier,
		membership: membership,
		log:        noplog.Logger(),
	}
}

func (s *SubscriptionService) WithLogger(l *slog.Logger) *SubscriptionService {
	s.log = l
	return s
}

// Status returns the subscription control state for the event and profile.
func (s *SubscriptionService) Status(
	ctx context.Context,
	event *models.Event,
	profile models.Profile,
) (SubscriptionStatus, error) {
	chat, err := subscriptionChat(event)
	if err != nil {
		return SubscriptionStatus{}, nil
	}
	subscription, err := s.store.SubscriptionGet(ctx, models.SubscriptionKey{
		SubscriberID: profile.ID,
		ChatID:       chat.ID,
	})
	if err != nil {
		return SubscriptionStatus{}, fmt.Errorf("failed to get subscription: %w", err)
	}
	if subscription != nil {
		return SubscriptionStatus{Available: true, Subscribed: true}, nil
	}
	if err = s.validateAccess(ctx, event, profile); err != nil {
		if errors.Is(err, ErrSubscriptionUnavailable) {
			return SubscriptionStatus{}, nil
		}
		return SubscriptionStatus{}, err
	}
	return SubscriptionStatus{Available: true}, nil
}

// Subscribe subscribes a profile to new event posts in the event chat.
// Returns created=false if the subscription already exists.
func (s *SubscriptionService) Subscribe(
	ctx context.Context,
	eventID string,
	profile models.Profile,
) (*models.Subscription, bool, error) {
	event, err := s.store.EventGet(ctx, eventID)
	if err != nil {
		return nil, false, fmt.Errorf("failed to get event: %w", err)
	}
	if err = s.validateAccess(ctx, event, profile); err != nil {
		return nil, false, err
	}
	chat := event.Post.Chat
	subscription := &models.Subscription{
		Subscriber: profile,
		Chat:       *chat,
		CreatedAt:  nowFn(),
	}

	tx, err := s.store.Begin(ctx)
	if err != nil {
		return nil, false, fmt.Errorf("failed to begin tx: %w", err)
	}
	defer tx.Rollback()
	created, err := tx.SubscriptionCreate(ctx, subscription)
	if err != nil {
		return nil, false, fmt.Errorf("failed to create subscription: %w", err)
	}
	if !created {
		subscription, err = tx.SubscriptionGet(ctx, subscription.Key())
		if err != nil {
			return nil, false, fmt.Errorf("failed to get existing subscription: %w", err)
		}
	} else if err = tx.HistoryCreate(ctx, &models.HistoryItem{
		Action:    models.HistoryUserSubscribed,
		Initiator: &profile,
		EventID:   &event.ID,
		Details:   subscription,
		CreatedAt: nowFn(),
	}); err != nil {
		return nil, false, fmt.Errorf("failed to create subscription history: %w", err)
	}
	if err = tx.Commit(); err != nil {
		return nil, false, fmt.Errorf("failed to commit tx: %w", err)
	}
	s.log.Info("[subscription service] subscription created",
		"subscription", subscription,
		"created", created,
		trace.Attr(ctx))
	return subscription, created, nil
}

// Unsubscribe removes a profile subscription to the event chat.
// Returns removed=false if the subscription does not exist.
func (s *SubscriptionService) Unsubscribe(
	ctx context.Context,
	eventID string,
	profile models.Profile,
) (*models.Subscription, bool, error) {
	event, err := s.store.EventGet(ctx, eventID)
	if err != nil {
		return nil, false, fmt.Errorf("failed to get event: %w", err)
	}
	if event == nil || event.Post == nil || event.Post.Chat == nil {
		return nil, false, ErrSubscriptionUnavailable
	}

	tx, err := s.store.Begin(ctx)
	if err != nil {
		return nil, false, fmt.Errorf("failed to begin tx: %w", err)
	}
	defer tx.Rollback()
	subscription, err := tx.SubscriptionRemove(ctx, models.SubscriptionKey{
		SubscriberID: profile.ID,
		ChatID:       event.Post.Chat.ID,
	})
	if err != nil {
		return nil, false, fmt.Errorf("failed to remove subscription: %w", err)
	}
	removed := subscription != nil
	if removed {
		if err = tx.HistoryCreate(ctx, &models.HistoryItem{
			Action:    models.HistoryUserUnsubscribed,
			Initiator: &profile,
			EventID:   &event.ID,
			Details:   subscription,
			CreatedAt: nowFn(),
		}); err != nil {
			return nil, false, fmt.Errorf("failed to create unsubscription history: %w", err)
		}
	}
	if err = tx.Commit(); err != nil {
		return nil, false, fmt.Errorf("failed to commit tx: %w", err)
	}
	s.log.Info("[subscription service] subscription removed",
		"subscription", subscription,
		"removed", removed,
		trace.Attr(ctx))
	return subscription, removed, nil
}

// HandleEventPublished notifies eligible subscribers about a published event at most once.
func (s *SubscriptionService) HandleEventPublished(ctx context.Context, event *models.Event) error {
	chat, err := subscriptionChat(event)
	if err != nil {
		return err
	}
	if event.SubscribersNotified {
		return nil
	}
	if s.membership == nil {
		return ErrSubscriptionUnavailable
	}

	subscriptions, err := s.store.SubscriptionGetByChatID(ctx, chat.ID)
	if err != nil {
		return fmt.Errorf("failed to get chat subscriptions: %w", err)
	}
	set, err := s.store.EventSubscribersNotifiedSet(ctx, event.ID)
	if err != nil {
		return fmt.Errorf("failed to set subscribers notified: %w", err)
	}
	if !set {
		return nil
	}
	notificationEvent := cloneEvent(event)
	notificationEvent.SubscribersNotified = true
	go s.notifyEventSubscribers(ctx, notificationEvent, subscriptions)
	return nil
}

func (s *SubscriptionService) notifyEventSubscribers(
	ctx context.Context,
	event *models.Event,
	subscriptions []*models.Subscription,
) {
	chat := event.Post.Chat
	strict, err := s.membershipCheckStrict(ctx, chat.ID)
	if err != nil {
		s.log.Error("[subscription service] failed to determine membership policy: "+err.Error(),
			"chat", chat.LogValue(),
			trace.Attr(ctx))
		return
	}
	for _, subscription := range subscriptions {
		if strict {
			membership, err := s.membershipGet(chat.ID, subscription.Subscriber.ID)
			if err != nil {
				s.log.Error("[subscription service] failed to get subscriber membership: "+err.Error(),
					"subscription", subscription,
					trace.Attr(ctx))
				continue
			}
			if !subscriptionMemberAllowed(*chat, membership) {
				continue
			}
		}
		profile := subscription.Subscriber
		s.notifier.Notify(ctx, &models.Notification{
			TmplCode:  models.TmplNewEvent,
			Recipient: &profile,
			Payload: models.NotificationPayload{
				Event: event,
			},
		})
	}
}

func (s *SubscriptionService) validateAccess(
	ctx context.Context,
	event *models.Event,
	profile models.Profile,
) error {
	chat, err := subscriptionChat(event)
	if err != nil {
		return err
	}
	strict, err := s.membershipCheckStrict(ctx, chat.ID)
	if err != nil {
		return err
	}
	if !strict {
		return nil
	}
	membership, err := s.membershipGet(chat.ID, profile.ID)
	if err != nil {
		return ErrSubscriptionUnavailable
	}
	if !subscriptionMemberAllowed(*chat, membership) {
		return ErrSubscriptionUnavailable
	}
	return nil
}

// membershipCheckStrict reports whether subscriber membership must be verified.
// Telegram only guarantees getChatMember for other users when the bot is an administrator:
// https://core.telegram.org/bots/api#getchatmember. We therefore use strict checks only
// for administrator bots. Non-administrator status or a Telegram error enables relaxed
// mode, accepting disclosure of the chat name, ID, and event publication fact. Revisit
// this policy if Telegram changes the getChatMember guarantees.
func (s *SubscriptionService) membershipCheckStrict(ctx context.Context, chatID int64) (bool, error) {
	if s.membership == nil {
		return false, ErrSubscriptionUnavailable
	}
	botMembership, err := s.membershipGet(chatID, config.BotProfile().ID)
	if err != nil {
		s.log.Warn("[subscription service] failed to get bot membership; using relaxed membership policy: "+err.Error(),
			"chat_id", chatID,
			trace.Attr(ctx))
		return false, nil
	}
	return botMembership.Administrator, nil
}

func (s *SubscriptionService) membershipGet(chatID, userID int64) (models.Membership, error) {
	if s.membership == nil {
		return models.Membership{}, ErrSubscriptionUnavailable
	}
	return s.membership(chatID, userID)
}

func subscriptionChat(event *models.Event) (*models.Chat, error) {
	if event == nil || event.Removed || event.Post == nil || event.Post.Chat == nil ||
		event.Post.ChatMessageID == 0 ||
		(event.Post.Chat.Type != models.ChatSuper && event.Post.Chat.Type != models.ChatChannel) {
		return nil, ErrSubscriptionUnavailable
	}
	return event.Post.Chat, nil
}

func subscriptionMemberAllowed(chat models.Chat, membership models.Membership) bool {
	if chat.Username != "" {
		return !membership.Banned
	}
	return membership.Member
}

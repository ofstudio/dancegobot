package services

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	tele "gopkg.in/telebot.v4"

	"github.com/ofstudio/dancegobot/internal/config"
	"github.com/ofstudio/dancegobot/internal/models"
	storepkg "github.com/ofstudio/dancegobot/internal/store"
)

func TestSubscriptionStatusEligibility(t *testing.T) {
	config.SetBotProfile(&tele.User{ID: 999, FirstName: "Bot"})
	ctx := context.Background()
	service, st := newSubscriptionServiceTest(t, func(_ int64, userID int64) (models.Membership, error) {
		switch userID {
		case 999:
			return models.Membership{Administrator: true, Member: true}, nil
		case 1:
			return models.Membership{Member: true}, nil
		case 2:
			return models.Membership{Banned: true}, nil
		case 3:
			return models.Membership{}, nil
		default:
			return models.Membership{}, errors.New("telegram unavailable")
		}
	}, nil)

	privateEvent := subscriptionTestEvent("private", models.Chat{
		ID: -1001, Type: models.ChatSuper, Title: "Private",
	})
	publicEvent := subscriptionTestEvent("public", models.Chat{
		ID: -1002, Type: models.ChatChannel, Title: "Public", Username: "public_chat",
	})
	require.NoError(t, st.EventUpsert(ctx, privateEvent))
	require.NoError(t, st.EventUpsert(ctx, publicEvent))

	tests := []struct {
		name      string
		event     *models.Event
		profileID int64
		available bool
	}{
		{name: "private member", event: privateEvent, profileID: 1, available: true},
		{name: "private non-member", event: privateEvent, profileID: 3},
		{name: "public non-member", event: publicEvent, profileID: 3, available: true},
		{name: "public banned", event: publicEvent, profileID: 2},
		{name: "membership error", event: privateEvent, profileID: 4},
		{name: "nil event", profileID: 1},
		{name: "missing post", event: &models.Event{}, profileID: 1},
		{name: "basic group", event: subscriptionTestEvent("group", models.Chat{ID: -3, Type: models.ChatGroup}), profileID: 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, err := service.Status(ctx, tt.event, models.Profile{ID: tt.profileID, FirstName: "User"})
			require.NoError(t, err)
			require.Equal(t, tt.available, status.Available)
			require.False(t, status.Subscribed)
		})
	}

	existing := &models.Subscription{
		Subscriber: models.Profile{ID: 3, FirstName: "Former member"},
		Chat:       *privateEvent.Post.Chat,
	}
	created, err := st.SubscriptionCreate(ctx, existing)
	require.NoError(t, err)
	require.True(t, created)
	status, err := service.Status(ctx, privateEvent, existing.Subscriber)
	require.NoError(t, err)
	require.Equal(t, SubscriptionStatus{Available: true, Subscribed: true}, status)
}

func TestSubscriptionSubscribeUnsubscribe(t *testing.T) {
	config.SetBotProfile(&tele.User{ID: 999, FirstName: "Bot"})
	ctx := context.Background()
	accessAllowed := true
	service, st := newSubscriptionServiceTest(t, func(_ int64, userID int64) (models.Membership, error) {
		if userID == 999 {
			return models.Membership{Administrator: true, Member: true}, nil
		}
		if !accessAllowed {
			return models.Membership{}, errors.New("membership unavailable")
		}
		return models.Membership{Member: true}, nil
	}, nil)
	event := subscriptionTestEvent("event", models.Chat{ID: -1001, Type: models.ChatSuper, Title: "Dance"})
	require.NoError(t, st.EventUpsert(ctx, event))
	profile := models.Profile{ID: 1, FirstName: "User"}

	subscription, created, err := service.Subscribe(ctx, event.ID, profile)
	require.NoError(t, err)
	require.True(t, created)
	require.Equal(t, profile.ID, subscription.Subscriber.ID)
	same, created, err := service.Subscribe(ctx, event.ID, profile)
	require.NoError(t, err)
	require.False(t, created)
	require.Equal(t, subscription, same)
	require.Equal(t, 1, historyActionCount(t, st, models.HistoryUserSubscribed))
	accessAllowed = false
	_, _, err = service.Subscribe(ctx, event.ID, profile)
	require.ErrorIs(t, err, ErrSubscriptionUnavailable)

	removedSubscription, removed, err := service.Unsubscribe(ctx, event.ID, profile)
	require.NoError(t, err)
	require.True(t, removed)
	require.Equal(t, subscription, removedSubscription)
	removedSubscription, removed, err = service.Unsubscribe(ctx, event.ID, profile)
	require.NoError(t, err)
	require.False(t, removed)
	require.Nil(t, removedSubscription)
	require.Equal(t, 1, historyActionCount(t, st, models.HistoryUserUnsubscribed))
}

func TestSubscriptionRequiresBotAdministrator(t *testing.T) {
	config.SetBotProfile(&tele.User{ID: 999, FirstName: "Bot"})
	ctx := context.Background()
	service, st := newSubscriptionServiceTest(t, func(_ int64, userID int64) (models.Membership, error) {
		if userID == 999 {
			return models.Membership{Member: true}, nil
		}
		return models.Membership{Member: true}, nil
	}, nil)
	event := subscriptionTestEvent("event", models.Chat{ID: -1001, Type: models.ChatSuper})
	require.NoError(t, st.EventUpsert(ctx, event))

	status, err := service.Status(ctx, event, models.Profile{ID: 1, FirstName: "User"})
	require.NoError(t, err)
	require.Equal(t, SubscriptionStatus{}, status)
	_, _, err = service.Subscribe(ctx, event.ID, models.Profile{ID: 1, FirstName: "User"})
	require.ErrorIs(t, err, ErrSubscriptionUnavailable)
	require.ErrorIs(t, service.HandleEventPublished(ctx, event), ErrSubscriptionUnavailable)
	stored, err := st.EventGet(ctx, event.ID)
	require.NoError(t, err)
	require.False(t, stored.SubscribersNotified)
}

func TestSubscriptionHandleEventPublishedAtMostOnce(t *testing.T) {
	config.SetBotProfile(&tele.User{ID: 999, FirstName: "Bot"})
	ctx := context.Background()
	var mu sync.Mutex
	var notifications []*models.Notification
	service, st := newSubscriptionServiceTest(t, func(_ int64, userID int64) (models.Membership, error) {
		switch userID {
		case 999:
			return models.Membership{Administrator: true, Member: true}, nil
		case 1:
			return models.Membership{Member: true}, nil
		case 2:
			return models.Membership{}, nil
		default:
			return models.Membership{}, errors.New("telegram unavailable")
		}
	}, func(notification *models.Notification) error {
		mu.Lock()
		defer mu.Unlock()
		notifications = append(notifications, notification)
		return nil
	})
	event := subscriptionTestEvent("event", models.Chat{ID: -1001, Type: models.ChatSuper, Title: "Dance"})
	require.NoError(t, st.EventUpsert(ctx, event))
	for id := int64(1); id <= 3; id++ {
		created, err := st.SubscriptionCreate(ctx, &models.Subscription{
			Subscriber: models.Profile{ID: id, FirstName: "User"},
			Chat:       *event.Post.Chat,
		})
		require.NoError(t, err)
		require.True(t, created)
	}

	require.NoError(t, service.HandleEventPublished(ctx, event))
	require.Eventually(t, func() bool {
		mu.Lock()
		defer mu.Unlock()
		return len(notifications) == 1
	}, time.Second, 10*time.Millisecond)
	require.NoError(t, service.HandleEventPublished(ctx, event))
	time.Sleep(20 * time.Millisecond)

	mu.Lock()
	require.Len(t, notifications, 1)
	require.Equal(t, models.TmplNewEvent, notifications[0].TmplCode)
	require.Equal(t, int64(1), notifications[0].Recipient.ID)
	require.Equal(t, event.ID, notifications[0].Payload.Event.ID)
	mu.Unlock()
	storedEvent, err := st.EventGet(ctx, event.ID)
	require.NoError(t, err)
	require.True(t, storedEvent.SubscribersNotified)
	require.Eventually(t, func() bool {
		return historyActionCount(t, st, models.HistoryNotificationSent) == 1
	}, time.Second, 10*time.Millisecond)
}

func TestSubscriptionHandleEventPublishedClaimsEmptyFanout(t *testing.T) {
	config.SetBotProfile(&tele.User{ID: 999, FirstName: "Bot"})
	ctx := context.Background()
	service, st := newSubscriptionServiceTest(t, func(_ int64, _ int64) (models.Membership, error) {
		return models.Membership{Administrator: true, Member: true}, nil
	}, nil)
	event := subscriptionTestEvent("event", models.Chat{ID: -1001, Type: models.ChatSuper})
	require.NoError(t, st.EventUpsert(ctx, event))
	require.NoError(t, service.HandleEventPublished(ctx, event))
	storedEvent, err := st.EventGet(ctx, event.ID)
	require.NoError(t, err)
	require.True(t, storedEvent.SubscribersNotified)
}

func newSubscriptionServiceTest(
	t *testing.T,
	membership MembershipFunc,
	notify NotifyFunc,
) (*SubscriptionService, *storepkg.SQLiteStore) {
	t.Helper()
	cfg := config.Default()
	db, err := storepkg.NewSQLite(":memory:", cfg.DB.Version)
	require.NoError(t, err)
	st := storepkg.NewSQLiteStore(db)
	t.Cleanup(st.Close)
	if notify == nil {
		notify = func(*models.Notification) error { return nil }
	}
	notifier := NewNotifierService(cfg.Settings, st, notify)
	return NewSubscriptionService(st, notifier, membership), st
}

func subscriptionTestEvent(id string, chat models.Chat) *models.Event {
	return &models.Event{
		ID:      id,
		Caption: "Event",
		Owner:   models.Profile{ID: 100, FirstName: "Owner"},
		Post: &models.Post{
			InlineMessageID: "inline_" + id,
			Chat:            &chat,
			ChatMessageID:   10,
		},
	}
}

func historyActionCount(t *testing.T, st *storepkg.SQLiteStore, action models.HistoryAction) int {
	t.Helper()
	var count int
	require.NoError(t, st.DB().QueryRow(`SELECT COUNT(*) FROM history WHERE action = ?1`, action).Scan(&count))
	return count
}

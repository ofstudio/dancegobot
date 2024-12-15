package services

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/ofstudio/dancegobot/internal/config"
	"github.com/ofstudio/dancegobot/internal/models"
	"github.com/ofstudio/dancegobot/pkg/noplog"
	"github.com/ofstudio/dancegobot/pkg/randtoken"
	"github.com/ofstudio/dancegobot/pkg/trace"
)

var nowFn = func() time.Time {
	return time.Now().UTC()
}

// EventService is a service that manages dance events
type EventService struct {
	cfg      config.Settings
	store    Store
	notifier *NotifierService
	renderer *RenderService
	log      *slog.Logger
}

func NewEventService(cfg config.Settings, store Store, r *RenderService, n *NotifierService) *EventService {
	return &EventService{
		cfg:      cfg,
		store:    store,
		renderer: r,
		notifier: n,
		log:      noplog.Logger(),
	}
}

func (s *EventService) WithLogger(l *slog.Logger) *EventService {
	s.log = l
	return s
}

// NewID generates a new event ID
func (s *EventService) NewID() string {
	return randtoken.New(s.cfg.EventIDLen)
}

// Create creates a new event.
func (s *EventService) Create(ctx context.Context, event *models.Event) error {
	if err := s.validateEvent(event); err != nil {
		return fmt.Errorf("failed to validate event: %w", err)
	}

	if err := s.store.EventUpsert(ctx, event); err != nil {
		return fmt.Errorf("failed to upsert event: %w", err)
	}

	go s.historyInsert(ctx, &models.HistoryItem{
		Action:    models.HistoryEventCreated,
		Initiator: &event.Owner,
		EventID:   &event.ID,
		Details:   event,
		CreatedAt: nowFn(),
	})

	return nil
}

// Get returns an event by ID.
func (s *EventService) Get(ctx context.Context, id string) (*models.Event, error) {
	event, err := s.store.EventGet(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get event: %w", err)
	}
	return event, nil
}

// RegistrationGet returns registration for the given event by profile and role.
func (s *EventService) RegistrationGet(event *models.Event, profile *models.Profile, role models.Role) *models.Registration {
	return NewEventHandler(event).RegistrationGet(&models.Dancer{
		Profile:   profile,
		FullName:  profile.FullName(),
		Role:      role,
		CreatedAt: nowFn(),
	})
}

// PostAdd adds information about the post where the event is published.
func (s *EventService) PostAdd(
	ctx context.Context,
	eventID string,
	inlineMessageID string,
) (*models.Event, *models.Post, error) {
	if inlineMessageID == "" {
		return nil, nil, fmt.Errorf("inline message ID must be provided")
	}

	var event *models.Event
	post := &models.Post{InlineMessageID: inlineMessageID}
	err := s.handle(ctx, eventID, func(h *EventHandler) {
		h.Event().Post = post
		event = h.Event()
		h.hist = append(h.hist, &models.HistoryItem{
			Action:    models.HistoryPostAdded,
			Initiator: &h.event.Owner,
			EventID:   &h.event.ID,
			Details:   h.event.Post,
		})
	})
	return event, post, err
}

// PostChatAdd adds information about a chat where the event is published.
func (s *EventService) PostChatAdd(
	ctx context.Context,
	eventID string,
	chat *models.Chat,
	chatMessageID int,
) (*models.Event, *models.Post, error) {
	if chat == nil {
		return nil, nil, fmt.Errorf("chat must be provided")
	}
	if chatMessageID == 0 {
		return nil, nil, fmt.Errorf("chat message ID must be provided")
	}

	// Update the event
	var event *models.Event
	var post *models.Post
	err := s.handle(ctx, eventID, func(h *EventHandler) {
		if h.Event().Post == nil {
			h.Event().Post = &models.Post{}
		}
		h.Event().Post.Chat = chat
		h.Event().Post.ChatMessageID = chatMessageID
		event = h.Event()
		post = h.Event().Post
		h.hist = append(h.hist, &models.HistoryItem{
			Action:    models.HistoryPostChatAdded,
			Initiator: &h.event.Owner,
			EventID:   &h.event.ID,
			Details:   chat,
		})
	})
	return event, post, err
}

// CoupleAdd registers a couple for the event.
// If the partner initially was registered as a single, the partner will be notified.
// The partner can be either specified by a profile or a full name.
func (s *EventService) CoupleAdd(
	ctx context.Context,
	eventID string,
	profile *models.Profile,
	role models.Role,
	other any,
) (*models.Registration, error) {

	if err := s.validateProfile(profile); err != nil {
		return nil, fmt.Errorf("failed to validate profile: %w", err)
	}

	dancer := &models.Dancer{
		Profile:   profile,
		FullName:  profile.FullName(),
		Role:      role,
		CreatedAt: nowFn(),
	}
	var partner *models.Dancer
	switch v := other.(type) {
	case *models.Profile:
		if err := s.validateProfile(v); err != nil {
			return nil, fmt.Errorf("failed to validate other person profile: %w", err)
		}
		partner = &models.Dancer{
			Profile:   v,
			FullName:  v.FullName(),
			Role:      role.Opposite(),
			CreatedAt: nowFn(),
		}
	case string:
		if err := s.validateFullname(v); err != nil {
			return nil, fmt.Errorf("failed to validate other person name: %w", err)
		}
		partner = &models.Dancer{
			FullName:  v,
			Role:      role.Opposite(),
			CreatedAt: nowFn(),
		}
	default:
		return nil, fmt.Errorf("invalid type of other person: %T", other)
	}

	var reg *models.Registration
	err := s.handle(ctx, eventID, func(h *EventHandler) {
		reg = h.CoupleAdd(dancer, partner)
	})
	return reg, err
}

// SingleAdd adds a single dancer to the event.
// If auto pair is enabled, tries to pair the dancer with another single dancer.
func (s *EventService) SingleAdd(
	ctx context.Context,
	eventID string,
	profile *models.Profile,
	role models.Role,
) (*models.Registration, error) {
	if err := s.validateProfile(profile); err != nil {
		return nil, fmt.Errorf("failed to validate profile: %w", err)
	}
	var reg *models.Registration
	err := s.handle(ctx, eventID, func(h *EventHandler) {
		reg = h.SingleAdd(&models.Dancer{
			Profile:   profile,
			FullName:  profile.FullName(),
			Role:      role,
			CreatedAt: nowFn(),
		})
	})
	return reg, err
}

// DancerRemove removes the dancer from the event.
//
// If the dancer is in a couple, and the partner initially signed up as a single,
// the partner will be moved to the singles list back (or auto paired if enabled)
// and a notification will be created.
// Otherwise, the partner will be removed from the event as well.
//
// If the dancer is in a couple, and a couple was created by the partner,
// the partner will be notified that the dancer has left the event.
func (s *EventService) DancerRemove(
	ctx context.Context,
	eventID string,
	profile *models.Profile,
) (*models.Registration, error) {
	if err := s.validateProfile(profile); err != nil {
		return nil, fmt.Errorf("failed to validate profile: %w", err)
	}
	var reg *models.Registration
	err := s.handle(ctx, eventID, func(h *EventHandler) {
		reg = h.DancerRemove(&models.Dancer{
			Profile:  profile,
			FullName: profile.FullName(),
		})
	})
	return reg, err
}

// handle is a wrapper for the event handler.
func (s *EventService) handle(
	ctx context.Context,
	eventID string,
	handlerFunc func(*EventHandler),
) error {
	// Begin tx
	tx, err := s.store.BeginTx(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin tx: %w", err)
	}
	//goland:noinspection ALL
	defer tx.Rollback()

	// Get event
	event, err := tx.EventGet(ctx, eventID)
	if err != nil {
		return fmt.Errorf("failed to get event: %w", err)
	}

	// Create event handler
	handler := NewEventHandler(event)

	// Run update function
	handlerFunc(handler)

	// After event handling is done, we need to:
	// - upsert the event in the store
	// - commit the transaction
	// - render event post
	// - add history items
	// - send the notifications
	if err = tx.EventUpsert(ctx, handler.Event()); err != nil {
		return fmt.Errorf("failed to upsert event: %w", err)
	}
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit tx: %w", err)
	}
	go s.renderer.Render(ctx, event)
	go s.historyInsert(ctx, handler.History()...)
	go s.notify(ctx, handler.Notifications()...)

	return nil
}

// historyInsert inserts a history item.
func (s *EventService) historyInsert(ctx context.Context, items ...*models.HistoryItem) {
	for _, item := range items {
		if err := s.store.HistoryInsert(ctx, item); err != nil {
			s.log.Error("[event service] failed to insert history item: "+err.Error(), trace.Attr(ctx))
		}
	}
}

// notify sends notifications.
func (s *EventService) notify(ctx context.Context, items ...*models.Notification) {
	for _, item := range items {
		s.notifier.Notify(ctx, item)
	}
}

// validateEvent validates the event.
func (s *EventService) validateEvent(e *models.Event) error {
	errs := errMap{}
	if len(e.ID) != s.cfg.EventIDLen {
		errs["id"] = fmt.Errorf("event ID must be %d characters long", s.cfg.EventIDLen)
	}
	if len(e.Caption) > s.cfg.EventTextMaxLen {
		errs["text"] = fmt.Errorf("event text must be at most %d characters long", s.cfg.EventTextMaxLen)
	}
	errs["owner"] = s.validateProfile(&e.Owner)
	return errs.Filter()
}

func (s *EventService) validateDancer(d *models.Dancer) error {
	errs := errMap{}
	if d.Profile == nil {
		errs["profile"] = fmt.Errorf("dancer profile must be provided")
	} else {
		errs["profile"] = s.validateProfile(d.Profile)
	}
	errs["full_name"] = s.validateFullname(d.FullName)
	errs["role"] = s.validateRole(d.Role)
	return errs.Filter()
}

func (s *EventService) validateProfile(p *models.Profile) error {
	errs := errMap{}
	if p.ID < 1 {
		errs["id"] = fmt.Errorf("profile ID must be positive")
	}
	if p.FirstName == "" {
		errs["first_name"] = fmt.Errorf("profile first name must be provided")
	}
	return errs.Filter()
}

func (s *EventService) validateRole(r models.Role) error {
	if r != models.RoleLeader && r != models.RoleFollower {
		return fmt.Errorf("role must be either %q or %q", models.RoleLeader, models.RoleFollower)
	}
	return nil
}

func (s *EventService) validateFullname(fn string) error {
	if len(fn) < 1 || len(fn) > s.cfg.DancerNameMaxLen {
		return fmt.Errorf("full name must be between 1 and %d characters long", s.cfg.DancerNameMaxLen)
	}
	return nil
}

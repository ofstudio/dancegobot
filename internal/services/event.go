package services

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/ofstudio/dancegobot/helpers"
	"github.com/ofstudio/dancegobot/helpers/trace"
	"github.com/ofstudio/dancegobot/internal/config"
	"github.com/ofstudio/dancegobot/internal/models"
)

var nowFn = time.Now

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
		log:      helpers.NopLogger(),
	}
}

func (s *EventService) WithLogger(l *slog.Logger) *EventService {
	s.log = l
	return s
}

// NewID generates a new event ID
func (s *EventService) NewID() string {
	return string(helpers.RandToken(s.cfg.EventIDLen))
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
		Profile:   &event.Owner,
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

func (s *EventService) DancerGet(event *models.Event, profile *models.Profile, role models.Role) *models.Dancer {
	return newEventHandler(event).DancerGetByProfile(profile, role)
}

func (s *EventService) CoupleAdd(
	ctx context.Context,
	eventID string,
	profile *models.Profile,
	role models.Role,
	other any,
) (*models.EventUpdate, error) {

	if err := s.validateProfile(profile); err != nil {
		return nil, fmt.Errorf("failed to validate profile: %w", err)
	}

	var (
		otherProfile *models.Profile
		otherName    string
	)
	switch v := other.(type) {
	case *models.Profile:
		if err := s.validateProfile(v); err != nil {
			return nil, fmt.Errorf("failed to validate other person profile: %w", err)
		}
		otherProfile = v
	case string:
		if err := s.validateFullname(v); err != nil {
			return nil, fmt.Errorf("failed to validate other person name: %w", err)
		}
		otherName = v
	default:
		return nil, fmt.Errorf("invalid type of other person: %T", other)
	}

	return s.updateWrapper(ctx, eventID, func(h *eventHandler) *models.EventUpdate {
		dancer := h.DancerGetByProfile(profile, role)
		var partner *models.Dancer
		if otherProfile != nil {
			partner = h.DancerGetByProfile(otherProfile, role.Opposite())
		} else {
			partner = h.DancerGetByName(otherName, role.Opposite())
		}

		return h.CoupleAdd(dancer, partner)
	})
}

// SingleAdd adds a single dancer to the event.
func (s *EventService) SingleAdd(
	ctx context.Context,
	eventID string,
	profile *models.Profile,
	role models.Role,
) (*models.EventUpdate, error) {
	if err := s.validateProfile(profile); err != nil {
		return nil, fmt.Errorf("failed to validate profile: %w", err)
	}
	return s.updateWrapper(ctx, eventID, func(h *eventHandler) *models.EventUpdate {
		dancer := h.DancerGetByProfile(profile, role)
		return h.SingleAdd(dancer)
	})
}

// DancerRemove removes  the dancer (and couple if any) from the event.
func (s *EventService) DancerRemove(
	ctx context.Context,
	eventID string,
	profile *models.Profile,
) (*models.EventUpdate, error) {
	if err := s.validateProfile(profile); err != nil {
		return nil, fmt.Errorf("failed to validate profile: %w", err)
	}
	return s.updateWrapper(ctx, eventID, func(h *eventHandler) *models.EventUpdate {
		dancer := h.DancerGetByProfile(profile, models.RoleLeader)
		return h.DancerRemove(dancer)
	})
}

// updateWrapper is a wrapper for the event updates
func (s *EventService) updateWrapper(
	ctx context.Context,
	eventID string,
	updateFn func(*eventHandler) *models.EventUpdate,
) (*models.EventUpdate, error) {
	// Begin tx
	tx, err := s.store.BeginTx(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin tx: %w", err)
	}
	//goland:noinspection ALL
	defer tx.Rollback()

	// Get event
	event, err := tx.EventGet(ctx, eventID)
	if err != nil {
		return nil, fmt.Errorf("failed to get event: %w", err)
	}

	// Create event handler
	handler := newEventHandler(event)

	// Run update function
	upd := updateFn(handler)
	if upd.Result != models.ResultSuccess {
		return upd, nil
	}

	// If the update is successful,
	// - update the event
	// - commit the transaction
	// - render event announcement
	// - add history items
	// - send the notifications
	if err = tx.EventUpsert(ctx, upd.Event); err != nil {
		return nil, fmt.Errorf("failed to upsert event: %w", err)
	}
	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit tx: %w", err)
	}
	go s.renderer.Render(ctx, event)
	go s.historyInsert(ctx, handler.History()...)
	go s.notify(ctx, handler.Notifications()...)

	return upd, nil
}

// historyInsert inserts a history item.
func (s *EventService) historyInsert(ctx context.Context, items ...*models.HistoryItem) {
	for _, item := range items {
		if err := s.store.HistoryInsert(ctx, item); err != nil {
			s.log.Error("[event service] failed to insert history item: "+err.Error(),
				"item", item, trace.Attr(ctx))
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
	errs := helpers.Errors{}
	if len(e.ID) != s.cfg.EventIDLen {
		errs["id"] = fmt.Errorf("event ID must be %d characters long", s.cfg.EventIDLen)
	}
	if len(e.Caption) > s.cfg.EventTextMaxLen {
		errs["text"] = fmt.Errorf("event text must be at most %d characters long", s.cfg.EventTextMaxLen)
	}

	if e.MessageID == "" {
		errs["message_id"] = fmt.Errorf("event message ID must be provided")
	}

	errs["owner"] = s.validateProfile(&e.Owner)

	return errs.Filter()
}

func (s *EventService) validateDancer(d *models.Dancer) error {
	errs := helpers.Errors{}
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
	errs := helpers.Errors{}
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

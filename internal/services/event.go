package services

import (
	"context"
	"fmt"
	"log/slog"
	"time"
	"unicode/utf8"

	"github.com/ofstudio/dancegobot/internal/config"
	"github.com/ofstudio/dancegobot/internal/models"
	"github.com/ofstudio/dancegobot/internal/store"
	"github.com/ofstudio/dancegobot/pkg/errutil"
	"github.com/ofstudio/dancegobot/pkg/noplog"
	"github.com/ofstudio/dancegobot/pkg/randtoken"
	"github.com/ofstudio/dancegobot/pkg/trace"
)

// EventService is a service that manages dance events
type EventService struct {
	cfg      config.Settings
	store    store.Store
	notifier *NotifierService
	renderer *RenderService
	log      *slog.Logger
}

func NewEventService(cfg config.Settings, store store.Store, r *RenderService, n *NotifierService) *EventService {
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

// Start starts draft cleanup scheduler.
func (s *EventService) Start(ctx context.Context) {
	go s.draftsCleanupScheduler(ctx)
}

// Create creates a new event.
func (s *EventService) Create(
	ctx context.Context,
	caption string,
	owner models.Profile,
	settings models.EventSettings,
) (*models.Event, error) {

	event := &models.Event{
		ID:        randtoken.New(s.cfg.EventIDLen),
		Caption:   caption,
		Settings:  settings,
		Owner:     owner,
		CreatedAt: nowFn(),
	}
	if err := s.validateEvent(event); err != nil {
		return nil, fmt.Errorf("failed to validate event: %w", err)
	}

	if err := s.store.EventUpsert(ctx, event); err != nil {
		return nil, fmt.Errorf("failed to upsert event: %w", err)
	}

	go s.historyItemsCreate(ctx, &models.HistoryItem{
		Action:    models.HistoryEventCreated,
		Initiator: &event.Owner,
		EventID:   &event.ID,
		Details:   event,
		CreatedAt: nowFn(),
	})

	return event, nil
}

// Get returns an event by ID.
func (s *EventService) Get(ctx context.Context, id string) (*models.Event, error) {
	event, err := s.store.EventGet(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get event: %w", err)
	}
	return event, nil
}

// GetMy returns slice of event id related to the specified profile:
//   - non-draft events owned by the user
//   - events where the user is a participant: either in a couple or as a single
func (s *EventService) GetMy(ctx context.Context, profile *models.Profile) ([]string, error) {
	if profile == nil {
		return nil, fmt.Errorf("profile is nil")
	}
	ids, err := s.store.EventGetMy(ctx, profile)
	if err != nil {
		return nil, fmt.Errorf("failed to get my events: %w", err)
	}
	return ids, nil
}

// CanManage returns true if the profile can manage the event.
func (s *EventService) CanManage(event *models.Event, profile *models.Profile) bool {
	if event == nil || profile == nil {
		return false
	}
	return NewEventHandler(event).CanManage(profile)
}

// UpdateSettings updates event settings.
func (s *EventService) UpdateSettings(
	ctx context.Context,
	eventID string,
	initiator *models.Profile,
	settings models.EventSettings,
) (*models.Event, error) {
	if initiator == nil {
		return nil, fmt.Errorf("initiator is nil")
	}
	var event *models.Event
	err := s.update(ctx, eventID, func(h *EventHandler) error {
		event = h.Event()
		return h.UpdateSettings(initiator, settings)
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update event settings: %w", err)
	}
	return event, nil
}

// RegistrationGet returns registration for the given event by profile and role.
// If the dancer is not registered, returns a new registration.
// If event or profile is nil, returns nil.
func (s *EventService) RegistrationGet(event *models.Event, profile *models.Profile, role models.Role) *models.Registration {
	if event == nil || profile == nil {
		return nil
	}
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
		return nil, nil, fmt.Errorf("inline message is empty")
	}

	var event *models.Event
	var post *models.Post
	err := s.update(ctx, eventID, func(h *EventHandler) error {
		if h.Event().Post == nil {
			h.Event().Post = &models.Post{}
		}
		h.Event().Post.InlineMessageID = inlineMessageID
		event = h.Event()
		post = h.Event().Post
		h.hist = append(h.hist, &models.HistoryItem{
			Action:    models.HistoryPostAdded,
			Initiator: &h.event.Owner,
			EventID:   &h.event.ID,
			Details:   h.event.Post,
		})
		return nil
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
	err := s.update(ctx, eventID, func(h *EventHandler) error {
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
		return nil
	})
	return event, post, err
}

// LimitChangeAffected returns list of affected couples after the limit change.
// Returns true as second argument if the limit was increased, otherwise false.
// Returns the start index of the affected couples as the third argument.
func (s *EventService) LimitChangeAffected(event *models.Event, oldLimit int) ([]models.Couple, bool, int) {
	if event == nil {
		return nil, false, 0
	}
	return NewEventHandler(event).LimitChangeAffected(oldLimit)
}

// LimitChangeNotify notifies the affected dancers about the event limit change.
func (s *EventService) LimitChangeNotify(ctx context.Context, eventID string, oldLimit int) {
	err := s.update(ctx, eventID, func(h *EventHandler) error {
		h.LimitChangeNotify(oldLimit)
		return nil
	})
	if err != nil {
		s.log.Error("[event service] failed to notify about event limit change: "+err.Error(),
			"event_id", eventID, trace.Attr(ctx))
	}
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
	if profile == nil {
		return nil, fmt.Errorf("profile is nil")
	}
	dancer := &models.Dancer{
		Profile:   profile,
		FullName:  profile.FullName(),
		Role:      role,
		CreatedAt: nowFn(),
	}
	if err := s.validateDancer(dancer); err != nil {
		return nil, fmt.Errorf("failed to validate dancer: %w", err)
	}

	var partner *models.Dancer
	switch v := other.(type) {
	case *models.Profile:
		partner = &models.Dancer{
			Profile:   v,
			FullName:  v.FullName(),
			Role:      role.Opposite(),
			CreatedAt: nowFn(),
		}
	case string:
		partner = &models.Dancer{
			FullName:  v,
			Role:      role.Opposite(),
			CreatedAt: nowFn(),
		}
	default:
		return nil, fmt.Errorf("invalid type of other person: %T", other)
	}
	if err := s.validateDancer(partner); err != nil {
		return nil, fmt.Errorf("failed to validate partner: %w", err)
	}

	var reg *models.Registration
	err := s.update(ctx, eventID, func(h *EventHandler) error {
		reg = h.CoupleAdd(dancer, partner)
		return nil
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

	if profile == nil {
		return nil, fmt.Errorf("profile is nil")
	}
	dancer := &models.Dancer{
		Profile:   profile,
		FullName:  profile.FullName(),
		Role:      role,
		CreatedAt: nowFn(),
	}
	if err := s.validateDancer(dancer); err != nil {
		return nil, fmt.Errorf("failed to validate dancer: %w", err)
	}

	var reg *models.Registration
	err := s.update(ctx, eventID, func(h *EventHandler) error {
		reg = h.SingleAdd(dancer)
		return nil
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
	if profile == nil {
		return nil, fmt.Errorf("profile is nil")
	}
	var reg *models.Registration
	err := s.update(ctx, eventID, func(h *EventHandler) error {
		reg = h.DancerRemove(&models.Dancer{
			Profile:  profile,
			FullName: profile.FullName(),
		})
		return nil
	})
	return reg, err
}

// update is a wrapper for the event handler.
func (s *EventService) update(
	ctx context.Context,
	eventID string,
	updateFunc func(*EventHandler) error,
) error {
	// Begin tx
	tx, err := s.store.Begin(ctx)
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
	if err = updateFunc(handler); err != nil {
		return fmt.Errorf("failed to update event: %w", err)
	}

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
	go s.historyItemsCreate(ctx, handler.History()...)
	go s.notificationsSend(ctx, handler.Notifications()...)

	return nil
}

// historyItemsCreate creates history items in the store.
func (s *EventService) historyItemsCreate(ctx context.Context, items ...*models.HistoryItem) {
	for _, item := range items {
		if err := s.store.HistoryCreate(ctx, item); err != nil {
			s.log.Error("[event service] failed to insert history item: "+err.Error(), trace.Attr(ctx))
		}
	}
}

// notificationsSend sends notifications.
func (s *EventService) notificationsSend(ctx context.Context, items ...*models.Notification) {
	for _, item := range items {
		s.notifier.Notify(ctx, item)
	}
}

func (s *EventService) draftsCleanupScheduler(ctx context.Context) {
	if s.cfg.DraftCleanupEvery == 0 {
		s.log.Info("[event service] drafts cleanup is disabled")
		return
	}

	s.log.Info("[event service] starting drafts cleanup scheduler",
		slog.Duration("interval", s.cfg.DraftCleanupEvery),
		slog.Duration("older_than", s.cfg.DraftCleanupOlderThan))

	ticker := time.NewTicker(s.cfg.DraftCleanupEvery)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			s.log.Info("[event service] drafts cleanup scheduler stopped")
			return
		case <-ticker.C:
			s.draftsCleanup(trace.Context(ctx, "drafts_cleanup_"+randtoken.New(4)))
		}
	}
}

func (s *EventService) draftsCleanup(ctx context.Context) {
	before := time.Now().Add(-s.cfg.DraftCleanupOlderThan)
	ids, err := s.store.EventRemoveDraftsBefore(ctx, before)
	if err != nil {
		s.log.Error("[event service] failed to remove draft events: "+err.Error(), trace.Attr(ctx))
		return
	}
	s.log.Info("[event service] removed draft events",
		slog.Int("count", len(ids)),
		trace.Attr(ctx),
	)
	count, err := s.store.HistoryRemoveByEventIDs(ctx, ids)
	if err != nil {
		s.log.Error("[event service] failed to remove draft events history items: "+err.Error(), trace.Attr(ctx))
		return
	}
	s.log.Info("[event service] removed draft events history items",
		slog.Int("count", count),
		trace.Attr(ctx),
	)
}

// validateEvent validates the event.
func (s *EventService) validateEvent(e *models.Event) error {
	var err error
	if utf8.RuneCountInString(e.Caption) > s.cfg.EventTextMaxLen {
		err = errutil.Append(err, fmt.Errorf("event text must be at most %d characters long, got %d",
			s.cfg.EventTextMaxLen, utf8.RuneCountInString(e.Caption)))
	}
	err = errutil.Append(err, s.validateProfile(&e.Owner))
	return err
}

func (s *EventService) validateDancer(d *models.Dancer) error {
	if d == nil {
		return fmt.Errorf("dancer is nil")
	}
	var err error
	if d.Profile != nil {
		err = errutil.Append(err, s.validateProfile(d.Profile))
	}
	err = errutil.Append(err, s.validateFullname(d.FullName))
	err = errutil.Append(err, s.validateRole(d.Role))
	return err
}

func (s *EventService) validateProfile(p *models.Profile) error {
	if p == nil {
		return fmt.Errorf("profile is nil")
	}
	var err error
	if p.ID < 1 {
		err = errutil.Append(err, fmt.Errorf("profile ID must be positive, got %d", p.ID))
	}
	if p.FirstName == "" {
		err = errutil.Append(err, fmt.Errorf("profile first name must be provided"))
	}
	return err
}

func (s *EventService) validateRole(r models.Role) error {
	if r != models.RoleLeader && r != models.RoleFollower {
		return fmt.Errorf("invalid role: %q", r)
	}
	return nil
}

func (s *EventService) validateFullname(name string) error {
	n := utf8.RuneCountInString(name)
	if n < 1 || n > s.cfg.DancerNameMaxLen {
		return fmt.Errorf("full name must be between 1 and %d characters long, got %d",
			s.cfg.DancerNameMaxLen, n)
	}
	return nil
}

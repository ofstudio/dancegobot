package services

import (
	"context"
	"log/slog"
	"strconv"
	"sync"
	"sync/atomic"

	"github.com/ofstudio/dancegobot/internal/config"
	"github.com/ofstudio/dancegobot/internal/models"
	"github.com/ofstudio/dancegobot/internal/store"
	"github.com/ofstudio/dancegobot/pkg/noplog"
	"github.com/ofstudio/dancegobot/pkg/repeater"
	"github.com/ofstudio/dancegobot/pkg/trace"
)

// NotifyFunc is an executor that sends Telegram notification.
type NotifyFunc func(*models.Notification) error

// NotifierService is a service that sends notifications to users.
type NotifierService struct {
	cfg        config.Settings
	store      store.Store
	notifyFunc NotifyFunc
	repeater   *repeater.Repeater
	sequence   atomic.Uint64
	log        *slog.Logger
}

func NewNotifierService(cfg config.Settings, store store.Store, notifyFunc NotifyFunc) *NotifierService {
	return &NotifierService{
		cfg:        cfg,
		store:      store,
		notifyFunc: notifyFunc,
		repeater:   repeater.NewRepeater(cfg.NotifierRepeats),
		log:        noplog.Logger(),
	}
}

func (s *NotifierService) WithLogger(l *slog.Logger) *NotifierService {
	s.log = l
	return s
}

// Notify sends a notification immediately and schedules in-memory retries on failure.
// History is recorded only after success or exhaustion of all attempts.
func (s *NotifierService) Notify(ctx context.Context, n *models.Notification) {
	if ctx.Err() != nil {
		return
	}

	// Retain a private snapshot: callers may change their data after Notify returns.
	notification := *n
	notification.Recipient = cloneProfilePtr(n.Recipient)
	notification.Payload = cloneNotificationPayload(n.Payload)
	id := strconv.FormatUint(s.sequence.Add(1), 10)
	err := s.send(ctx, &notification, id, 1)
	n.Error = notification.Error
	if err != nil {
		if len(s.cfg.NotifierRepeats) == 0 {
			s.historyCreate(ctx, &notification)
			return
		}
		s.repeat(ctx, &notification, id)
		return
	}
	s.historyCreate(ctx, &notification)
}

// send performs one delivery attempt. All errors are eligible for retries;
// the configured schedule is the only backoff policy, including for Telegram 429.
func (s *NotifierService) send(ctx context.Context, n *models.Notification, id string, attempt int) error {
	n.Error = ""
	err := s.notifyFunc(n)
	if err != nil {
		n.Error = err.Error()
		s.log.Error("[notifier service] failed to send notification: "+err.Error(),
			"notification", n, "notification_id", id, "attempt", attempt, trace.Attr(ctx))
		return err
	}
	s.log.Info("[notifier service] notification sent",
		"notification", n, "notification_id", id, "attempt", attempt, trace.Attr(ctx))
	return nil
}

func (s *NotifierService) repeat(ctx context.Context, n *models.Notification, id string) {
	repeatCtx, cancel := context.WithCancel(ctx)
	var mu sync.Mutex
	attempt := 1
	s.repeater.AddTask(repeatCtx, id, func(taskCtx context.Context, _ string) {
		// Repeater callbacks may overlap; serialize one delivery and recheck cancellation.
		mu.Lock()
		defer mu.Unlock()
		if taskCtx.Err() != nil {
			return
		}
		attempt++
		err := s.send(taskCtx, n, id, attempt)
		if err != nil {
			if attempt < 1+len(s.cfg.NotifierRepeats) {
				return
			}
			cancel()
			s.log.Error("[notifier service] notification delivery retries exhausted",
				"notification", n, "notification_id", id, "attempt", attempt, trace.Attr(ctx))
			s.historyCreate(ctx, n)
			return
		}
		cancel()
		// Keep the caller context for history; the repeat context was just canceled.
		s.historyCreate(ctx, n)
	})
}

func (s *NotifierService) historyCreate(ctx context.Context, n *models.Notification) {
	var eventID *string
	if n.Payload.Event != nil {
		eventID = &n.Payload.Event.ID
	}
	h := &models.HistoryItem{
		Action:    models.HistoryNotificationSent,
		Initiator: config.BotProfile(),
		EventID:   eventID,
		Details:   n,
		CreatedAt: nowFn(),
	}
	if err := s.store.HistoryCreate(ctx, h); err != nil {
		s.log.Error("[notifier service] failed to insert history item: "+err.Error(), trace.Attr(ctx))
	}
}

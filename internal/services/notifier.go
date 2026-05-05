package services

import (
	"context"
	"log/slog"

	"github.com/ofstudio/dancegobot/internal/config"
	"github.com/ofstudio/dancegobot/internal/models"
	"github.com/ofstudio/dancegobot/internal/store"
	"github.com/ofstudio/dancegobot/pkg/noplog"
	"github.com/ofstudio/dancegobot/pkg/trace"
)

// NotifyFunc is an executor that sends Telegram notification.
type NotifyFunc func(*models.Notification) error

// NotifierService is a service that sends notifications to users.
type NotifierService struct {
	cfg        config.Settings
	store      store.Store
	notifyFunc NotifyFunc
	log        *slog.Logger
}

func NewNotifierService(cfg config.Settings, store store.Store, notifyFunc NotifyFunc) *NotifierService {
	return &NotifierService{
		cfg:        cfg,
		store:      store,
		notifyFunc: notifyFunc,
		log:        noplog.Logger(),
	}
}

func (s *NotifierService) WithLogger(l *slog.Logger) *NotifierService {
	s.log = l
	return s
}

// Notify sends a notification to the user.
func (s *NotifierService) Notify(ctx context.Context, n *models.Notification) {
	if err := s.notifyFunc(n); err != nil {
		s.log.Error("[notifier service] failed to send notification: "+err.Error(), trace.Attr(ctx))
		n.Error = err.Error()
	} else {
		s.log.Info("[notifier service] notification sent", "", n, trace.Attr(ctx))
	}

	// Insert history item
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

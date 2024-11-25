package services

import (
	"context"
	"log/slog"

	"github.com/ofstudio/dancegobot/internal/config"
	"github.com/ofstudio/dancegobot/internal/models"
	"github.com/ofstudio/dancegobot/pkg/noplog"
	"github.com/ofstudio/dancegobot/pkg/trace"
)

type NotifyFunc func(*models.Notification) error

// NotifierService is a service that sends notifications to users.
type NotifierService struct {
	cfg   config.Settings
	store Store
	do    NotifyFunc
	log   *slog.Logger
}

func NewNotifierService(cfg config.Settings, store Store, f NotifyFunc) *NotifierService {
	return &NotifierService{
		cfg:   cfg,
		store: store,
		do:    f,
		log:   noplog.Logger(),
	}
}

func (s *NotifierService) WithLogger(l *slog.Logger) *NotifierService {
	s.log = l
	return s
}

// Notify sends a notification to the user.
func (s *NotifierService) Notify(ctx context.Context, n *models.Notification) {
	if err := s.do(n); err != nil {
		s.log.Error("[notification service] failed to send notification: "+err.Error(), trace.Attr(ctx))
		n.Error = err.Error()
	} else {
		s.log.Info("[notification service] notification sent", "", n, trace.Attr(ctx))
	}

	h := &models.HistoryItem{
		Action:    models.HistoryNotificationSent,
		Initiator: n.Initiator,
		EventID:   n.EventID,
		Details:   n,
		CreatedAt: nowFn(),
	}

	if err := s.store.HistoryInsert(ctx, h); err != nil {
		s.log.Error("[notification service] failed to insert history item: "+err.Error(), trace.Attr(ctx))

	}
}

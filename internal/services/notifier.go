package services

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	tele "gopkg.in/telebot.v4"

	"github.com/ofstudio/dancegobot/internal/config"
	"github.com/ofstudio/dancegobot/internal/locale"
	"github.com/ofstudio/dancegobot/internal/models"
	"github.com/ofstudio/dancegobot/internal/telegram/views"
	"github.com/ofstudio/dancegobot/pkg/noplog"
	"github.com/ofstudio/dancegobot/pkg/trace"
)

// NotifierService is a service that sends notifications to users.
type NotifierService struct {
	cfg   config.Settings
	store Store
	api   tele.API
	log   *slog.Logger
}

func NewNotifierService(cfg config.Settings, store Store, api tele.API) *NotifierService {
	return &NotifierService{
		cfg:   cfg,
		store: store,
		api:   api,
		log:   noplog.Logger(),
	}
}

func (s *NotifierService) WithLogger(l *slog.Logger) *NotifierService {
	s.log = l
	return s
}

// Notify sends a notification to the user.
func (s *NotifierService) Notify(ctx context.Context, n *models.Notification) {
	if err := s.send(n); err != nil {
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

func (s *NotifierService) send(n *models.Notification) error {
	if n.Event != nil {
		n.EventID = &n.Event.ID
	}

	t, ok := locale.Notifications[n.TmplCode]
	if !ok {
		return fmt.Errorf("unknown notification template: %s", n.TmplCode)
	}
	user := &tele.User{ID: n.Recipient.ID}

	var initiatorName string
	if n.Initiator != nil {
		initiatorName = views.FmtName(n.Initiator.FullName(), n.Initiator)
	}
	text := fmt.Sprintf(t, n.Event.Caption, initiatorName)
	_, err := s.api.Send(user, text, tele.ModeHTML, tele.NoPreview, tele.RemoveKeyboard)

	if err != nil && !errors.Is(err, tele.ErrTrueResult) {
		return err
	}
	return nil
}

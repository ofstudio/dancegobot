package services

import (
	"context"
	"errors"
	"log/slog"

	tele "gopkg.in/telebot.v4"

	"github.com/ofstudio/dancegobot/internal/config"
	"github.com/ofstudio/dancegobot/internal/models"
	"github.com/ofstudio/dancegobot/internal/telegram/views"
	"github.com/ofstudio/dancegobot/pkg/noplog"
	"github.com/ofstudio/dancegobot/pkg/repeater"
	"github.com/ofstudio/dancegobot/pkg/trace"
)

// RenderService renders events announcements.
type RenderService struct {
	cfg      config.Settings
	store    Store
	repeater *repeater.Repeater
	bot      tele.API
	log      *slog.Logger
}

func NewRenderService(cfg config.Settings, store Store, bot tele.API) *RenderService {
	return &RenderService{
		cfg:      cfg,
		store:    store,
		repeater: repeater.NewRepeater(cfg.RendererRepeats),
		bot:      bot,
		log:      noplog.Logger(),
	}
}

func (s *RenderService) WithLogger(l *slog.Logger) *RenderService {
	s.log = l
	return s
}

// Render renders the event announcement and schedules a render repeat.
func (s *RenderService) Render(ctx context.Context, event *models.Event) {
	if err := s.render(event); err != nil {
		s.log.Error("[render service] failed to render event: "+err.Error(), trace.Attr(ctx))
	}
	s.repeater.AddTask(ctx, event.ID, s.renderRepeat)
}

func (s *RenderService) renderRepeat(ctx context.Context, eventID string) {
	event, err := s.store.EventGet(ctx, eventID)
	if err != nil {
		s.log.Error("[render service] failed to get event: "+err.Error(), trace.Attr(ctx))
		return
	}
	if err = s.render(event); err != nil {
		s.log.Error("[render service] failed to render event: "+err.Error(), trace.Attr(ctx))
		return
	}
}

func (s *RenderService) render(event *models.Event) error {
	err := views.Render(s.bot, s.cfg.BotName, event)
	if err != nil && !errors.Is(err, tele.ErrTrueResult) {
		return err
	}
	return nil
}

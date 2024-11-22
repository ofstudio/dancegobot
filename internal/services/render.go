package services

import (
	"context"
	"errors"
	"log/slog"

	tele "gopkg.in/telebot.v4"

	"github.com/ofstudio/dancegobot/helpers"
	"github.com/ofstudio/dancegobot/helpers/repeater"
	"github.com/ofstudio/dancegobot/helpers/trace"
	"github.com/ofstudio/dancegobot/internal/config"
	"github.com/ofstudio/dancegobot/internal/models"
	"github.com/ofstudio/dancegobot/internal/telegram/views"
)

// RenderService renders events announcements.
type RenderService struct {
	cfg      config.Settings
	store    Store
	repeater *repeater.Repeater
	bot      tele.API
	botName  string
	log      *slog.Logger
}

func NewRenderService(cfg config.Settings, store Store, bot tele.API, botName string) *RenderService {
	return &RenderService{
		cfg:      cfg,
		store:    store,
		repeater: repeater.NewRepeater(cfg.RendererRepeats),
		bot:      bot,
		botName:  botName,
		log:      helpers.NopLogger(),
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
	err := views.Render(s.bot, s.botName, event)
	if err != nil && !errors.Is(err, tele.ErrTrueResult) {
		return err
	}
	return nil
}
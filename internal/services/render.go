package services

import (
	"context"
	"log/slog"

	"github.com/ofstudio/dancegobot/internal/config"
	"github.com/ofstudio/dancegobot/internal/models"
	"github.com/ofstudio/dancegobot/pkg/noplog"
	"github.com/ofstudio/dancegobot/pkg/repeater"
	"github.com/ofstudio/dancegobot/pkg/trace"
)

type RenderFunc func(event *models.Event) error

// RenderService renders events announcements.
type RenderService struct {
	cfg      config.Settings
	store    Store
	do       RenderFunc
	repeater *repeater.Repeater
	log      *slog.Logger
}

func NewRenderService(cfg config.Settings, store Store, f RenderFunc) *RenderService {
	return &RenderService{
		cfg:      cfg,
		store:    store,
		do:       f,
		repeater: repeater.NewRepeater(cfg.RendererRepeats),
		log:      noplog.Logger(),
	}
}

func (s *RenderService) WithLogger(l *slog.Logger) *RenderService {
	s.log = l
	return s
}

// Render renders the event announcement and schedules a render repeat.
func (s *RenderService) Render(ctx context.Context, event *models.Event) {
	s.render(ctx, event)
	s.repeater.AddTask(ctx, event.ID, s.renderRepeat)
}

func (s *RenderService) renderRepeat(ctx context.Context, eventID string) {
	event, err := s.store.EventGet(ctx, eventID)
	if err != nil {
		s.log.Error("[render service] failed to get event: "+err.Error(), trace.Attr(ctx))
		return
	}
	s.render(ctx, event)
}

func (s *RenderService) render(ctx context.Context, event *models.Event) {
	if err := s.do(event); err != nil {
		s.log.Error("[render service] failed to render event: "+err.Error(), trace.Attr(ctx))
	}
}

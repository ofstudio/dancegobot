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

type RenderFunc func(event *models.Event, inlineMessageID string) error

// RenderService renders events posts.
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

// Render renders the event post and schedules a render repeat.
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
	if event.Post == nil {
		s.log.Error("[render service] failed to render event: post is nil",
			"event_id", event.ID,
			trace.Attr(ctx))
		return
	}
	if event.Post.InlineMessageID == "" {
		s.log.Warn("[render service] skipping render: inline message ID is empty",
			"event_id", event.ID,
			trace.Attr(ctx))
		return
	}
	if err := s.do(event, event.Post.InlineMessageID); err != nil {
		s.log.Error(
			"[render service] failed to render event: "+err.Error(),
			"event_id", event.ID,
			trace.Attr(ctx))
	}
}

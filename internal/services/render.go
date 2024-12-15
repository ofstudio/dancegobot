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
	store      Store
	renderFunc RenderFunc
	queue      chan *models.Event
	repeater   *repeater.Repeater
	log        *slog.Logger
}

func NewRenderService(cfg config.Settings, store Store, renderFunc RenderFunc) *RenderService {
	return &RenderService{
		store:      store,
		renderFunc: renderFunc,
		queue:      make(chan *models.Event),
		repeater:   repeater.NewRepeater(cfg.RendererRepeats),
		log:        noplog.Logger(),
	}
}

func (s *RenderService) WithLogger(l *slog.Logger) *RenderService {
	s.log = l
	return s
}

// Start starts the render queue.
func (s *RenderService) Start(ctx context.Context) {
	go s.queueHandler(ctx)
}

// Render renders the event post and schedules a render repeat.
func (s *RenderService) Render(ctx context.Context, event *models.Event) {
	if event == nil {
		s.log.Error("[render service] event is nil", trace.Attr(ctx))
		return
	}
	s.render(ctx, event)
	s.repeater.AddTask(ctx, event.ID, s.renderRepeat)
}

// renderRepeat retrieves the event from the store and renders it again.
func (s *RenderService) renderRepeat(ctx context.Context, eventID string) {
	event, err := s.store.EventGet(ctx, eventID)
	if err != nil {
		s.log.Error("[render service] failed to get event: "+err.Error(), trace.Attr(ctx))
		return
	}
	s.render(ctx, event)
}

// render adds the event to the render queue.
func (s *RenderService) render(ctx context.Context, event *models.Event) {
	switch {
	case event == nil:
		s.log.Error("[render service] failed to render event: event is nil", trace.Attr(ctx))
		return
	case event.Post == nil || event.Post.InlineMessageID == "":
		s.log.Warn("[render service] skipping render: inline message ID is not set",
			"event", event.LogValue(),
			trace.Attr(ctx))
		return
	default:
		// Add event to the rendering queue.
		// This will wait for the previous rendering to complete
		s.queue <- event
	}
}

// queueHandler reads events from the queue and renders them.
// The queue is necessary in case of high load, to avoid a situation
// where earlier rendering requests are processed by Telegram later than later ones.
// See: https://github.com/ofstudio/dancegobot/issues/6
func (s *RenderService) queueHandler(ctx context.Context) {
	s.log.Info("[render service] render queue started")
	for {
		select {
		case event := <-s.queue:
			if err := s.renderFunc(event, event.Post.InlineMessageID); err != nil {
				s.log.Error(
					"[render service] failed to render event: "+err.Error(),
					"event", event.LogValue(),
					trace.Attr(ctx))
			}
			s.render(ctx, event)
		case <-ctx.Done():
			s.log.Info("[render service] render queue stopped")
			return
		}
	}
}

package services

import (
	"context"
	"log/slog"
	"strings"
	"time"

	"github.com/ofstudio/dancegobot/internal/config"
	"github.com/ofstudio/dancegobot/internal/models"
	"github.com/ofstudio/dancegobot/internal/store"
	"github.com/ofstudio/dancegobot/pkg/noplog"
	"github.com/ofstudio/dancegobot/pkg/repeater"
	"github.com/ofstudio/dancegobot/pkg/trace"
)

type RenderFunc func(event *models.Event, inlineMessageID string) error

// RenderService renders events posts.
type RenderService struct {
	cfg        config.Settings
	store      store.Store
	renderFunc RenderFunc
	queue      chan queueItem
	repeater   *repeater.Repeater
	log        *slog.Logger
}

func NewRenderService(cfg config.Settings, store store.Store, renderFunc RenderFunc) *RenderService {
	return &RenderService{
		cfg:        cfg,
		store:      store,
		renderFunc: renderFunc,
		queue:      make(chan queueItem),
		repeater:   repeater.NewRepeater(cfg.RendererRepeats),
		log:        noplog.Logger(),
	}
}

func (s *RenderService) WithLogger(l *slog.Logger) *RenderService {
	s.log = l
	return s
}

// Start starts the render queue and re-renders recent events at startup.
func (s *RenderService) Start(ctx context.Context) {
	go s.queueHandler(trace.Context(ctx, "render_queue_handler"))
	go s.renderAtStartup(trace.Context(ctx, "render_at_startup"))
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
		s.log.Info("[render service] skipping render: inline message ID is not set",
			"event", event.LogValue(),
			trace.Attr(ctx))
		return
	case event.Removed:
		s.log.Info("[render service] skipping render: event marked as removed",
			"event", event.LogValue(), trace.Attr(ctx))
		return
	default:
		// Add event to the rendering queue.
		// This will wait for the previous rendering to complete
		s.queue <- queueItem{event: event, ctx: ctx}
	}
}

// renderAtStartup re-renders recent events on startup.
func (s *RenderService) renderAtStartup(ctx context.Context) {
	if s.cfg.ReRenderOnStartup == 0 {
		s.log.Info("[render service] re-rendering at startup is disabled", trace.Attr(ctx))
		return
	}

	events, err := s.store.EventGetUpdatedAfter(ctx, time.Now().Add(-s.cfg.ReRenderOnStartup))
	if err != nil {
		s.log.Error("[render service] failed to get events to re-render: "+err.Error(), trace.Attr(ctx))
		return
	}
	s.log.Info("[render service] re-rendering recent events at startup",
		slog.Duration("updated_within", s.cfg.ReRenderOnStartup),
		slog.Int("count", len(events)),
		trace.Attr(ctx))

	for _, event := range events {
		s.render(ctx, event)
	}
	s.log.Info("[render service] re-rendering at startup completed", trace.Attr(ctx))
}

// queueHandler reads events from the queue and renders them.
// The queue is necessary in case of high load, to avoid a situation
// where earlier rendering requests are processed by Telegram later than later ones.
// See: https://github.com/ofstudio/dancegobot/issues/6
func (s *RenderService) queueHandler(ctx context.Context) {
	s.log.Info("[render service] render queue started", trace.Attr(ctx))
	for {
		select {
		case item := <-s.queue:
			if err := s.renderFunc(item.event, item.event.Post.InlineMessageID); err != nil {
				if s.errIsPostRemoved(err) {
					s.log.Error("[render service] event post was probably removed: "+err.Error(),
						"event", item.event.LogValue(), trace.Attr(item.ctx))
					s.eventRenderFail(item.ctx, item.event.ID)
				} else {
					s.log.Error("[render service] failed to render event: "+err.Error(),
						"event", item.event.LogValue(), trace.Attr(item.ctx))
				}
			}
		case <-ctx.Done():
			s.log.Info("[render service] render queue stopped", trace.Attr(ctx))
			return
		}
	}
}

// eventRenderFail increases the rendering failure counter for the event.
// If the fail counter exceeds the limit, the event is marked as removed.
func (s *RenderService) eventRenderFail(ctx context.Context, eventID string) {
	// Start transaction.
	tx, err := s.store.Begin(ctx)
	if err != nil {
		s.log.Error("[render service] failed to start transaction: "+err.Error(), trace.Attr(ctx))
		return
	}
	//goland:noinspection ALL
	defer tx.Rollback()

	// Get event.
	event, err := tx.EventGet(ctx, eventID)
	if err != nil {
		s.log.Error("[render service] failed to get event: "+err.Error(), trace.Attr(ctx))
		return
	}

	// Update rendering failure count.
	event.RenderFails++

	// If the number of rendering failures exceeds the limit, mark the event as removed.
	if event.RenderFails >= s.cfg.RenderFailsMax {
		event.Removed = true
		// Create history item.
		if err = tx.HistoryCreate(ctx, &models.HistoryItem{
			Action:    models.HistoryEventRemoved,
			Initiator: config.BotProfile(),
			EventID:   &event.ID,
			Details:   event,
		}); err != nil {
			s.log.Error("[render service] failed to create history item: "+err.Error(), trace.Attr(ctx))
		}
		s.log.Info("[render service] event marked as removed due to rendering failures",
			"event", event.LogValue(), trace.Attr(ctx))
	}

	// Update event.
	if err = tx.EventUpsert(ctx, event); err != nil {
		s.log.Error("[render service] failed to update event: "+err.Error(), trace.Attr(ctx))
		return
	}

	// Commit transaction.
	if err = tx.Commit(); err != nil {
		s.log.Error("[render service] failed to commit transaction: "+err.Error(), trace.Attr(ctx))
		return
	}
}

// errIsPostRemoved returns true if the error is due to event post was probably removed.
func (s *RenderService) errIsPostRemoved(err error) bool {
	return strings.Contains(err.Error(), "Bad Request: MESSAGE_ID_INVALID")
}

// queueItem - rendering queue item.
type queueItem struct {
	event *models.Event
	ctx   context.Context
}

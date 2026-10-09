package services

import (
	"context"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/ofstudio/dancegobot/internal/config"
	"github.com/ofstudio/dancegobot/internal/models"
	"github.com/ofstudio/dancegobot/internal/store"
	"github.com/ofstudio/dancegobot/pkg/noplog"
	"github.com/ofstudio/dancegobot/pkg/repeater"
	"github.com/ofstudio/dancegobot/pkg/trace"
)

// RenderFunc is an executor that renders an event post in Telegram.
type RenderFunc func(event *models.Event, inlineMessageID string) error

// RenderService renders events posts.
type RenderService struct {
	cfg        config.Settings
	store      store.Store
	renderFunc RenderFunc
	queue      chan queueItem
	repeater   *repeater.Repeater
	ctx        context.Context
	log        *slog.Logger

	// Independent initial delay may reduce missing mobile posts: https://github.com/ofstudio/dancegobot/issues/21
	initialRepeater *repeater.Repeater
	initialMu       sync.Mutex
	initialDone     map[string]bool
}

func NewRenderService(cfg config.Settings, store store.Store, renderFunc RenderFunc) *RenderService {
	return &RenderService{
		cfg:        cfg,
		store:      store,
		renderFunc: renderFunc,
		queue:      make(chan queueItem),
		repeater:   repeater.NewRepeater(cfg.RendererRepeats),
		// One-shot delay may reduce missing mobile posts: https://github.com/ofstudio/dancegobot/issues/21
		initialRepeater: repeater.NewRepeater([]time.Duration{cfg.RendererInitialDelay}),
		initialDone:     make(map[string]bool),
		log:             noplog.Logger(),
	}
}

func (s *RenderService) WithLogger(l *slog.Logger) *RenderService {
	s.log = l
	return s
}

// Start starts the render queue and re-renders recent events at startup.
func (s *RenderService) Start(ctx context.Context) {
	s.ctx = ctx
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
	if event == nil {
		s.log.Error("[render service] event not found", trace.Attr(ctx))
		return
	}
	s.render(ctx, event)
}

// renderable checks whether the event has an editable post.
func (s *RenderService) renderable(ctx context.Context, event *models.Event) bool {
	switch {
	case event == nil:
		s.log.Error("[render service] failed to render event: event is nil", trace.Attr(ctx))
		return false
	case event.Post == nil || event.Post.InlineMessageID == "":
		s.log.Info("[render service] skipping render: inline message ID is not set",
			"event", event.LogValue(),
			trace.Attr(ctx))
		return false
	case event.Removed:
		s.log.Info("[render service] skipping render: event marked as removed",
			"event", event.LogValue(), trace.Attr(ctx))
		return false
	}
	return true
}

// render adds the event to the render queue.
func (s *RenderService) render(ctx context.Context, event *models.Event) {
	if !s.renderable(ctx, event) {
		return
	}
	// Gate every render path to potentially reduce missing posts: https://github.com/ofstudio/dancegobot/issues/21
	if s.delayInitialRender(ctx, event.ID) {
		return
	}
	s.enqueue(queueItem{event: event, ctx: ctx})
}

// delayInitialRender gives mobile clients time to receive the original inline post
// before its first edit, which may reduce missing posts. This is only a hypothesis:
// https://github.com/ofstudio/dancegobot/issues/21
// State is process-local: a restart also delays the first edit of existing posts.
func (s *RenderService) delayInitialRender(ctx context.Context, eventID string) bool {
	if s.cfg.RendererInitialDelay <= 0 {
		return false
	}
	s.initialMu.Lock()
	defer s.initialMu.Unlock()
	if done, exists := s.initialDone[eventID]; exists {
		return !done
	}
	s.initialDone[eventID] = false
	s.log.Info("[render service] initial render scheduled",
		"event_id", eventID, "delay", s.cfg.RendererInitialDelay, trace.Attr(ctx))
	s.initialRepeater.AddTask(s.ctx, eventID, func(_ context.Context, id string) {
		// Repeater cancels its callback context on completion. The queued read
		// must instead use the service lifetime, independently of render repeats.
		s.enqueue(queueItem{eventID: id, ctx: trace.Context(s.ctx, "render_initial")})
	})
	return true
}

// enqueue waits for the previous render to finish without blocking shutdown.
func (s *RenderService) enqueue(item queueItem) {
	if s.ctx.Err() != nil {
		return
	}
	if item.ctx.Err() != nil {
		return
	}
	select {
	case s.queue <- item:
	case <-s.ctx.Done():
	case <-item.ctx.Done():
	}
}

// initialEvent completes the delay before reading the latest state. Changes made
// during this read must queue normally so they cannot be lost at the boundary.
// The delay may reduce missing mobile posts: https://github.com/ofstudio/dancegobot/issues/21
func (s *RenderService) initialEvent(ctx context.Context, eventID string) *models.Event {
	s.initialMu.Lock()
	s.initialDone[eventID] = true
	s.initialMu.Unlock()
	s.log.Info("[render service] initial render executing",
		"event_id", eventID, "delay", s.cfg.RendererInitialDelay, trace.Attr(ctx))
	event, err := s.store.EventGet(ctx, eventID)
	if err != nil {
		s.log.Error("[render service] failed to get event: "+err.Error(),
			"event_id", eventID, trace.Attr(ctx))
		return nil
	}
	return event
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
			if ctx.Err() != nil {
				return
			}
			if item.eventID != "" {
				// Finish the experimental delay for missing posts: https://github.com/ofstudio/dancegobot/issues/21
				item.event = s.initialEvent(item.ctx, item.eventID)
				if !s.renderable(item.ctx, item.event) {
					continue
				}
			}
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
	if event == nil {
		s.log.Error("[render service] event not found", trace.Attr(ctx))
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
	event   *models.Event
	eventID string // Deferred initial renders carry only the ID, never a stale snapshot.
	ctx     context.Context
}

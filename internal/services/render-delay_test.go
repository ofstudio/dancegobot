package services

import (
	"context"
	"errors"
	"sync"
	"testing"
	"testing/synctest"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/ofstudio/dancegobot/internal/config"
	"github.com/ofstudio/dancegobot/internal/models"
	"github.com/ofstudio/dancegobot/internal/store"
)

func TestRenderInitialDelay(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		cfg := initialRenderConfig()
		cfg.RendererRepeats = []time.Duration{3 * time.Second}
		st := &initialRenderStore{}
		s, rendered, cancel := startInitialRenderTest(cfg, st)
		defer cancel()
		a, b := testRenderEvent("a"), testRenderEvent("b")
		st.set(a, b)
		for range 10 {
			go s.Render(s.ctx, a)
		}
		synctest.Wait()
		require.Empty(t, rendered)

		time.Sleep(time.Second)
		s.Render(s.ctx, b)
		synctest.Wait()
		time.Sleep(3 * time.Second)
		synctest.Wait()
		require.Empty(t, rendered, "ordinary repeats must also respect the delay")

		latest := *a
		latest.Caption = "latest state"
		st.set(&latest, b)
		for range 10 {
			go s.Render(s.ctx, a)
		}
		synctest.Wait()
		time.Sleep(time.Second)
		synctest.Wait()
		require.Len(t, rendered, 1, "pending requests must not move the deadline")
		require.Equal(t, &latest, <-rendered)

		time.Sleep(time.Second)
		synctest.Wait()
		require.Len(t, rendered, 1, "each event has an independent deadline")
		require.Equal(t, b, <-rendered)

		s.Render(s.ctx, &latest)
		synctest.Wait()
		require.Len(t, rendered, 1, "later requests must render immediately")
		require.Equal(t, &latest, <-rendered)
	})
}

func TestRenderInitialDelayDisabled(t *testing.T) {
	for _, delay := range []time.Duration{0, -time.Second} {
		t.Run(delay.String(), func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				cfg := initialRenderConfig()
				cfg.RendererInitialDelay = delay
				s, rendered, cancel := startInitialRenderTest(cfg, &initialRenderStore{})
				defer cancel()
				event := testRenderEvent("immediate")
				s.Render(context.Background(), event)
				synctest.Wait()
				require.Len(t, rendered, 1)
				require.Equal(t, event, <-rendered)
			})
		})
	}
}

func TestRenderInitialDelayStartsOnlyForEditablePosts(t *testing.T) {
	for _, name := range []string{"missing", "draft", "chat only", "removed"} {
		t.Run(name, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				st := &initialRenderStore{}
				s, rendered, cancel := startInitialRenderTest(initialRenderConfig(), st)
				defer cancel()
				event := testRenderEvent("eligible")
				switch name {
				case "missing":
					event = nil
				case "draft":
					event.Post = nil
				case "chat only":
					event.Post.InlineMessageID = ""
				case "removed":
					event.Removed = true
				}
				s.Render(context.Background(), event)
				synctest.Wait()
				time.Sleep(10 * time.Second)
				synctest.Wait()
				require.Empty(t, rendered)

				event = testRenderEvent("eligible")
				st.set(event)
				s.Render(context.Background(), event)
				synctest.Wait()
				time.Sleep(4 * time.Second)
				synctest.Wait()
				require.Empty(t, rendered)
				time.Sleep(time.Second)
				synctest.Wait()
				require.Len(t, rendered, 1)
			})
		})
	}
}

func TestRenderInitialDelayRechecksStoredEvent(t *testing.T) {
	for _, name := range []string{"missing", "draft", "chat only", "removed", "store error"} {
		t.Run(name, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				st := &initialRenderStore{}
				s, rendered, cancel := startInitialRenderTest(initialRenderConfig(), st)
				defer cancel()
				event := testRenderEvent("changed")
				s.Render(context.Background(), event)
				synctest.Wait()
				latest := testRenderEvent(event.ID)
				switch name {
				case "missing":
					latest = nil
				case "draft":
					latest.Post = nil
				case "chat only":
					latest.Post.InlineMessageID = ""
				case "removed":
					latest.Removed = true
				case "store error":
					st.err = errors.New("store unavailable")
				}
				st.set(latest)
				time.Sleep(5 * time.Second)
				synctest.Wait()
				require.Empty(t, rendered)

				st.err = nil
				st.set(event)
				s.renderRepeat(context.Background(), event.ID)
				synctest.Wait()
				require.Len(t, rendered, 1, "a failed initial read must not start another delay")
			})
		})
	}
}

func TestRenderInitialDelaySurvivesRepeatCancellation(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		cfg := initialRenderConfig()
		cfg.RendererRepeats = []time.Duration{3 * time.Second, 10 * time.Second}
		st := &initialRenderStore{}
		s, rendered, cancel := startInitialRenderTest(cfg, st)
		defer cancel()
		event := testRenderEvent("repeat")
		st.set(event)
		ctx, cancelRepeat := context.WithCancel(context.Background())
		s.Render(ctx, event)
		synctest.Wait()
		time.Sleep(time.Second)
		cancelRepeat()
		s.Render(s.ctx, event)
		synctest.Wait()
		time.Sleep(3 * time.Second)
		synctest.Wait()
		require.Empty(t, rendered)
		time.Sleep(time.Second)
		synctest.Wait()
		require.Len(t, rendered, 1)
		<-rendered
		time.Sleep(6 * time.Second)
		synctest.Wait()
		require.Len(t, rendered, 1, "ordinary repeats continue after the initial render")
	})
}

func TestRenderInitialDelayConcurrentUpdateDuringRead(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		reading, release := make(chan struct{}), make(chan struct{})
		st := &initialRenderStore{afterRead: func() {
			close(reading)
			<-release
		}}
		s, rendered, cancel := startInitialRenderTest(initialRenderConfig(), st)
		defer cancel()
		event := testRenderEvent("boundary")
		st.set(event)
		s.Render(context.Background(), event)
		synctest.Wait()
		time.Sleep(5 * time.Second)
		<-reading
		synctest.Wait() // The one-shot repeater has returned and canceled its context.

		latest := *event
		latest.Caption = "updated during the read"
		go s.Render(context.Background(), &latest)
		synctest.Wait()
		require.Empty(t, rendered)
		close(release)
		synctest.Wait()
		require.Len(t, rendered, 2)
		require.Equal(t, event, <-rendered, "the queued read must have a live context")
		require.Equal(t, &latest, <-rendered, "the concurrent update must not be lost")
	})
}

func TestRenderInitialDelayStartupAndRestart(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		cfg := initialRenderConfig()
		cfg.ReRenderOnStartup = time.Hour
		st := &initialRenderStore{}
		st.set(testRenderEvent("old post"))
		for range 2 {
			_, rendered, cancel := startInitialRenderTest(cfg, st)
			synctest.Wait()
			require.Empty(t, rendered)
			time.Sleep(4 * time.Second)
			synctest.Wait()
			require.Empty(t, rendered)
			time.Sleep(time.Second)
			synctest.Wait()
			require.Len(t, rendered, 1)
			cancel()
			synctest.Wait()
		}
	})
}

func TestRenderInitialDelayShutdown(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		s, rendered, cancel := startInitialRenderTest(initialRenderConfig(), &initialRenderStore{})
		s.Render(context.Background(), testRenderEvent("pending"))
		synctest.Wait()
		cancel()
		time.Sleep(10 * time.Second)
		synctest.Wait()
		require.Empty(t, rendered)
	})
}

func TestRenderInitialDelayShutdownWithBusyQueue(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		st := &initialRenderStore{}
		a, b := testRenderEvent("a"), testRenderEvent("b")
		st.set(a, b)
		release := make(chan struct{})
		calls := 0
		s := NewRenderService(initialRenderConfig(), st, func(*models.Event, string) error {
			calls++
			<-release
			return nil
		})
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		s.Start(ctx)
		s.Render(ctx, a)
		synctest.Wait()
		time.Sleep(5 * time.Second)
		synctest.Wait()
		require.Equal(t, 1, calls)
		s.Render(ctx, b)
		go s.Render(context.Background(), a)
		synctest.Wait()
		time.Sleep(5 * time.Second)
		synctest.Wait() // Both a normal request and the deferred request wait for the queue.
		cancel()
		synctest.Wait()
		close(release)
		synctest.Wait()
		require.Equal(t, 1, calls)
	})
}

func initialRenderConfig() config.Settings {
	cfg := config.Default().Settings
	cfg.RendererRepeats = nil
	cfg.ReRenderOnStartup = 0
	return cfg
}

func startInitialRenderTest(cfg config.Settings, st store.Store) (*RenderService, chan *models.Event, context.CancelFunc) {
	rendered := make(chan *models.Event, 32)
	s := NewRenderService(cfg, st, func(event *models.Event, _ string) error {
		rendered <- event
		return nil
	})
	ctx, cancel := context.WithCancel(context.Background())
	s.Start(ctx)
	return s, rendered, cancel
}

type initialRenderStore struct {
	store.Store
	mu        sync.Mutex
	events    []*models.Event
	err       error
	afterRead func()
}

func (s *initialRenderStore) set(events ...*models.Event) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.events = events
}

func (s *initialRenderStore) EventGet(ctx context.Context, id string) (*models.Event, error) {
	s.mu.Lock()
	var event *models.Event
	for _, candidate := range s.events {
		if candidate != nil && candidate.ID == id {
			event = candidate
		}
	}
	err := s.err
	s.mu.Unlock()
	if err != nil {
		return nil, err
	}
	if s.afterRead != nil {
		s.afterRead()
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return event, nil
}

func (s *initialRenderStore) EventGetUpdatedAfter(context.Context, time.Time) ([]*models.Event, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.events, nil
}

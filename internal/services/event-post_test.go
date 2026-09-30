package services

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/ofstudio/dancegobot/internal/config"
	"github.com/ofstudio/dancegobot/internal/models"
	"github.com/ofstudio/dancegobot/internal/store"
)

func TestEventServicePostAddRejectsReplacement(t *testing.T) {
	for _, withChat := range []bool{false, true} {
		t.Run(map[bool]string{false: "without chat", true: "with chat"}[withChat], func(t *testing.T) {
			event := &models.Event{ID: "event", Post: &models.Post{InlineMessageID: "original"}, SubscribersNotified: true}
			if withChat {
				event.Post.Chat = &models.Chat{ID: -1001, Type: models.ChatSuper, Title: "Dance"}
				event.Post.ChatMessageID = 10
			}
			before := cloneEvent(event)
			st := &postAddConflictStore{event: event}
			s := NewEventService(config.Default().Settings, st, nil, nil)
			require.NotPanics(t, func() {
				updated, err := s.PostAdd(context.Background(), event.ID, "replacement")
				require.ErrorContains(t, err, "event post inline message ID is already set")
				require.Nil(t, updated)
			})
			require.Equal(t, before, event)
			require.Equal(t, 1, st.begins)
			require.Equal(t, 1, st.reads)
			require.Equal(t, 1, st.rollbacks)
		})
	}
}

func TestEventServicePostAddRejectsEmptyID(t *testing.T) {
	st := &postAddConflictStore{}
	s := NewEventService(config.Default().Settings, st, nil, nil)
	event, err := s.PostAdd(context.Background(), "event", "")
	require.ErrorContains(t, err, "inline message is empty")
	require.Nil(t, event)
	require.Zero(t, st.begins)
}

func TestEventServicePostAddIsIdempotentAndImmutable(t *testing.T) {
	for _, tc := range []struct {
		name string
		post *models.Post
	}{
		{"new post", nil},
		{"empty post", &models.Post{}},
		{"chat already known", &models.Post{
			Chat: &models.Chat{ID: -1001, Type: models.ChatSuper, Title: "Dance"}, ChatMessageID: 10,
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s, st, rendered := newEventPostServiceTest(t)
			ctx := context.Background()
			event := &models.Event{
				ID: "event", Owner: models.Profile{ID: 1, FirstName: "Owner"},
				Post: tc.post, SubscribersNotified: true,
			}
			require.NoError(t, st.EventUpsert(ctx, event))
			want := cloneEvent(event)
			if want.Post == nil {
				want.Post = &models.Post{}
			}
			want.Post.InlineMessageID = "original"

			updated, err := s.PostAdd(ctx, event.ID, "original")
			require.NoError(t, err)
			require.Equal(t, want, updated)
			waitPostRender(t, rendered, "original")
			require.Eventually(t, func() bool {
				return historyActionCount(t, st, models.HistoryPostAdded) == 1
			}, time.Second, 10*time.Millisecond)

			updated, err = s.PostAdd(ctx, event.ID, "original")
			require.NoError(t, err)
			require.Equal(t, want, updated)
			waitPostRender(t, rendered, "original")

			updated, err = s.PostAdd(ctx, event.ID, "replacement")
			require.ErrorContains(t, err, "event post inline message ID is already set")
			require.Nil(t, updated)
			stored, err := st.EventGet(ctx, event.ID)
			require.NoError(t, err)
			require.Equal(t, want, stored)
			require.Equal(t, 1, historyActionCount(t, st, models.HistoryPostAdded))
			require.Zero(t, historyActionCount(t, st, models.HistoryNotificationSent))
			select {
			case id := <-rendered:
				t.Fatalf("unexpected render: %s", id)
			default:
			}
		})
	}
}

func TestEventServicePostAddConcurrentIDs(t *testing.T) {
	s, st, rendered := newEventPostServiceTest(t)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)
	event := &models.Event{
		ID: "event", Owner: models.Profile{ID: 1}, SubscribersNotified: true,
		Post: &models.Post{Chat: &models.Chat{ID: -1001, Type: models.ChatSuper}, ChatMessageID: 10},
	}
	require.NoError(t, st.EventUpsert(ctx, event))
	arrivals := make(chan struct{}, 2)
	release := make(chan struct{})
	var once sync.Once
	unblock := func() { once.Do(func() { close(release) }) }
	defer unblock()
	s.store = &postAddBarrierStore{Store: st, arrivals: arrivals, release: release}
	type result struct {
		id    string
		event *models.Event
		err   error
	}
	done := make(chan result, 2)
	for _, id := range []string{"first", "second"} {
		go func() {
			updated, err := s.PostAdd(ctx, event.ID, id)
			done <- result{id: id, event: updated, err: err}
		}()
	}
	for range 2 {
		select {
		case <-arrivals:
		case <-ctx.Done():
			t.Fatal("post operations did not reach Begin")
		}
	}
	unblock()
	var winner string
	failures := 0
	for range 2 {
		select {
		case r := <-done:
			if r.err != nil {
				require.ErrorContains(t, r.err, "event post inline message ID is already set")
				require.Nil(t, r.event)
				failures++
				continue
			}
			require.Empty(t, winner)
			winner = r.id
			require.Equal(t, r.id, r.event.Post.InlineMessageID)
		case <-ctx.Done():
			t.Fatal("post operations did not finish")
		}
	}
	require.NotEmpty(t, winner)
	require.Equal(t, 1, failures)
	waitPostRender(t, rendered, winner)
	stored, err := st.EventGet(ctx, event.ID)
	require.NoError(t, err)
	want := cloneEvent(event)
	want.Post.InlineMessageID = winner
	require.Equal(t, want, stored)
	require.Eventually(t, func() bool {
		return historyActionCount(t, st, models.HistoryPostAdded) == 1
	}, time.Second, 10*time.Millisecond)
}

func newEventPostServiceTest(t *testing.T) (*EventService, *store.SQLiteStore, <-chan string) {
	t.Helper()
	s, st := newEventServiceTest(t)
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	rendered := make(chan string, 4)
	s.renderer = NewRenderService(s.cfg, st, func(_ *models.Event, id string) error {
		rendered <- id
		return nil
	})
	s.renderer.Start(ctx)
	return s, st, rendered
}

func waitPostRender(t *testing.T, rendered <-chan string, want string) {
	t.Helper()
	select {
	case id := <-rendered:
		require.Equal(t, want, id)
	case <-time.After(time.Second):
		t.Fatal("post was not rendered")
	}
}

// Unimplemented methods fail if a rejected publication attempts to save or commit.
type postAddConflictStore struct {
	store.Store
	event     *models.Event
	begins    int
	reads     int
	rollbacks int
}

func (s *postAddConflictStore) Begin(context.Context) (store.Store, error) {
	s.begins++
	return s, nil
}

func (s *postAddConflictStore) EventGet(context.Context, string) (*models.Event, error) {
	s.reads++
	return s.event, nil
}

func (s *postAddConflictStore) Rollback() error {
	s.rollbacks++
	return nil
}

type postAddBarrierStore struct {
	store.Store
	arrivals chan<- struct{}
	release  <-chan struct{}
}

func (s *postAddBarrierStore) Begin(ctx context.Context) (store.Store, error) {
	s.arrivals <- struct{}{}
	select {
	case <-s.release:
		return s.Store.Begin(ctx)
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

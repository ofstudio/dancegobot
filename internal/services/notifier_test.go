package services

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	tele "gopkg.in/telebot.v4"

	"github.com/ofstudio/dancegobot/internal/config"
	"github.com/ofstudio/dancegobot/internal/models"
	storepkg "github.com/ofstudio/dancegobot/internal/store"
)

func TestNotifierServiceNotify(t *testing.T) {
	ctx := context.Background()

	t.Run("stores successful notification history", func(t *testing.T) {
		service, st := newNotifierServiceTest(t, nil)
		notification := testNotification()

		service.Notify(ctx, notification)

		require.Empty(t, notification.Error)
		assertNotificationHistory(t, st, "notify_event", "")
	})

	t.Run("stores failed notification history with error", func(t *testing.T) {
		service, st := newNotifierServiceTest(t, errors.New("telegram blocked bot"))
		notification := testNotification()

		service.Notify(ctx, notification)

		require.Equal(t, "telegram blocked bot", notification.Error)
		assertNotificationHistory(t, st, "notify_event", "telegram blocked bot")
	})
}

func TestNotifierServiceRetries(t *testing.T) {
	tests := []struct {
		name string
		err  error
	}{
		{name: "network timeout", err: context.DeadlineExceeded},
		{name: "unknown error", err: errors.New("unknown delivery failure")},
		{name: "rate limit", err: tele.NewError(429, "Too Many Requests")},
		{name: "server failure", err: tele.ErrInternal},
		{name: "invalid request", err: tele.NewError(400, "Bad Request")},
		{name: "blocked bot", err: tele.ErrBlockedByUser},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var calls atomic.Int64
			var staleError atomic.Bool
			service, st, ctx := newNotifierRetryTest(t, []time.Duration{
				5 * time.Millisecond, 10 * time.Millisecond, 15 * time.Millisecond,
			}, func(n *models.Notification) error {
				if n.Error != "" {
					staleError.Store(true)
				}
				if calls.Add(1) <= 2 {
					return tt.err
				}
				return nil
			})
			notification := testNotification()
			service.Notify(ctx, notification)
			require.Equal(t, tt.err.Error(), notification.Error)

			waitNotificationHistory(t, st, 1)
			assertNotificationHistory(t, st, "notify_event", "")
			require.Equal(t, int64(3), calls.Load())
			require.False(t, staleError.Load())
			require.Never(t, func() bool { return calls.Load() > 3 }, 25*time.Millisecond, time.Millisecond)
			// Retries must not write back to the caller's notification after Notify returns.
			require.Equal(t, tt.err.Error(), notification.Error)
		})
	}
}

func TestNotifierServiceImmediateSuccessDoesNotRetry(t *testing.T) {
	var calls atomic.Int64
	service, st, ctx := newNotifierRetryTest(t, []time.Duration{5 * time.Millisecond}, func(*models.Notification) error {
		calls.Add(1)
		return nil
	})
	notification := testNotification()
	notification.Error = "previous failure"
	service.Notify(ctx, notification)
	require.Empty(t, notification.Error)
	assertNotificationHistory(t, st, "notify_event", "")
	require.Never(t, func() bool { return calls.Load() > 1 }, 20*time.Millisecond, time.Millisecond)
}

func TestNotifierServiceExhaustsRetries(t *testing.T) {
	var calls atomic.Int64
	var active atomic.Int64
	var overlap atomic.Bool
	service, st, ctx := newNotifierRetryTest(t, []time.Duration{
		time.Millisecond, time.Millisecond, time.Millisecond, time.Millisecond,
		time.Millisecond, time.Millisecond, time.Millisecond, time.Millisecond,
	}, func(*models.Notification) error {
		if active.Add(1) > 1 {
			overlap.Store(true)
		}
		defer active.Add(-1)
		call := calls.Add(1)
		// Make scheduled callbacks overlap so serialization is exercised.
		time.Sleep(5 * time.Millisecond)
		return fmt.Errorf("delivery failure %d", call)
	})
	service.Notify(ctx, testNotification())
	waitNotificationHistory(t, st, 1)
	assertNotificationHistory(t, st, "notify_event", "delivery failure 9")
	require.Equal(t, int64(9), calls.Load())
	require.False(t, overlap.Load())
	require.Never(t, func() bool { return calls.Load() > 9 }, 20*time.Millisecond, time.Millisecond)
}

func TestNotifierServicePendingHistoryAndSnapshot(t *testing.T) {
	release := make(chan struct{})
	var once sync.Once
	t.Cleanup(func() { once.Do(func() { close(release) }) })
	entered := make(chan struct{})
	got := make(chan models.Notification, 1)
	var calls atomic.Int64
	service, st, ctx := newNotifierRetryTest(t, []time.Duration{time.Millisecond}, func(n *models.Notification) error {
		if calls.Add(1) == 1 {
			return errors.New("offline")
		}
		close(entered)
		<-release
		got <- *n
		return nil
	})
	notification := testNotification()
	notification.Payload.Event.Post = &models.Post{Chat: &models.Chat{Title: "Original chat"}}
	notification.Payload.Event.Couples = []models.Couple{{Dancers: []models.Dancer{*notification.Payload.Partner}}}
	service.Notify(ctx, notification)
	notification.Recipient.FirstName = "Changed recipient"
	notification.Payload.Event.ID = "changed_event"
	notification.Payload.Event.Caption = "Changed caption"
	notification.Payload.Event.Post.Chat.Title = "Changed chat"
	notification.Payload.Event.Couples[0].Dancers[0].Profile.FirstName = "Changed partner"
	notification.Payload.Partner.FullName = "Changed partner name"
	notification.Error = "Changed caller error"

	select {
	case <-entered:
	case <-time.After(time.Second):
		t.Fatal("retry did not start")
	}
	var count int
	require.NoError(t, st.DB().Get(&count, "SELECT COUNT(*) FROM history"))
	require.Zero(t, count)
	once.Do(func() { close(release) })
	select {
	case received := <-got:
		require.Equal(t, "Recipient", received.Recipient.FirstName)
		require.Equal(t, "notify_event", received.Payload.Event.ID)
		require.Equal(t, "Notify event", received.Payload.Event.Caption)
		require.Equal(t, "Original chat", received.Payload.Event.Post.Chat.Title)
		require.Equal(t, "Partner", received.Payload.Event.Couples[0].Dancers[0].Profile.FirstName)
		require.Equal(t, "Partner", received.Payload.Partner.FullName)
		require.Empty(t, received.Error)
	case <-time.After(time.Second):
		t.Fatal("retry did not finish")
	}
	waitNotificationHistory(t, st, 1)
	assertNotificationHistory(t, st, "notify_event", "")
	require.Equal(t, "Changed caller error", notification.Error)
}

func TestNotifierServiceOverlappingRetriesStopAfterSuccess(t *testing.T) {
	var calls atomic.Int64
	service, st, ctx := newNotifierRetryTest(t, []time.Duration{
		time.Millisecond, time.Millisecond, time.Millisecond,
	}, func(*models.Notification) error {
		if calls.Add(1) == 1 {
			return errors.New("offline")
		}
		time.Sleep(10 * time.Millisecond)
		return nil
	})
	service.Notify(ctx, testNotification())
	waitNotificationHistory(t, st, 1)
	require.Never(t, func() bool { return calls.Load() > 2 }, 25*time.Millisecond, time.Millisecond)
	assertNotificationHistory(t, st, "notify_event", "")
}

func TestNotifierServiceIdenticalNotificationsAreIndependent(t *testing.T) {
	var mu sync.Mutex
	counts := make(map[*models.Notification]int)
	service, st, ctx := newNotifierRetryTest(t, []time.Duration{5 * time.Millisecond}, func(n *models.Notification) error {
		mu.Lock()
		defer mu.Unlock()
		counts[n]++
		if counts[n] == 1 {
			return errors.New("offline")
		}
		return nil
	})
	var wg sync.WaitGroup
	for range 2 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			service.Notify(ctx, testNotification())
		}()
	}
	wg.Wait()
	waitNotificationHistory(t, st, 2)
	mu.Lock()
	defer mu.Unlock()
	require.Len(t, counts, 2)
	for _, count := range counts {
		require.Equal(t, 2, count)
	}
}

func TestNotifierServiceCancellationAndRestart(t *testing.T) {
	entered := make(chan struct{})
	release := make(chan struct{})
	var once sync.Once
	t.Cleanup(func() { once.Do(func() { close(release) }) })
	var calls atomic.Int64
	service, st, parent := newNotifierRetryTest(t, []time.Duration{
		time.Millisecond, time.Millisecond, time.Millisecond,
	}, func(*models.Notification) error {
		if calls.Add(1) == 2 {
			close(entered)
			<-release
		}
		return errors.New("offline")
	})
	ctx, cancel := context.WithCancel(parent)
	defer cancel()
	service.Notify(ctx, testNotification())
	select {
	case <-entered:
	case <-time.After(time.Second):
		t.Fatal("retry did not start")
	}
	cancel()
	once.Do(func() { close(release) })

	var restoredCalls atomic.Int64
	NewNotifierService(service.cfg, st, func(*models.Notification) error {
		restoredCalls.Add(1)
		return nil
	})
	require.Never(t, func() bool {
		return calls.Load() > 2 || restoredCalls.Load() > 0
	}, 30*time.Millisecond, time.Millisecond)
	service.Notify(ctx, testNotification())
	require.Equal(t, int64(2), calls.Load())
	var count int
	require.NoError(t, st.DB().Get(&count, "SELECT COUNT(*) FROM history"))
	require.Zero(t, count)
}

func TestNotifierServiceHistoryFailureDoesNotRetryDelivery(t *testing.T) {
	var calls atomic.Int64
	service, st, ctx := newNotifierRetryTest(t, []time.Duration{
		5 * time.Millisecond, 10 * time.Millisecond, 15 * time.Millisecond,
	}, func(*models.Notification) error {
		if calls.Add(1) == 1 {
			return errors.New("offline")
		}
		return nil
	})
	broken := &notifierHistoryFailureStore{Store: st}
	service.store = broken
	service.Notify(ctx, testNotification())
	require.Eventually(t, func() bool { return broken.calls.Load() == 1 }, time.Second, time.Millisecond)
	require.Never(t, func() bool {
		return calls.Load() > 2 || broken.calls.Load() > 1
	}, 25*time.Millisecond, time.Millisecond)
}

type notifierHistoryFailureStore struct {
	storepkg.Store
	calls atomic.Int64
}

func (s *notifierHistoryFailureStore) HistoryCreate(context.Context, *models.HistoryItem) error {
	s.calls.Add(1)
	return errors.New("history unavailable")
}

func newNotifierServiceTest(t *testing.T, notifyErr error) (*NotifierService, *storepkg.SQLiteStore) {
	t.Helper()
	service, st, _ := newNotifierRetryTest(t, nil, func(*models.Notification) error { return notifyErr })
	return service, st
}

func newNotifierRetryTest(t *testing.T, repeats []time.Duration, notify NotifyFunc) (*NotifierService, *storepkg.SQLiteStore, context.Context) {
	t.Helper()

	cfg := config.Default()
	cfg.Settings.NotifierRepeats = repeats
	db, err := storepkg.NewSQLite(":memory:", cfg.DB.Version)
	require.NoError(t, err)

	st := storepkg.NewSQLiteStore(db)
	t.Cleanup(st.Close)

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	service := NewNotifierService(cfg.Settings, st, notify)
	return service, st, ctx
}

func waitNotificationHistory(t *testing.T, st *storepkg.SQLiteStore, want int) {
	t.Helper()
	require.Eventually(t, func() bool {
		var count int
		err := st.DB().Get(&count, "SELECT COUNT(*) FROM history WHERE action = ?", models.HistoryNotificationSent)
		if err != nil {
			return false
		}
		return count == want
	}, time.Second, time.Millisecond)
}

func testNotification() *models.Notification {
	eventID := "notify_event"
	return &models.Notification{
		TmplCode:  models.TmplRegisteredWithSingle,
		Recipient: &models.Profile{ID: 1, FirstName: "Recipient"},
		Payload: models.NotificationPayload{
			Event: &models.Event{
				ID:      eventID,
				Caption: "Notify event",
				Owner:   models.Profile{ID: 2, FirstName: "Owner"},
			},
			Partner: &models.Dancer{
				Profile:  &models.Profile{ID: 3, FirstName: "Partner"},
				FullName: "Partner",
				Role:     models.RoleLeader,
			},
		},
	}
}

func assertNotificationHistory(t *testing.T, st *storepkg.SQLiteStore, eventID, errText string) {
	t.Helper()

	var count int
	require.NoError(t, st.DB().GetContext(context.Background(), &count,
		`SELECT COUNT(*) FROM history WHERE action = ? AND event_id = ?`,
		models.HistoryNotificationSent, eventID))
	require.Equal(t, 1, count)

	var data string
	require.NoError(t, st.DB().GetContext(context.Background(), &data,
		`SELECT data FROM history WHERE action = ? AND event_id = ?`,
		models.HistoryNotificationSent, eventID))
	if errText == "" {
		require.NotContains(t, data, `"error"`)
	} else {
		require.Contains(t, data, errText)
	}
}

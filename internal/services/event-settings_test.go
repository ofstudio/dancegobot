package services

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/ofstudio/dancegobot/internal/models"
	"github.com/ofstudio/dancegobot/internal/store"
)

func TestEventServiceConcurrentSettingsFields(t *testing.T) {
	s, st := newEventServiceTest(t)
	event := &models.Event{ID: "event", Owner: models.Profile{ID: 1}, Settings: models.EventSettings{Limit: 2}}
	require.NoError(t, st.EventUpsert(context.Background(), event))

	results := runConcurrentSettings(t, s, []settingsOperation{
		func(ctx context.Context) settingsResult {
			e, err := s.SettingsAutoPairingToggle(ctx, event.ID, event.Owner)
			return settingsResult{event: e, err: err}
		},
		func(ctx context.Context) settingsResult {
			e, err := s.SettingsClosedToggle(ctx, event.ID, event.Owner)
			return settingsResult{event: e, err: err}
		},
		func(ctx context.Context) settingsResult {
			e, old, err := s.SettingsLimitSet(ctx, event.ID, event.Owner, 7)
			return settingsResult{event: e, oldLimit: old, err: err}
		},
	})
	require.True(t, results[0].event.Settings.AutoPairing)
	require.True(t, results[1].event.Settings.Closed)
	require.Equal(t, 7, results[2].event.Settings.Limit)
	require.Equal(t, 2, results[2].oldLimit)
	stored, err := st.EventGet(context.Background(), event.ID)
	require.NoError(t, err)
	require.Equal(t, models.EventSettings{Limit: 7, AutoPairing: true, Closed: true}, stored.Settings)
	require.Eventually(t, func() bool {
		return historyActionCount(t, st, models.HistoryEventSettingsUpdated) == 3
	}, time.Second, 10*time.Millisecond)
	require.Zero(t, historyActionCount(t, st, models.HistoryNotificationSent))
}

func TestEventServiceConcurrentSettingsToggles(t *testing.T) {
	for _, tc := range []struct {
		name string
		run  func(*EventService, context.Context, string, models.Profile) (*models.Event, error)
	}{
		{"auto pairing", (*EventService).SettingsAutoPairingToggle},
		{"registration closure", (*EventService).SettingsClosedToggle},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s, st := newEventServiceTest(t)
			initial := models.EventSettings{Limit: 7, AutoPairing: true, Closed: true}
			event := &models.Event{ID: "event", Owner: models.Profile{ID: 1}, Settings: initial}
			require.NoError(t, st.EventUpsert(context.Background(), event))
			op := func(ctx context.Context) settingsResult {
				e, err := tc.run(s, ctx, event.ID, event.Owner)
				return settingsResult{event: e, err: err}
			}
			results := runConcurrentSettings(t, s, []settingsOperation{op, op})
			require.NotEqual(t, results[0].event.Settings, results[1].event.Settings)
			stored, err := st.EventGet(context.Background(), event.ID)
			require.NoError(t, err)
			require.Equal(t, initial, stored.Settings)
			require.Eventually(t, func() bool {
				return historyActionCount(t, st, models.HistoryEventSettingsUpdated) == 2
			}, time.Second, 10*time.Millisecond)
		})
	}
}

func TestEventServiceConcurrentSettingsLimits(t *testing.T) {
	s, st := newEventServiceTest(t)
	event := &models.Event{
		ID: "event", Owner: models.Profile{ID: 1},
		Settings: models.EventSettings{Limit: 2, AutoPairing: true, Closed: true},
		Couples:  append(sampleEvent().Couples, anotherSampleEvent().Couples[0]),
	}
	require.NoError(t, st.EventUpsert(context.Background(), event))
	ops := make([]settingsOperation, 0, 2)
	for _, limit := range []int{1, 3} {
		ops = append(ops, func(ctx context.Context) settingsResult {
			e, old, err := s.SettingsLimitSet(ctx, event.ID, event.Owner, limit)
			return settingsResult{event: e, oldLimit: old, err: err}
		})
	}
	results := runConcurrentSettings(t, s, ops)
	first, second := results[0], results[1]
	if first.oldLimit != 2 {
		first, second = second, first
	}
	require.Equal(t, 2, first.oldLimit)
	require.Equal(t, first.event.Settings.Limit, second.oldLimit)
	for _, result := range results {
		affected := NewEventHandler(result.event).LimitChangeGetAffected(result.oldLimit)
		switch result.event.Settings.Limit {
		case 1:
			require.Equal(t, result.oldLimit-1, len(affected.Couples))
			require.Equal(t, 1, affected.Position)
			require.False(t, affected.Increased)
		case 3:
			require.Equal(t, 3-result.oldLimit, len(affected.Couples))
			require.Equal(t, result.oldLimit, affected.Position)
			require.True(t, affected.Increased)
		}
	}
	stored, err := st.EventGet(context.Background(), event.ID)
	require.NoError(t, err)
	require.Equal(t, second.event.Settings, stored.Settings)
	require.True(t, stored.Settings.AutoPairing)
	require.True(t, stored.Settings.Closed)
	require.Eventually(t, func() bool {
		return historyActionCount(t, st, models.HistoryEventSettingsUpdated) == 2
	}, time.Second, 10*time.Millisecond)
}

func TestEventServiceRejectsUnauthorizedSettingsToggles(t *testing.T) {
	for _, run := range []func(*EventService, context.Context, string, models.Profile) (*models.Event, error){
		(*EventService).SettingsAutoPairingToggle,
		(*EventService).SettingsClosedToggle,
	} {
		s, st := newEventServiceTest(t)
		ctx := context.Background()
		event := &models.Event{ID: "event", Owner: models.Profile{ID: 1}, Settings: models.EventSettings{Limit: 7}}
		require.NoError(t, st.EventUpsert(ctx, event))
		updated, err := run(s, ctx, event.ID, models.Profile{ID: 2})
		require.ErrorContains(t, err, "profile is not allowed")
		require.Nil(t, updated)
		stored, err := st.EventGet(ctx, event.ID)
		require.NoError(t, err)
		require.Equal(t, event.Settings, stored.Settings)
		require.Zero(t, historyActionCount(t, st, models.HistoryEventSettingsUpdated))
		require.Zero(t, historyActionCount(t, st, models.HistoryNotificationSent))
	}
}

type settingsResult struct {
	event    *models.Event
	oldLimit int
	err      error
}

type settingsOperation func(context.Context) settingsResult

// Hold every operation before Begin so all contenders reach the transaction boundary together.
func runConcurrentSettings(t *testing.T, s *EventService, ops []settingsOperation) []settingsResult {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)
	arrivals := make(chan struct{}, len(ops))
	release := make(chan struct{})
	var once sync.Once
	unblock := func() { once.Do(func() { close(release) }) }
	defer unblock()
	s.store = &settingsBarrierStore{Store: s.store, arrivals: arrivals, release: release}

	type indexedResult struct {
		index int
		settingsResult
	}
	done := make(chan indexedResult, len(ops))
	for i, op := range ops {
		go func() { done <- indexedResult{index: i, settingsResult: op(ctx)} }()
	}
	for range ops {
		select {
		case <-arrivals:
		case <-ctx.Done():
			t.Fatal("settings operations did not reach Begin")
		}
	}
	unblock()
	results := make([]settingsResult, len(ops))
	for range ops {
		select {
		case result := <-done:
			require.NoError(t, result.err)
			require.NotNil(t, result.event)
			results[result.index] = result.settingsResult
		case <-ctx.Done():
			t.Fatal("settings operations did not finish")
		}
	}
	return results
}

type settingsBarrierStore struct {
	store.Store
	arrivals chan<- struct{}
	release  <-chan struct{}
}

func (s *settingsBarrierStore) Begin(ctx context.Context) (store.Store, error) {
	s.arrivals <- struct{}{}
	select {
	case <-s.release:
		return s.Store.Begin(ctx)
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (s *settingsBarrierStore) EventGet(context.Context, string) (*models.Event, error) {
	return nil, errors.New("event settings must be read inside the transaction")
}

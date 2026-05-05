package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/ofstudio/dancegobot/internal/config"
	"github.com/ofstudio/dancegobot/internal/models"
	storepkg "github.com/ofstudio/dancegobot/internal/store"
)

func TestRenderServiceEventRenderFail(t *testing.T) {
	ctx := context.Background()

	t.Run("increments fail counter below threshold", func(t *testing.T) {
		service, st := newRenderTestService(t, 2)
		event := testRenderEvent("render_fail_increment")
		require.NoError(t, st.EventUpsert(ctx, event))

		service.eventRenderFail(ctx, event.ID)

		got, err := st.EventGet(ctx, event.ID)
		require.NoError(t, err)
		require.NotNil(t, got)
		require.Equal(t, 1, got.RenderFails)
		require.False(t, got.Removed)

		var count int
		require.NoError(t, st.DB().GetContext(ctx, &count,
			`SELECT COUNT(*) FROM history WHERE event_id = ? AND action = ?`,
			event.ID, models.HistoryEventRemoved))
		require.Equal(t, 0, count)
	})

	t.Run("marks event removed at threshold", func(t *testing.T) {
		service, st := newRenderTestService(t, 2)
		event := testRenderEvent("render_fail_removed")
		event.RenderFails = 1
		require.NoError(t, st.EventUpsert(ctx, event))

		service.eventRenderFail(ctx, event.ID)

		got, err := st.EventGet(ctx, event.ID)
		require.NoError(t, err)
		require.NotNil(t, got)
		require.Equal(t, 2, got.RenderFails)
		require.True(t, got.Removed)

		var count int
		require.NoError(t, st.DB().GetContext(ctx, &count,
			`SELECT COUNT(*) FROM history WHERE event_id = ? AND action = ?`,
			event.ID, models.HistoryEventRemoved))
		require.Equal(t, 1, count)
	})

	t.Run("missing event is ignored", func(t *testing.T) {
		service, _ := newRenderTestService(t, 1)

		require.NotPanics(t, func() {
			service.eventRenderFail(ctx, "missing_event")
		})
	})
}

func TestRenderServiceErrIsPostRemoved(t *testing.T) {
	service, _ := newRenderTestService(t, 1)

	require.True(t, service.errIsPostRemoved(errors.New("telegram: Bad Request: MESSAGE_ID_INVALID")))
	require.False(t, service.errIsPostRemoved(errors.New("telegram: Bad Request: message is not modified")))
}

func newRenderTestService(t *testing.T, renderFailsMax int) (*RenderService, *storepkg.SQLiteStore) {
	t.Helper()

	cfg := config.Default()
	cfg.Settings.RenderFailsMax = renderFailsMax
	cfg.Settings.RendererRepeats = []time.Duration{}

	db, err := storepkg.NewSQLite(":memory:", cfg.DB.Version)
	require.NoError(t, err)

	st := storepkg.NewSQLiteStore(db)
	t.Cleanup(st.Close)

	service := NewRenderService(cfg.Settings, st, func(*models.Event, string) error {
		return nil
	})
	return service, st
}

func testRenderEvent(id string) *models.Event {
	return &models.Event{
		ID:      id,
		Caption: "Render fail event",
		Post: &models.Post{
			InlineMessageID: "inline_" + id,
		},
		Owner:     models.Profile{ID: 100, FirstName: "Owner"},
		CreatedAt: time.Date(2026, 5, 6, 12, 0, 0, 0, time.UTC),
	}
}

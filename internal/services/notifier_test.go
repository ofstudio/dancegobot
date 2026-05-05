package services

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

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

func newNotifierServiceTest(t *testing.T, notifyErr error) (*NotifierService, *storepkg.SQLiteStore) {
	t.Helper()

	cfg := config.Default()
	db, err := storepkg.NewSQLite(":memory:", cfg.DB.Version)
	require.NoError(t, err)

	st := storepkg.NewSQLiteStore(db)
	t.Cleanup(st.Close)

	service := NewNotifierService(cfg.Settings, st, func(*models.Notification) error {
		return notifyErr
	})
	return service, st
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

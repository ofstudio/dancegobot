package e2e

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/ofstudio/dancegobot/internal/config"
	"github.com/ofstudio/dancegobot/internal/models"
	"github.com/ofstudio/dancegobot/internal/store"
)

func TestNewEventNotificationRetries(t *testing.T) {
	tests := []struct {
		name        string
		code        int
		description string
		parameters  map[string]any
	}{
		{name: "server failure", code: 500, description: "Internal Server Error"},
		{name: "rate limit ignores retry_after", code: 429, description: "Too Many Requests: retry after 3600", parameters: map[string]any{"retry_after": 3600}},
		{name: "blocked bot", code: 403, description: "Forbidden: bot was blocked by the user"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := newEnvWithConfig(t, func(cfg *config.Config) {
				cfg.NotifierRepeats = []time.Duration{10 * time.Millisecond, 30 * time.Millisecond, 50 * time.Millisecond}
			})
			st, ok := env.app.Store.(*store.SQLiteStore)
			require.True(t, ok)
			first := testEvent("retry_source", "First event", userJane)
			require.NoError(t, env.app.Store.EventUpsert(context.Background(), first))
			publishSubscriptionEvent(env, first)
			subscribeToEvent(env, first, 800)

			second := testEvent("retry_target", "Second event", userJane)
			require.NoError(t, env.app.Store.EventUpsert(context.Background(), second))
			env.tg.RespondErrorWithParameters("sendMessage", tt.code, tt.description, tt.parameters)
			publishSubscriptionEvent(env, second)
			failed := env.waitSendMessage(userJohn.ID, "Новая запись")
			require.True(t, env.eventGet(second.ID).SubscribersNotified)

			// A repeated publication must not create another delivery task, even during retries.
			publishSubscriptionEvent(env, second)
			delivered := env.waitSendMessage(userJohn.ID, "Новая запись")
			require.Equal(t, failed.String("text"), delivered.String("text"))
			require.Equal(t, failed.InlineKeyboardRaw(), delivered.InlineKeyboardRaw())
			require.Contains(t, delivered.InlineKeyboardRaw(), second.ID)

			require.Eventually(t, func() bool {
				var count int
				err := st.DB().Get(&count,
					"SELECT COUNT(*) FROM history WHERE action = ? AND event_id = ?",
					models.HistoryNotificationSent, second.ID)
				if err != nil {
					return false
				}
				return count == 1
			}, time.Second, time.Millisecond)
			var data string
			require.NoError(t, st.DB().Get(&data,
				"SELECT data FROM history WHERE action = ? AND event_id = ?",
				models.HistoryNotificationSent, second.ID))
			require.NotContains(t, data, `"error"`)
			// Let all configured retry slots elapse before asserting no extra requests.
			time.Sleep(60 * time.Millisecond)
			env.tg.AssertNoUnexpected()
		})
	}
}

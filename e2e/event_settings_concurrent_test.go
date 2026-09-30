package e2e

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	tele "gopkg.in/telebot.v4"

	"github.com/ofstudio/dancegobot/internal/models"
	"github.com/ofstudio/dancegobot/internal/telegram/views"
	"github.com/ofstudio/dancegobot/pkg/teletest"
)

func TestEventSettingsConcurrentCallbacks(t *testing.T) {
	env := newEnv(t)
	event := testLimitEvent("concurrent_settings", userJohn, 7)
	require.NoError(t, env.app.Store.EventUpsert(context.Background(), event))
	updates := []tele.Update{
		env.callback(tele.Callback{
			Sender: userJohn, Message: &tele.Message{ID: 100, Chat: privateChat(userJohn)},
			Data: "\f" + views.BtnEventSettingsAutoPair.Unique + "|" + event.ID + "|0|rand",
		}),
		env.callback(tele.Callback{
			Sender: userJohn, Message: &tele.Message{ID: 101, Chat: privateChat(userJohn)},
			Data: "\f" + views.BtnEventSettingsClose.Unique + "|" + event.ID + "|0|rand",
		}),
	}
	start := make(chan struct{})
	done := make(chan struct{}, len(updates))
	for _, update := range updates {
		go func() {
			<-start
			env.process(update)
			done <- struct{}{}
		}()
	}
	close(start)
	for range updates {
		select {
		case <-done:
		case <-time.After(5 * time.Second):
			t.Fatal("concurrent settings callbacks did not finish")
		}
	}
	for _, update := range updates {
		answer := env.tg.WaitFor("answerCallbackQuery", func(req teletest.Request) bool {
			return req.String("callback_query_id") == update.Callback.ID
		})
		require.Empty(t, answer.String("text"))
		require.False(t, answer.JSON.Get("show_alert").Bool())
		env.tg.WaitFor("editMessageText", func(req teletest.Request) bool {
			return req.ChatIDInt() == userJohn.ID && req.Int("message_id") == update.Callback.Message.ID
		})
		env.tg.WaitFor("editMessageText", func(req teletest.Request) bool {
			return req.String("inline_message_id") == event.Post.InlineMessageID
		})
	}
	require.Equal(t, models.EventSettings{Limit: 7, AutoPairing: true, Closed: true}, env.eventGet(event.ID).Settings)
}

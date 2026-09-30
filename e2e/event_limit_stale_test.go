package e2e

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	tele "gopkg.in/telebot.v4"

	"github.com/ofstudio/dancegobot/internal/locale"
	"github.com/ofstudio/dancegobot/internal/models"
	"github.com/ofstudio/dancegobot/pkg/teletest"
)

func TestEventLimitDelayedConfirmation(t *testing.T) {
	env := newEnv(t)
	frank := &tele.User{ID: 206, FirstName: "Frank"}
	event := testLimitEvent("delayed_limit_confirmation", userJohn, 4,
		testLimitCouple(limitUserAlice, limitUserBob, limitUserAlice, 0, false),
		testLimitCouple(userJohn, limitUserDan, userJohn, time.Second, false),
		testLimitCouple(userJane, limitUserCarol, userJane, 2*time.Second, false),
		testLimitCouple(limitUserEve, frank, limitUserEve, 3*time.Second, false),
	)
	require.NoError(t, env.app.Store.EventUpsert(env.ctx, event))

	env.process(env.callback(tele.Callback{
		Sender: userJohn, Message: &tele.Message{ID: 100, Chat: privateChat(userJohn)},
		Data: "\fevt_set_lim_num|" + event.ID + "|2|0",
	}))
	env.tg.Wait("answerCallbackQuery")
	env.tg.WaitFor("editMessageText", func(req teletest.Request) bool {
		return req.String("inline_message_id") == event.Post.InlineMessageID
	})
	env.tg.WaitFor("editMessageText", func(req teletest.Request) bool {
		return req.ChatIDInt() == userJohn.ID
	})
	prompt := env.waitSendMessage(userJohn.ID, "Лимит пар уменьшился")
	require.Contains(t, prompt.InlineKeyboardRaw(), "lim_chg_ntf|"+event.ID)

	// Jane's couple moves into the active list while the owner has not confirmed the old notice.
	env.process(env.message(limitUserAlice, "/start "+signupPayload(event.ID, models.RoleLeader)))
	env.waitSendMessage(limitUserAlice.ID, "Bob")
	env.process(env.message(limitUserAlice, locale.BtnDancerRemove))
	env.waitSendMessage(limitUserAlice.ID, locale.ResultSuccessRemoved)
	env.tg.WaitFor("editMessageText", func(req teletest.Request) bool {
		return req.String("inline_message_id") == event.Post.InlineMessageID &&
			!strings.Contains(req.String("text"), "Alice")
	})
	env.waitSendMessage(userJane.ID, "вышли из списка ожидания")
	reg, err := env.app.Services.Event.RegistrationGet(context.Background(), event.ID,
		models.NewProfile(*userJane), models.RoleLeader)
	require.NoError(t, err)
	require.False(t, reg.WaitList)

	env.process(env.callback(tele.Callback{
		Sender: userJohn, Message: &tele.Message{ID: 101, Chat: privateChat(userJohn)},
		Data: "\flim_chg_ntf|" + event.ID + "|rand",
	}))
	env.tg.Wait("answerCallbackQuery")
	env.tg.Wait("editMessageReplyMarkup")
	env.waitSendMessage(userJohn.ID, locale.LimitChangedNotified)
	notification := env.waitSendMessage(limitUserEve.ID, "теперь в списке ожидания")
	require.Contains(t, notification.String("text"), "Frank")
	reg, err = env.app.Services.Event.RegistrationGet(context.Background(), event.ID,
		models.NewProfile(*limitUserEve), models.RoleLeader)
	require.NoError(t, err)
	require.True(t, reg.WaitList)

	// Eve is last in the saved batch, so waiting for her notification drains the fanout.
	env.tg.AssertNoUnexpected()
}

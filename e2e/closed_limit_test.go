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

func TestEventClosedRegistration(t *testing.T) {
	t.Run("owner can close event registration", func(t *testing.T) {
		env := newEnv(t)
		event := testEvent("closed_registration", "Closed registration", userJohn)
		require.NoError(t, env.app.Store.EventUpsert(context.Background(), event))

		env.process(env.callback(tele.Callback{
			Sender:  userJohn,
			Message: &tele.Message{ID: 100, Chat: privateChat(userJohn)},
			Data:    "\fevt_set_close|" + event.ID + "|0|rand",
		}))
		env.tg.Wait("answerCallbackQuery")
		editPost := env.tg.WaitFor("editMessageText", func(req teletest.Request) bool {
			return req.String("inline_message_id") == event.Post.InlineMessageID
		})
		require.True(t, strings.HasPrefix(editPost.String("text"), locale.IconPostClosed))
		require.Contains(t, editPost.InlineKeyboardRaw(), "post_closed")
		editSettings := env.tg.WaitFor("editMessageText", func(req teletest.Request) bool {
			return req.ChatIDInt() == userJohn.ID
		})
		require.Contains(t, editSettings.String("text"), locale.EventSettingsClosed[true])

		require.True(t, env.eventGet(event.ID).Settings.Closed)

		env.process(env.message(userJane, "/start "+signupPayload(event.ID, models.RoleLeader)))
		env.waitSendMessage(userJane.ID, locale.ResultEventClosed)
	})

	t.Run("closed post button returns closed notice", func(t *testing.T) {
		env := newEnv(t)
		env.process(env.callback(tele.Callback{
			Sender:    userJane,
			MessageID: "inline_closed_registration",
			Data:      "\fpost_closed|rand",
		}))
		resp := env.tg.Wait("answerCallbackQuery")
		require.Equal(t, locale.ResultEventClosed, resp.String("text"))
	})
}

func TestEventLimitWaitlistSignup(t *testing.T) {
	env := newEnv(t)
	event := testLimitEvent("limit_waitlist_signup", userJane, 1,
		testLimitCouple(limitUserAlice, limitUserBob, limitUserAlice, 0, false),
	)
	require.NoError(t, env.app.Store.EventUpsert(context.Background(), event))

	env.process(env.message(userJohn, "/start "+signupPayload(event.ID, models.RoleLeader)))
	env.waitSendMessage(userJohn.ID, locale.SignupNotRegistered)

	env.process(env.message(userJohn, "Manual Partner"))
	env.tg.WaitFor("editMessageText", func(req teletest.Request) bool {
		return req.String("inline_message_id") == event.Post.InlineMessageID &&
			strings.Contains(req.String("text"), locale.PostCouplesWait) &&
			strings.Contains(req.String("text"), "Manual Partner")
	})
	env.waitSendMessage(userJohn.ID, locale.ResultCoupleWaitlist)

	reg, err := env.app.EventService.RegistrationGet(context.Background(), event.ID,
		models.NewProfile(*userJohn), models.RoleLeader)
	require.NoError(t, err)
	require.True(t, reg.WaitList)
}

func TestEventLimitWaitlistLeftAfterRemoval(t *testing.T) {
	env := newEnv(t)
	event := testLimitEvent("limit_waitlist_left", userJohn, 1,
		testLimitCouple(userJohn, limitUserBob, userJohn, 0, false),
		testLimitCouple(userJane, limitUserCarol, userJane, time.Second, false),
	)
	require.NoError(t, env.app.Store.EventUpsert(context.Background(), event))

	env.process(env.message(userJohn, "/start "+signupPayload(event.ID, models.RoleLeader)))
	env.waitSendMessage(userJohn.ID, "Bob")

	env.process(env.message(userJohn, locale.BtnDancerRemove))
	env.tg.WaitFor("editMessageText", func(req teletest.Request) bool {
		return req.String("inline_message_id") == event.Post.InlineMessageID &&
			strings.Contains(req.String("text"), "Jane Doe") &&
			!strings.Contains(req.String("text"), locale.PostCouplesWait)
	})
	env.waitSendMessage(userJohn.ID, locale.ResultSuccessRemoved)
	env.tg.WaitFor("sendMessage", func(req teletest.Request) bool {
		return req.ChatIDInt() == userJane.ID &&
			strings.Contains(req.String("text"), "Carol") &&
			strings.Contains(req.String("text"), "списка ожидания")
	})

	reg, err := env.app.EventService.RegistrationGet(context.Background(), event.ID,
		models.NewProfile(*userJane), models.RoleLeader)
	require.NoError(t, err)
	require.False(t, reg.WaitList)
}

func TestEventLimitChangeNotification(t *testing.T) {
	env := newEnv(t)
	event := testLimitEvent("limit_change_notify", userJohn, 1,
		testLimitCouple(limitUserAlice, limitUserBob, limitUserAlice, 0, false),
		testLimitCouple(userJane, limitUserCarol, userJane, time.Second, false),
	)
	require.NoError(t, env.app.Store.EventUpsert(context.Background(), event))

	env.process(env.callback(tele.Callback{
		Sender:  userJohn,
		Message: &tele.Message{ID: 100, Chat: privateChat(userJohn)},
		Data:    "\fevt_set_lim_num|" + event.ID + "|2|0",
	}))
	env.tg.Wait("answerCallbackQuery")
	env.tg.WaitFor("editMessageText", func(req teletest.Request) bool {
		return req.String("inline_message_id") == event.Post.InlineMessageID &&
			!strings.Contains(req.String("text"), locale.PostCouplesWait)
	})
	env.tg.WaitFor("editMessageText", func(req teletest.Request) bool {
		return req.ChatIDInt() == userJohn.ID && strings.Contains(req.String("text"), "Приходят первые 2 пары")
	})
	prompt := env.tg.WaitFor("sendMessage", func(req teletest.Request) bool {
		return req.ChatIDInt() == userJohn.ID && strings.Contains(req.String("text"), "Лимит пар увеличился")
	})
	require.Contains(t, prompt.InlineKeyboardRaw(), "lim_chg_ntf")

	env.process(env.callback(tele.Callback{
		Sender:  userJohn,
		Message: &tele.Message{ID: 101, Chat: privateChat(userJohn)},
		Data:    "\flim_chg_ntf|" + event.ID + "|rand",
	}))
	env.tg.Wait("answerCallbackQuery")
	env.tg.Wait("editMessageReplyMarkup")
	env.waitSendMessage(userJohn.ID, locale.LimitChangedNotified)
	env.waitSendMessage(userJane.ID, "увеличил лимит пар")
}

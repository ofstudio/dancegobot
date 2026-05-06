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

func TestEventAutoPairSignup(t *testing.T) {
	env := newEnv(t)
	event := testEvent("auto_pair_signup", "Auto pair signup", userJohn)
	event.Settings.AutoPairing = true
	require.NoError(t, env.app.Store.EventUpsert(context.Background(), event))

	env.process(env.message(userJohn, "/start "+signupPayload(event.ID, models.RoleLeader)))
	req := env.waitSendMessage(userJohn.ID, locale.SignupNotRegistered)
	require.Contains(t, req.ReplyMarkup().Get("keyboard").Raw, locale.BtnAsSingle[models.RoleLeader])

	env.process(env.message(userJohn, locale.BtnAsSingle[models.RoleLeader]))
	env.tg.WaitFor("editMessageText", func(req teletest.Request) bool {
		return req.String("inline_message_id") == event.Post.InlineMessageID &&
			strings.Contains(req.String("text"), "John") &&
			strings.Contains(req.String("text"), locale.PostSingles[models.RoleLeader])
	})
	env.waitSendMessage(userJohn.ID, "Добавил тебя в список ищущих пару")

	env.process(env.message(userJane, "/start "+signupPayload(event.ID, models.RoleFollower)))
	req = env.waitSendMessage(userJane.ID, locale.SignupNotRegistered)
	require.Contains(t, req.ReplyMarkup().Get("keyboard").Raw, locale.BtnAsSingle[models.RoleFollower])

	env.process(env.message(userJane, locale.BtnAsSingle[models.RoleFollower]))
	env.tg.WaitFor("editMessageText", func(req teletest.Request) bool {
		return req.String("inline_message_id") == event.Post.InlineMessageID &&
			strings.Contains(req.String("text"), "John") &&
			strings.Contains(req.String("text"), "Jane Doe") &&
			!strings.Contains(req.String("text"), locale.PostSingles[models.RoleLeader])
	})
	env.tg.WaitFor("sendMessage", func(req teletest.Request) bool {
		return req.ChatIDInt() == userJane.ID &&
			strings.Contains(req.String("text"), "Вы зарегистрировались в паре") &&
			strings.Contains(req.String("text"), "John")
	})
	env.tg.WaitFor("sendMessage", func(req teletest.Request) bool {
		return req.ChatIDInt() == userJohn.ID &&
			strings.Contains(req.String("text"), "Я подобрал тебе в пару") &&
			strings.Contains(req.String("text"), "Jane Doe")
	})

	reg, err := env.app.EventService.RegistrationGet(context.Background(), event.ID,
		models.NewProfile(*userJohn), models.RoleLeader)
	require.NoError(t, err)
	require.Equal(t, models.StatusInCouple, reg.Status)
	require.False(t, reg.WaitList)
}

func TestEventManualSingleNumberSignup(t *testing.T) {
	env := newEnv(t)
	event := testEvent("manual_single_number_signup", "Manual single number signup", userJohn)
	jane := models.NewProfile(*userJane)
	event.Singles = []models.Dancer{{
		Profile:   &jane,
		FullName:  jane.FullName(),
		Role:      models.RoleFollower,
		AsSingle:  true,
		CreatedAt: time.Now().UTC(),
	}}
	require.NoError(t, env.app.Store.EventUpsert(context.Background(), event))

	env.process(env.message(userJohn, "/start "+signupPayload(event.ID, models.RoleLeader)))
	req := env.waitSendMessage(userJohn.ID, locale.SignupNotRegistered)
	require.Contains(t, req.ReplyMarkup().Get("keyboard").Raw, "1. Jane Doe (@jane_doe)")
	require.NotContains(t, req.ReplyMarkup().Get("keyboard").Raw, locale.BtnAsSingle[models.RoleLeader])

	env.process(env.message(userJohn, "1."))
	env.tg.WaitFor("editMessageText", func(req teletest.Request) bool {
		return strings.Contains(req.String("text"), "John") &&
			strings.Contains(req.String("text"), "Jane Doe")
	})
	env.tg.WaitFor("sendMessage", func(req teletest.Request) bool {
		return req.ChatIDInt() == userJohn.ID &&
			strings.Contains(req.String("text"), "Вы зарегистрировались в паре") &&
			strings.Contains(req.String("text"), "Jane Doe")
	})
	env.tg.WaitFor("sendMessage", func(req teletest.Request) bool {
		return req.ChatIDInt() == userJane.ID &&
			strings.Contains(req.String("text"), "зарегистрировался с тобой") &&
			strings.Contains(req.String("text"), "John")
	})
}

func TestManualPartnerHTMLEscaped(t *testing.T) {
	env := newEnv(t)
	event := testEvent("manual_partner_html_escape", "Manual partner HTML escape", userJohn)
	require.NoError(t, env.app.Store.EventUpsert(context.Background(), event))

	env.process(env.message(userJohn, "/start "+signupPayload(event.ID, models.RoleLeader)))
	env.waitSendMessage(userJohn.ID, locale.SignupNotRegistered)

	env.process(env.message(userJohn, "Мария <Follower> & Co"))
	edit := env.tg.Wait("editMessageText")
	require.Contains(t, edit.String("text"), "Мария &lt;Follower&gt; &amp; Co")
	require.NotContains(t, edit.String("text"), "Мария <Follower>")
	env.waitSendMessage(userJohn.ID, "Вы зарегистрировались в паре")
}

func TestSignupSessionResetBySettings(t *testing.T) {
	env := newEnv(t)
	event := testEvent("signup_session_settings_reset", "Signup session reset", userJohn)
	require.NoError(t, env.app.Store.EventUpsert(context.Background(), event))

	env.process(env.message(userJohn, "/start "+signupPayload(event.ID, models.RoleLeader)))
	env.waitSendMessage(userJohn.ID, locale.SignupNotRegistered)

	env.process(env.message(userJohn, "/settings"))
	env.waitSendMessage(userJohn.ID, locale.UserSettingsDescription)

	env.process(env.message(userJohn, "Manual Partner"))
	require.Empty(t, env.eventGet(event.ID).Couples)
}

func TestEventSettingsAccessControl(t *testing.T) {
	env := newEnv(t)
	event := testEvent("settings_access_control", "Settings access", userJohn)
	require.NoError(t, env.app.Store.EventUpsert(context.Background(), event))

	env.process(env.callback(tele.Callback{
		Sender:  userJane,
		Message: &tele.Message{ID: 100, Chat: privateChat(userJane)},
		Data:    "\fevt_set_auto_pair|" + event.ID + "|0|rand",
	}))

	resp := env.tg.Wait("answerCallbackQuery")
	require.Equal(t, locale.ErrSomethingWrong, resp.String("text"))
	require.True(t, resp.Bool("show_alert"))
	require.False(t, env.eventGet(event.ID).Settings.AutoPairing)
}

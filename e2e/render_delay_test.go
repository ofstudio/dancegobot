package e2e

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	tele "gopkg.in/telebot.v4"

	"github.com/ofstudio/dancegobot/internal/config"
	"github.com/ofstudio/dancegobot/internal/locale"
	"github.com/ofstudio/dancegobot/internal/models"
)

func TestInitialRenderDelayPublicationAndSignup(t *testing.T) {
	for _, tc := range []struct {
		name          string
		chatFirst     bool
		callbackFirst bool
	}{
		{name: "chosen result before chat"},
		{name: "chat before chosen result", chatFirst: true},
		{name: "callback without chosen result", callbackFirst: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			env := newEnvWithConfig(t, enableInitialRenderDelay)
			created, err := env.app.Store.SubscriptionCreate(env.ctx, &models.Subscription{
				Subscriber: models.NewProfile(*userJohn),
				Chat:       models.NewChat(chatSuperGroup),
				CreatedAt:  time.Now(),
			})
			require.NoError(t, err)
			require.True(t, created)
			eventID := env.eventDraftCreate(queryA)
			inlineID := "inline_delayed_signup"
			post := env.channelPost(tele.Message{
				ID: 12345, Sender: userJohn, Chat: chatSuperGroup, Text: queryA.Text, Via: botUser,
				ReplyMarkup: &tele.ReplyMarkup{InlineKeyboard: [][]tele.InlineButton{{
					{Text: locale.RoleIcon[models.RoleLeader], Data: "\fsignup|" + eventID + "|leader|v1"},
				}}},
			})
			callback := env.callback(tele.Callback{
				Sender: userJohn, MessageID: inlineID,
				Data: "\fsignup|" + eventID + "|leader|v1",
			})
			if tc.chatFirst {
				env.process(post)
				require.Empty(t, env.eventGet(eventID).Post.InlineMessageID)
			}
			started := time.Now()
			if tc.callbackFirst {
				env.process(callback)
			} else {
				env.process(env.inlineResult(tele.InlineResult{
					Sender: userJohn, ResultID: eventID, Query: queryA.Text, MessageID: inlineID,
				}))
			}
			if !tc.chatFirst {
				env.process(post)
			}
			if !tc.callbackFirst {
				env.process(callback)
			}
			answer := env.tg.Wait("answerCallbackQuery")
			require.Regexp(t, rxUrlSignupLeader, answer.String("url"))
			notification := env.waitSendMessage(userJohn.ID, "Новая запись")
			require.Contains(t, notification.InlineKeyboardRaw(), "https://t.me/c/1234567890/12345")

			env.process(env.message(userJohn, "/start "+signupPayload(eventID, models.RoleLeader)))
			env.waitSendMessage(userJohn.ID, locale.SignupNotRegistered)
			env.process(env.message(userJohn, "Manual Partner"))
			env.waitSendMessage(userJohn.ID, "Вы зарегистрировались в паре")
			event := env.eventGet(eventID)
			require.Len(t, event.Couples, 1)
			require.Equal(t, inlineID, event.Post.InlineMessageID)
			require.Equal(t, 12345, event.Post.ChatMessageID)
			require.Equal(t, chatSuperGroup.ID, event.Post.Chat.ID)
			require.Less(t, time.Since(started), config.Default().RendererInitialDelay,
				"callback, subscription notification and signup must not wait for the post edit")
			env.tg.AssertNoUnexpected()

			edit := env.tg.Wait("editMessageText", 8*time.Second)
			require.GreaterOrEqual(t, time.Since(started), config.Default().RendererInitialDelay)
			require.Equal(t, inlineID, edit.String("inline_message_id"))
			require.Contains(t, edit.String("text"), "John")
			require.Contains(t, edit.String("text"), "Manual Partner")
			require.Contains(t, edit.InlineKeyboardRaw(), "signup-"+eventID+"-leader")
		})
	}
}

func TestInitialRenderDelayDoesNotDelayClosure(t *testing.T) {
	env := newEnvWithConfig(t, enableInitialRenderDelay)
	eventID := env.eventDraftCreate(queryA)
	inlineID := "inline_delayed_closure"
	started := time.Now()
	env.process(env.inlineResult(tele.InlineResult{
		Sender: userJohn, ResultID: eventID, Query: queryA.Text, MessageID: inlineID,
	}))
	env.process(env.callback(tele.Callback{
		Sender:  userJohn,
		Message: &tele.Message{ID: 100, Chat: privateChat(userJohn)},
		Data:    "\fevt_set_close|" + eventID + "|0|v1",
	}))
	env.tg.Wait("answerCallbackQuery")
	settings := env.tg.Wait("editMessageText")
	require.Equal(t, userJohn.ID, settings.ChatIDInt())
	require.Contains(t, settings.String("text"), locale.EventSettingsClosed[true])
	require.True(t, env.eventGet(eventID).Settings.Closed)

	// The original role button is still visible, but registration already uses the new settings.
	env.process(env.callback(tele.Callback{
		Sender: userJane, MessageID: inlineID, Data: "\fsignup|" + eventID + "|leader|v1",
	}))
	require.Regexp(t, rxUrlSignupLeader, env.tg.Wait("answerCallbackQuery").String("url"))
	env.process(env.message(userJane, "/start "+signupPayload(eventID, models.RoleLeader)))
	env.waitSendMessage(userJane.ID, locale.ResultEventClosed)
	require.Empty(t, env.eventGet(eventID).Couples)
	require.Less(t, time.Since(started), config.Default().RendererInitialDelay)
	env.tg.AssertNoUnexpected()

	edit := env.tg.Wait("editMessageText", 8*time.Second)
	require.GreaterOrEqual(t, time.Since(started), config.Default().RendererInitialDelay)
	require.Equal(t, inlineID, edit.String("inline_message_id"))
	require.Contains(t, edit.String("text"), locale.IconPostClosed)
	require.Contains(t, edit.InlineKeyboardRaw(), "post_closed|v1")
}

func enableInitialRenderDelay(cfg *config.Config) {
	// Exercise the experimental delay that may reduce missing mobile posts: https://github.com/ofstudio/dancegobot/issues/21
	cfg.RendererInitialDelay = config.Default().RendererInitialDelay
}

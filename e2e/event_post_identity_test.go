package e2e

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	tele "gopkg.in/telebot.v4"

	"github.com/ofstudio/dancegobot/internal/locale"
	"github.com/ofstudio/dancegobot/internal/models"
	"github.com/ofstudio/dancegobot/pkg/teletest"
)

func TestEventPostOriginalIdentity(t *testing.T) {
	for _, viaCallback := range []bool{false, true} {
		t.Run(map[bool]string{false: "chosen inline result", true: "signup callback"}[viaCallback], func(t *testing.T) {
			env := newEnv(t)
			eventID := env.eventDraftCreate(queryA)
			const originalID = "original-inline-message"
			publish := func(id string, conflict bool) {
				if viaCallback {
					env.process(env.callback(tele.Callback{
						Sender: userJohn, MessageID: id,
						Data: "\fsignup|" + eventID + "|leader|v1",
					}))
					answer := env.tg.Wait("answerCallbackQuery")
					if conflict {
						require.Equal(t, locale.ErrSomethingWrong, answer.String("text"))
						require.True(t, answer.Bool("show_alert"))
						require.Empty(t, answer.String("url"))
						return
					}
					require.Regexp(t, rxUrlSignupLeader, answer.String("url"))
				} else {
					env.process(env.inlineResult(tele.InlineResult{
						Sender: userJohn, ResultID: eventID, Query: queryA.Text, MessageID: id,
					}))
					if conflict {
						env.waitSendMessage(userJohn.ID, locale.ErrSomethingWrong)
						return
					}
				}
				edit := env.tg.Wait("editMessageText")
				require.Equal(t, originalID, edit.String("inline_message_id"))
			}
			publish(originalID, false)
			publishSubscriptionEvent(env, env.eventGet(eventID))
			before := env.eventGet(eventID)
			require.True(t, before.SubscribersNotified)

			publish("replacement-inline-message", true)
			env.process(env.channelPost(tele.Message{
				ID: 501, Sender: userJohn, Chat: chatSuperGroup, Via: botUser,
				ReplyMarkup: &tele.ReplyMarkup{InlineKeyboard: [][]tele.InlineButton{{{
					Text: locale.RoleIcon[models.RoleLeader],
					Data: "\fsignup|" + eventID + "|leader|v1",
				}}}},
			}))
			require.Equal(t, before, env.eventGet(eventID))
			env.tg.AssertNoUnexpected()

			// Replaying the original update still renders the original post and permits signup.
			publish(originalID, false)
			require.Equal(t, before, env.eventGet(eventID))
			env.process(env.message(userJane, "/start "+signupPayload(eventID, models.RoleLeader)))
			env.waitSendMessage(userJane.ID, locale.SignupNotRegistered)
			env.process(env.message(userJane, "Manual Partner"))
			edit := env.tg.WaitFor("editMessageText", func(req teletest.Request) bool {
				return req.String("inline_message_id") == originalID &&
					strings.Contains(req.String("text"), "Manual Partner")
			})
			require.Contains(t, edit.String("text"), userJane.FirstName)
			env.waitSendMessage(userJane.ID, "Вы зарегистрировались в паре")
			after := env.eventGet(eventID)
			require.Equal(t, before.Post, after.Post)
			require.True(t, after.SubscribersNotified)
			require.Len(t, after.Couples, 1)

			env.process(env.message(userJane, "/my"))
			my := env.waitSendMessage(userJane.ID, "Manual Partner")
			link := fmt.Sprintf("https://t.me/c/%d/%d", -before.Post.Chat.ID-1000000000000, before.Post.ChatMessageID)
			require.Contains(t, my.InlineKeyboardRaw(), link)
		})
	}
}

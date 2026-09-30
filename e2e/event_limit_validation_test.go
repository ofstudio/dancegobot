package e2e

import (
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	tele "gopkg.in/telebot.v4"

	"github.com/ofstudio/dancegobot/internal/locale"
	"github.com/ofstudio/dancegobot/internal/models"
	"github.com/ofstudio/dancegobot/internal/telegram/views"
	"github.com/ofstudio/dancegobot/pkg/teletest"
)

func TestEventSettingsInvalidLimitCallbacks(t *testing.T) {
	for _, tc := range []struct {
		name   string
		unique string
		values []string
	}{
		{"limit", views.BtnEventSettingsLimitNum.Unique,
			[]string{"-1", "21", "", "abc", "1.5", "999999999999999999999999999999"}},
		{"page", views.BtnEventSettingsLimit.Unique,
			[]string{"-1", "2", "", "abc", "1.5", "999999999999999999999999999999"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			for _, value := range tc.values {
				t.Run(value, func(t *testing.T) {
					env := newEnv(t)
					event := testLimitEvent("invalid_limit", userJohn, 1,
						testLimitCouple(limitUserAlice, limitUserBob, limitUserAlice, 0, false),
						testLimitCouple(userJane, limitUserCarol, userJane, time.Second, false),
					)
					require.NoError(t, env.app.Store.EventUpsert(env.ctx, event))
					before := env.eventGet(event.ID)
					profile := models.NewProfile(*userJohn)
					user, err := env.app.Services.User.Get(env.ctx, profile)
					require.NoError(t, err)
					user.Session = models.Session{Action: models.SessionSignup, EventID: event.ID}
					require.NoError(t, env.app.Services.User.UpdateSession(env.ctx, user))

					require.NotPanics(t, func() {
						env.process(env.callback(tele.Callback{
							Sender:  userJohn,
							Message: &tele.Message{ID: 100, Chat: privateChat(userJohn)},
							Data:    "\f" + tc.unique + "|" + event.ID + "|" + value + "|0|rand",
						}))
					})
					answer := env.tg.Wait("answerCallbackQuery")
					require.Equal(t, locale.ErrSomethingWrong, answer.String("text"))
					require.True(t, answer.Bool("show_alert"))
					require.Equal(t, before, env.eventGet(event.ID))
					user, err = env.app.Services.User.Get(env.ctx, profile)
					require.NoError(t, err)
					require.Equal(t, models.Session{}, user.Session)
					env.tg.AssertNoUnexpected()
				})
			}
		})
	}
}

func TestEventSettingsValidLimitCallbacks(t *testing.T) {
	for _, limit := range []int{0, 1, 20} {
		t.Run(strconv.Itoa(limit), func(t *testing.T) {
			env := newEnv(t)
			event := testLimitEvent("valid_limit", userJohn, 2)
			require.NoError(t, env.app.Store.EventUpsert(env.ctx, event))

			env.process(env.callback(tele.Callback{
				Sender:  userJohn,
				Message: &tele.Message{ID: 100, Chat: privateChat(userJohn)},
				Data:    "\fevt_set_lim_num|" + event.ID + "|" + strconv.Itoa(limit) + "|0",
			}))
			answer := env.tg.Wait("answerCallbackQuery")
			require.Empty(t, answer.String("text"))
			edit := env.tg.WaitFor("editMessageText", func(req teletest.Request) bool {
				return req.ChatIDInt() == userJohn.ID
			})
			require.Contains(t, edit.String("text"), locale.EventSettingsCaption)
			env.tg.WaitFor("editMessageText", func(req teletest.Request) bool {
				return req.String("inline_message_id") == event.Post.InlineMessageID
			})
			require.Equal(t, models.EventSettings{Limit: limit}, env.eventGet(event.ID).Settings)
		})
	}
}

func TestEventSettingsValidLimitPages(t *testing.T) {
	for _, page := range []int{0, 1} {
		t.Run(strconv.Itoa(page), func(t *testing.T) {
			env := newEnv(t)
			event := testLimitEvent("limit_pages", userJohn, 2)
			require.NoError(t, env.app.Store.EventUpsert(env.ctx, event))
			env.process(env.callback(tele.Callback{
				Sender:  userJohn,
				Message: &tele.Message{ID: 100, Chat: privateChat(userJohn)},
				Data:    "\fevt_set_lim|" + event.ID + "|" + strconv.Itoa(page) + "|0|rand",
			}))
			answer := env.tg.Wait("answerCallbackQuery")
			require.Empty(t, answer.String("text"))
			markup := env.tg.Wait("editMessageReplyMarkup")
			rows := markup.ReplyMarkup().Get("inline_keyboard").Array()
			require.Len(t, rows, 4)
			require.Contains(t, rows[0].Get("0.callback_data").String(), "evt_set_lim_num|"+event.ID+"|0|0|")
			for i := 0; i < 2; i++ {
				buttons := rows[i+1].Array()
				require.Len(t, buttons, 5)
				for j, button := range buttons {
					value := strconv.Itoa(page*10 + i*5 + j + 1)
					require.Equal(t, value, button.Get("text").String())
					require.Equal(t, "\fevt_set_lim_num|"+event.ID+"|"+value+"|0", button.Get("callback_data").String())
				}
			}
			require.Equal(t, event.Settings, env.eventGet(event.ID).Settings)
		})
	}
}

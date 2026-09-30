package e2e

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	tele "gopkg.in/telebot.v4"

	"github.com/ofstudio/dancegobot/internal/locale"
	"github.com/ofstudio/dancegobot/internal/models"
	"github.com/ofstudio/dancegobot/internal/telegram/views"
	"github.com/ofstudio/dancegobot/pkg/teletest"
)

func TestStart(t *testing.T) {
	env := newEnv(t)
	env.process(env.message(&tele.User{ID: 123456}, "/start"))
	req := env.waitSendMessage(123456, "Привет! Это бот")
	require.Contains(t, req.String("text"), "@test_bot")
}

func TestMyCommand(t *testing.T) {
	t.Run("no events", func(t *testing.T) {
		env := newEnv(t)
		env.process(env.message(userJohn, "/my"))
		req := env.waitSendMessage(userJohn.ID, locale.MyNoEvents)
		require.Equal(t, locale.BtnTry, req.ReplyMarkup().Get("inline_keyboard.0.0.text").String())
	})

	t.Run("owned event", func(t *testing.T) {
		env := newEnv(t)
		event := testEvent("my_owned_event", "Owner event", userJohn)
		require.NoError(t, env.app.Store.EventUpsert(context.Background(), event))

		env.process(env.message(userJohn, "/my"))
		req := env.waitSendMessage(userJohn.ID, "Owner event")
		require.Contains(t, req.InlineKeyboardRaw(), "signup-my_owned_event-leader")
		require.Contains(t, req.InlineKeyboardRaw(), "my_refresh|0")
		require.Contains(t, req.InlineKeyboardRaw(), "evt_set|my_owned_event|0")
	})

	t.Run("joined event", func(t *testing.T) {
		env := newEnv(t)
		event := testEvent("my_joined_event", "Joined event", userJane)
		john := models.NewProfile(*userJohn)
		event.Couples = []models.Couple{{
			Dancers: []models.Dancer{
				{Profile: &john, FullName: john.FullName(), Role: models.RoleLeader},
				{FullName: "Test Partner", Role: models.RoleFollower},
			},
			CreatedBy: john,
			CreatedAt: event.CreatedAt,
		}}
		require.NoError(t, env.app.Store.EventUpsert(context.Background(), event))

		env.process(env.message(userJohn, "/my"))
		req := env.waitSendMessage(userJohn.ID, "Joined event")
		require.Contains(t, req.InlineKeyboardRaw(), "signup-my_joined_event-leader")
		require.NotContains(t, req.InlineKeyboardRaw(), "evt_set|my_joined_event")
	})

	t.Run("manual username participant", func(t *testing.T) {
		env := newEnv(t)
		event := testEvent("my_manual_username", "Manual username event", userJane)
		event.Couples = []models.Couple{{
			Dancers: []models.Dancer{
				{FullName: "Manual Partner @JohnUser", Role: models.RoleLeader},
				{FullName: "Test Partner", Role: models.RoleFollower},
			},
			CreatedBy: models.NewProfile(*userJane),
			CreatedAt: event.CreatedAt,
		}}
		require.NoError(t, env.app.Store.EventUpsert(context.Background(), event))

		john := models.NewProfile(*userJohn)
		john.Username = "johnuser"
		user, err := env.app.Services.User.Get(context.Background(), john)
		require.NoError(t, err)
		user.Profile.Username = "johnuser"
		require.NoError(t, env.app.Services.User.UpdateSession(context.Background(), user))

		env.process(env.message(&tele.User{ID: userJohn.ID, FirstName: "John", Username: "johnuser"}, "/my"))
		req := env.waitSendMessage(userJohn.ID, "Manual username event")
		require.NotContains(t, req.InlineKeyboardRaw(), "evt_set|my_manual_username")
	})

	t.Run("pagination", func(t *testing.T) {
		env := newEnv(t)
		require.NoError(t, env.app.Store.EventUpsert(context.Background(), testEvent("my_page_first", "Page first", userJohn)))
		require.NoError(t, env.app.Store.EventUpsert(context.Background(), testEvent("my_page_second", "Page second", userJohn)))

		env.process(env.message(userJohn, "/my"))
		req := env.waitSendMessage(userJohn.ID, "Page")
		require.Contains(t, req.InlineKeyboardRaw(), "my_turn_page|1")

		env.process(env.callback(tele.Callback{
			Sender:  userJohn,
			Message: &tele.Message{ID: 1, Chat: privateChat(userJohn)},
			Data:    "\fmy_turn_page|1|rand-token",
		}))
		edit := env.tg.Wait("editMessageText")
		require.Contains(t, edit.String("text"), "Page")
		require.Contains(t, edit.InlineKeyboardRaw(), "my_turn_page|0")
		env.tg.Wait("answerCallbackQuery")
	})
}

func TestEventSettingsMissingEvent(t *testing.T) {
	cases := []struct {
		name       string
		unique     string
		args       string
		settings   models.EventSettings
		markupOnly bool
	}{
		{"settings scene", views.BtnEventSettings.Unique, "%s|0|rand",
			models.EventSettings{Limit: 1}, false},
		{"auto pairing", views.BtnEventSettingsAutoPair.Unique, "%s|0|rand",
			models.EventSettings{Limit: 1, AutoPairing: true}, false},
		{"close registration", views.BtnEventSettingsClose.Unique, "%s|0|rand",
			models.EventSettings{Limit: 1, Closed: true}, false},
		{"limit scene", views.BtnEventSettingsLimit.Unique, "%s|0|0|rand",
			models.EventSettings{Limit: 1}, true},
		{"limit number", views.BtnEventSettingsLimitNum.Unique, "%s|2|0",
			models.EventSettings{Limit: 2}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			env := newEnv(t)
			event := testLimitEvent("settings_existing", userJohn, 1)
			require.NoError(t, env.app.Store.EventUpsert(env.ctx, event))
			before := env.eventGet(event.ID)
			profile := models.NewProfile(*userJohn)
			user, err := env.app.Services.User.Get(env.ctx, profile)
			require.NoError(t, err)
			user.Session = models.Session{
				Action:  models.SessionSignup,
				EventID: event.ID,
				Role:    models.RoleLeader,
			}
			require.NoError(t, env.app.Services.User.UpdateSession(env.ctx, user))

			callback := tele.Callback{
				Sender:  userJohn,
				Message: &tele.Message{ID: 100, Chat: privateChat(userJohn)},
				Data:    "\f" + tc.unique + "|" + fmt.Sprintf(tc.args, "missing_event"),
			}
			require.NotPanics(t, func() {
				env.process(env.callback(callback))
			})
			answer := env.tg.Wait("answerCallbackQuery")
			require.Equal(t, locale.ErrSomethingWrong, answer.String("text"))
			require.True(t, answer.Bool("show_alert"))
			user, err = env.app.Services.User.Get(env.ctx, profile)
			require.NoError(t, err)
			require.Equal(t, models.Session{}, user.Session)
			missing, err := env.app.Services.Event.Get(env.ctx, "missing_event")
			require.NoError(t, err)
			require.Nil(t, missing)
			require.Equal(t, before, env.eventGet(event.ID))
			env.tg.AssertNoUnexpected()

			callback.Data = "\f" + tc.unique + "|" + fmt.Sprintf(tc.args, event.ID)
			env.process(env.callback(callback))
			answer = env.tg.Wait("answerCallbackQuery")
			require.Empty(t, answer.String("text"))
			if tc.markupOnly {
				markup := env.tg.Wait("editMessageReplyMarkup")
				require.Contains(t, markup.InlineKeyboardRaw(), "evt_set_lim_num|"+event.ID+"|2|0")
			} else {
				edit := env.tg.WaitFor("editMessageText", func(req teletest.Request) bool {
					return req.ChatIDInt() == userJohn.ID
				})
				require.Contains(t, edit.String("text"), locale.EventSettingsCaption)
				if tc.unique != views.BtnEventSettings.Unique {
					env.tg.WaitFor("editMessageText", func(req teletest.Request) bool {
						return req.String("inline_message_id") == event.Post.InlineMessageID
					})
				}
			}
			require.Equal(t, tc.settings, env.eventGet(event.ID).Settings)
		})
	}
}

func TestEventSettingsEscapesCaption(t *testing.T) {
	env := newEnv(t)
	event := testEvent("settings_escape_caption", "Settings <tag> & owner", userJohn)
	require.NoError(t, env.app.Store.EventUpsert(context.Background(), event))

	env.process(env.callback(tele.Callback{
		Sender:  userJohn,
		Message: &tele.Message{ID: 100, Chat: privateChat(userJohn)},
		Data:    "\fevt_set|" + event.ID + "|0|0|rand",
	}))

	env.tg.Wait("answerCallbackQuery")
	edit := env.tg.Wait("editMessageText")
	require.Contains(t, edit.String("text"), "Settings &lt;tag&gt; &amp; owner")
	require.NotContains(t, edit.String("text"), "Settings <tag> & owner")
}

func TestUserSettingsDefaultEventLimit(t *testing.T) {
	t.Run("owner can set default event limit", func(t *testing.T) {
		env := newEnv(t)

		env.process(env.message(userJohn, "/settings"))
		req := env.waitSendMessage(userJohn.ID, locale.UserSettingsDescription)
		require.Contains(t, req.String("text"), locale.EventSettingsLimitNone)
		require.Contains(t, req.InlineKeyboardRaw(), "usr_set_lim")

		env.process(env.callback(tele.Callback{
			Sender:  userJohn,
			Message: &tele.Message{ID: 100, Chat: privateChat(userJohn)},
			Data:    "\fusr_set_lim|0|rand",
		}))
		env.tg.Wait("answerCallbackQuery")
		markup := env.tg.Wait("editMessageReplyMarkup")
		require.Contains(t, markup.InlineKeyboardRaw(), "usr_set_lim_num|10")

		env.process(env.callback(tele.Callback{
			Sender:  userJohn,
			Message: &tele.Message{ID: 100, Chat: privateChat(userJohn)},
			Data:    "\fusr_set_lim|1|rand",
		}))
		env.tg.Wait("answerCallbackQuery")
		markup = env.tg.Wait("editMessageReplyMarkup")
		require.Contains(t, markup.InlineKeyboardRaw(), "usr_set_lim_num|12")

		env.process(env.callback(tele.Callback{
			Sender:  userJohn,
			Message: &tele.Message{ID: 100, Chat: privateChat(userJohn)},
			Data:    "\fusr_set_lim_num|12|rand",
		}))
		env.tg.Wait("answerCallbackQuery")
		edit := env.tg.Wait("editMessageText")
		require.Contains(t, edit.String("text"), "Приходят первые 12 пар")

		user, err := env.app.Services.User.Get(context.Background(), models.NewProfile(*userJohn))
		require.NoError(t, err)
		require.Equal(t, 12, user.Settings.Event.Limit)
	})

	t.Run("default limit is applied and can be overridden", func(t *testing.T) {
		env := newEnv(t)
		user, err := env.app.Services.User.Get(context.Background(), models.NewProfile(*userJohn))
		require.NoError(t, err)
		user.Settings.Event.Limit = 12
		require.NoError(t, env.app.Services.User.UpdateSettings(context.Background(), user))

		env.process(env.inlineQuery(tele.Query{Sender: userJohn, Text: "Default limited event", ChatType: "supergroup"}))
		req := env.tg.Wait("answerInlineQuery")
		defaultID := req.JSON.Get("results.0.id").String()
		require.Equal(t, "Лимит 12 пар", req.JSON.Get("results.0.description").String())
		require.Equal(t, 12, env.eventGet(defaultID).Settings.Limit)

		env.process(env.inlineQuery(tele.Query{Sender: userJohn, Text: "Override limited event /2", ChatType: "supergroup"}))
		req = env.tg.Wait("answerInlineQuery")
		overrideID := req.JSON.Get("results.0.id").String()
		require.Equal(t, "Лимит 2 пары", req.JSON.Get("results.0.description").String())
		require.Equal(t, 2, env.eventGet(overrideID).Settings.Limit)
	})

	t.Run("owner can reset default event limit", func(t *testing.T) {
		env := newEnv(t)
		user, err := env.app.Services.User.Get(context.Background(), models.NewProfile(*userJohn))
		require.NoError(t, err)
		user.Settings.Event.Limit = 12
		require.NoError(t, env.app.Services.User.UpdateSettings(context.Background(), user))

		env.process(env.callback(tele.Callback{
			Sender:  userJohn,
			Message: &tele.Message{ID: 100, Chat: privateChat(userJohn)},
			Data:    "\fusr_set_lim_num|0|rand",
		}))
		env.tg.Wait("answerCallbackQuery")
		edit := env.tg.Wait("editMessageText")
		require.Contains(t, edit.String("text"), locale.EventSettingsLimitNone)

		user, err = env.app.Services.User.Get(context.Background(), models.NewProfile(*userJohn))
		require.NoError(t, err)
		require.Equal(t, 0, user.Settings.Event.Limit)
	})
}

func TestUserSettingsDefaultAutoPairing(t *testing.T) {
	env := newEnv(t)

	env.process(env.callback(tele.Callback{
		Sender:  userJohn,
		Message: &tele.Message{ID: 100, Chat: privateChat(userJohn)},
		Data:    "\fusr_set_auto_pair|rand",
	}))
	env.tg.Wait("answerCallbackQuery")
	edit := env.tg.Wait("editMessageText")
	require.Contains(t, edit.String("text"), locale.EventSettingsAutoPair[true])

	user, err := env.app.Services.User.Get(context.Background(), models.NewProfile(*userJohn))
	require.NoError(t, err)
	require.True(t, user.Settings.Event.AutoPairing)

	env.process(env.inlineQuery(tele.Query{Sender: userJohn, Text: "Auto pair default event", ChatType: "supergroup"}))
	req := env.tg.Wait("answerInlineQuery")
	eventID := req.JSON.Get("results.0.id").String()
	require.True(t, env.eventGet(eventID).Settings.AutoPairing)
}

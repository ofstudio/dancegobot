package tests

import (
	"context"
	"net/http"
	"strings"

	"github.com/h2non/gock"
	tele "gopkg.in/telebot.v4"

	"github.com/ofstudio/dancegobot/internal/locale"
	"github.com/ofstudio/dancegobot/internal/models"
	"github.com/ofstudio/dancegobot/pkg/telegock"
)

func (suite *AppTestSuite) TestUserSettingsDefaultEventLimit() {
	suite.Run("owner can set default event limit", func() {
		gock.New(telegock.SendMessage).
			Reply(200).
			Filter(func(res *http.Response) bool {
				body := suite.Decode(res.Request.Body)
				markup := telegramReplyMarkup(body)
				text := body.Get("text").String()
				keyboard := markup.Get("inline_keyboard").Raw
				return body.Get("chat_id").Int() == userJohn.ID &&
					strings.Contains(text, locale.UserSettingsDescription) &&
					strings.Contains(text, locale.EventSettingsLimitNone) &&
					strings.Contains(keyboard, "usr_set_lim")
			}).
			JSON(telegock.Result(tele.Message{ID: 100}))

		gock.New(telegock.GetUpdates).
			Reply(200).
			JSON(telegock.Updates().Message(tele.Message{
				Sender: userJohn,
				Chat:   privateChat(userJohn),
				Text:   "/settings",
			}))

		suite.NoPending()
		suite.NoUnmatched()

		gock.New(telegock.EditMessageMarkup).
			Reply(200).
			Filter(func(res *http.Response) bool {
				body := suite.Decode(res.Request.Body)
				markup := telegramReplyMarkup(body)
				keyboard := markup.Get("inline_keyboard").Raw
				return body.Get("chat_id").Int() == userJohn.ID &&
					strings.Contains(keyboard, locale.BtnEventSettingsLimitNone) &&
					strings.Contains(keyboard, "usr_set_lim_num|10") &&
					strings.Contains(keyboard, "usr_set_lim|1")
			}).
			JSON(telegock.Result(tele.Message{}))

		gock.New(telegock.AnswerCallbackQuery).
			Reply(200).
			JSON(telegock.Result(true))

		gock.New(telegock.GetUpdates).
			Reply(200).
			JSON(telegock.Updates().CallbackQuery(tele.Callback{
				Sender:  userJohn,
				Message: &tele.Message{ID: 100, Chat: privateChat(userJohn)},
				Data:    "\fusr_set_lim|0|rand",
			}))

		suite.NoPending()
		suite.NoUnmatched()

		gock.New(telegock.EditMessageMarkup).
			Reply(200).
			Filter(func(res *http.Response) bool {
				body := suite.Decode(res.Request.Body)
				markup := telegramReplyMarkup(body)
				keyboard := markup.Get("inline_keyboard").Raw
				return body.Get("chat_id").Int() == userJohn.ID &&
					strings.Contains(keyboard, "usr_set_lim_num|12") &&
					strings.Contains(keyboard, "usr_set_lim_num|20") &&
					strings.Contains(keyboard, "usr_set_lim|0")
			}).
			JSON(telegock.Result(tele.Message{}))

		gock.New(telegock.AnswerCallbackQuery).
			Reply(200).
			JSON(telegock.Result(true))

		gock.New(telegock.GetUpdates).
			Reply(200).
			JSON(telegock.Updates().CallbackQuery(tele.Callback{
				Sender:  userJohn,
				Message: &tele.Message{ID: 100, Chat: privateChat(userJohn)},
				Data:    "\fusr_set_lim|1|rand",
			}))

		suite.NoPending()
		suite.NoUnmatched()

		gock.New(telegock.EditMessageText).
			Reply(200).
			Filter(func(res *http.Response) bool {
				body := suite.Decode(res.Request.Body)
				text := body.Get("text").String()
				markup := telegramReplyMarkup(body)
				return body.Get("chat_id").Int() == userJohn.ID &&
					strings.Contains(text, "Приходят первые 12 пар") &&
					strings.Contains(markup.Get("inline_keyboard").Raw, "usr_set_lim")
			}).
			JSON(telegock.Result(tele.Message{}))

		gock.New(telegock.AnswerCallbackQuery).
			Reply(200).
			JSON(telegock.Result(true))

		gock.New(telegock.GetUpdates).
			Reply(200).
			JSON(telegock.Updates().CallbackQuery(tele.Callback{
				Sender:  userJohn,
				Message: &tele.Message{ID: 100, Chat: privateChat(userJohn)},
				Data:    "\fusr_set_lim_num|12|rand",
			}))

		suite.NoPending()
		suite.NoUnmatched()

		user, err := suite.app.UserService.Get(context.Background(), models.NewProfile(*userJohn))
		suite.Require().NoError(err)
		suite.Equal(12, user.Settings.Event.Limit)
	})

	suite.Run("default limit is applied to new events and can be overridden", func() {
		user, err := suite.app.UserService.Get(context.Background(), models.NewProfile(*userJohn))
		suite.Require().NoError(err)
		user.Settings.Event.Limit = 12
		suite.Require().NoError(suite.app.UserService.UpdateSettings(context.Background(), user))

		var defaultEventID string
		gock.New(telegock.AnswerInlineQuery).
			Reply(200).
			Filter(func(res *http.Response) bool {
				body := suite.Decode(res.Request.Body)
				result := body.Get("results.0")
				defaultEventID = result.Get("id").String()
				return result.Get("description").String() == "Лимит 12 пар"
			}).
			JSON(telegock.Result(true))

		gock.New(telegock.GetUpdates).
			Reply(200).
			JSON(telegock.Updates().InlineQuery(tele.Query{
				Sender:   userJohn,
				Text:     "Default limited event",
				ChatType: "supergroup",
			}))

		suite.NoPending()
		suite.NoUnmatched()

		defaultEvent, err := suite.app.EventService.Get(context.Background(), defaultEventID)
		suite.Require().NoError(err)
		suite.Require().NotNil(defaultEvent)
		suite.Equal(12, defaultEvent.Settings.Limit)

		var overrideEventID string
		gock.New(telegock.AnswerInlineQuery).
			Reply(200).
			Filter(func(res *http.Response) bool {
				body := suite.Decode(res.Request.Body)
				result := body.Get("results.0")
				overrideEventID = result.Get("id").String()
				return result.Get("title").String() == "Override limited event" &&
					result.Get("description").String() == "Лимит 2 пары"
			}).
			JSON(telegock.Result(true))

		gock.New(telegock.GetUpdates).
			Reply(200).
			JSON(telegock.Updates().InlineQuery(tele.Query{
				Sender:   userJohn,
				Text:     "Override limited event /2",
				ChatType: "supergroup",
			}))

		suite.NoPending()
		suite.NoUnmatched()

		overrideEvent, err := suite.app.EventService.Get(context.Background(), overrideEventID)
		suite.Require().NoError(err)
		suite.Require().NotNil(overrideEvent)
		suite.Equal(2, overrideEvent.Settings.Limit)
	})

	suite.Run("owner can reset default event limit", func() {
		user, err := suite.app.UserService.Get(context.Background(), models.NewProfile(*userJohn))
		suite.Require().NoError(err)
		user.Settings.Event.Limit = 12
		suite.Require().NoError(suite.app.UserService.UpdateSettings(context.Background(), user))

		gock.New(telegock.EditMessageText).
			Reply(200).
			Filter(func(res *http.Response) bool {
				body := suite.Decode(res.Request.Body)
				return body.Get("chat_id").Int() == userJohn.ID &&
					strings.Contains(body.Get("text").String(), locale.EventSettingsLimitNone)
			}).
			JSON(telegock.Result(tele.Message{}))

		gock.New(telegock.AnswerCallbackQuery).
			Reply(200).
			JSON(telegock.Result(true))

		gock.New(telegock.GetUpdates).
			Reply(200).
			JSON(telegock.Updates().CallbackQuery(tele.Callback{
				Sender:  userJohn,
				Message: &tele.Message{ID: 100, Chat: privateChat(userJohn)},
				Data:    "\fusr_set_lim_num|0|rand",
			}))

		suite.NoPending()
		suite.NoUnmatched()

		user, err = suite.app.UserService.Get(context.Background(), models.NewProfile(*userJohn))
		suite.Require().NoError(err)
		suite.Equal(0, user.Settings.Event.Limit)
	})
}

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

func (suite *AppTestSuite) TestEventAutoPairSignup() {
	suite.Run("single dancers are paired automatically", func() {
		event := testMyEvent("auto_pair_signup", "Auto pair signup", userJohn)
		event.Settings.AutoPairing = true
		suite.Require().NoError(suite.app.Store.EventUpsert(context.Background(), event))

		gock.New(telegock.SendMessage).
			Reply(200).
			Filter(func(res *http.Response) bool {
				body := suite.Decode(res.Request.Body)
				markup := telegramReplyMarkup(body)
				return body.Get("chat_id").Int() == userJohn.ID &&
					body.Get("text").String() == locale.SignupNotRegistered &&
					strings.Contains(markup.Get("keyboard").Raw, locale.BtnAsSingle[models.RoleLeader])
			}).
			JSON(telegock.Result(tele.Message{}))

		gock.New(telegock.GetUpdates).
			Reply(200).
			JSON(telegock.Updates().Message(tele.Message{
				Sender: userJohn,
				Chat:   privateChat(userJohn),
				Text:   "/start rand-signup-" + event.ID + "-leader",
			}))

		suite.NoPending()
		suite.NoUnmatched()

		gock.New(telegock.EditMessageText).
			Reply(200).
			Filter(func(res *http.Response) bool {
				body := suite.Decode(res.Request.Body)
				if body.Get("inline_message_id").String() != event.Post.InlineMessageID {
					return false
				}
				text := body.Get("text").String()
				return strings.Contains(text, locale.PostSingles[models.RoleLeader]) &&
					strings.Contains(text, "John")
			}).
			JSON(telegock.Result(true))

		gock.New(telegock.SendMessage).
			Reply(200).
			Filter(func(res *http.Response) bool {
				body := suite.Decode(res.Request.Body)
				return body.Get("chat_id").Int() == userJohn.ID &&
					strings.Contains(body.Get("text").String(), "Добавил тебя в список ищущих пару")
			}).
			JSON(telegock.Result(tele.Message{}))

		gock.New(telegock.GetUpdates).
			Reply(200).
			JSON(telegock.Updates().Message(tele.Message{
				Sender: userJohn,
				Chat:   privateChat(userJohn),
				Text:   locale.BtnAsSingle[models.RoleLeader],
			}))

		suite.NoPending()
		suite.NoUnmatched()

		gock.New(telegock.SendMessage).
			Reply(200).
			Filter(func(res *http.Response) bool {
				body := suite.Decode(res.Request.Body)
				markup := telegramReplyMarkup(body)
				return body.Get("chat_id").Int() == userJane.ID &&
					body.Get("text").String() == locale.SignupNotRegistered &&
					strings.Contains(markup.Get("keyboard").Raw, locale.BtnAsSingle[models.RoleFollower])
			}).
			JSON(telegock.Result(tele.Message{}))

		gock.New(telegock.GetUpdates).
			Reply(200).
			JSON(telegock.Updates().Message(tele.Message{
				Sender: userJane,
				Chat:   privateChat(userJane),
				Text:   "/start rand-signup-" + event.ID + "-follower",
			}))

		suite.NoPending()
		suite.NoUnmatched()

		gock.New(telegock.EditMessageText).
			Reply(200).
			Filter(func(res *http.Response) bool {
				body := suite.Decode(res.Request.Body)
				if body.Get("inline_message_id").String() != event.Post.InlineMessageID {
					return false
				}
				text := body.Get("text").String()
				return strings.Contains(text, locale.PostCouples) &&
					strings.Contains(text, "John") &&
					strings.Contains(text, "Jane Doe") &&
					!strings.Contains(text, locale.PostSingles[models.RoleLeader])
			}).
			JSON(telegock.Result(true))

		gock.New(telegock.SendMessage).
			Reply(200).
			Filter(func(res *http.Response) bool {
				body := suite.Decode(res.Request.Body)
				text := body.Get("text").String()
				return body.Get("chat_id").Int() == userJane.ID &&
					strings.Contains(text, "Вы зарегистрировались в паре") &&
					strings.Contains(text, "John")
			}).
			JSON(telegock.Result(tele.Message{}))

		gock.New(telegock.SendMessage).
			Reply(200).
			Filter(func(res *http.Response) bool {
				body := suite.Decode(res.Request.Body)
				text := body.Get("text").String()
				return body.Get("chat_id").Int() == userJohn.ID &&
					strings.Contains(text, "Я подобрал тебе в пару") &&
					strings.Contains(text, "Jane Doe")
			}).
			JSON(telegock.Result(tele.Message{}))

		gock.New(telegock.GetUpdates).
			Reply(200).
			JSON(telegock.Updates().Message(tele.Message{
				Sender: userJane,
				Chat:   privateChat(userJane),
				Text:   locale.BtnAsSingle[models.RoleFollower],
			}))

		suite.NoPending()
		suite.NoUnmatched()

		updated, err := suite.app.EventService.Get(context.Background(), event.ID)
		suite.Require().NoError(err)
		suite.Require().NotNil(updated)
		suite.Require().Len(updated.Couples, 1)
		suite.True(updated.Couples[0].AutoPair)
		suite.Empty(updated.Singles)

		johnReg, err := suite.app.EventService.RegistrationGet(context.Background(),
			event.ID, models.NewProfile(*userJohn), models.RoleLeader)
		suite.Require().NoError(err)
		suite.Equal(models.StatusInCouple, johnReg.Status)

		janeReg, err := suite.app.EventService.RegistrationGet(context.Background(),
			event.ID, models.NewProfile(*userJane), models.RoleFollower)
		suite.Require().NoError(err)
		suite.Equal(models.StatusInCouple, janeReg.Status)
	})
}

func (suite *AppTestSuite) TestEventSettingsAccessControl() {
	suite.Run("non-owner cannot update event settings", func() {
		event := testMyEvent("settings_access_control", "Settings access", userJohn)
		suite.Require().NoError(suite.app.Store.EventUpsert(context.Background(), event))

		gock.New(telegock.AnswerCallbackQuery).
			Reply(200).
			Filter(func(res *http.Response) bool {
				body := suite.Decode(res.Request.Body)
				return body.Get("text").String() == locale.ErrSomethingWrong &&
					body.Get("show_alert").Bool()
			}).
			JSON(telegock.Result(true))

		gock.New(telegock.GetUpdates).
			Reply(200).
			JSON(telegock.Updates().CallbackQuery(tele.Callback{
				Sender:  userJane,
				Message: &tele.Message{ID: 100, Chat: privateChat(userJane)},
				Data:    "\fevt_set_auto_pair|" + event.ID + "|0|rand",
			}))

		suite.NoPending()
		suite.NoUnmatched()

		updated, err := suite.app.EventService.Get(context.Background(), event.ID)
		suite.Require().NoError(err)
		suite.Require().NotNil(updated)
		suite.False(updated.Settings.AutoPairing)
	})
}

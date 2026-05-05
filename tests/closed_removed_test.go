package tests

import (
	"context"
	"net/http"
	"strings"

	"github.com/h2non/gock"
	tele "gopkg.in/telebot.v4"

	"github.com/ofstudio/dancegobot/internal/locale"
	"github.com/ofstudio/dancegobot/pkg/telegock"
)

func (suite *AppTestSuite) TestEventClosedRegistration() {
	suite.Run("owner can close event registration", func() {
		event := testMyEvent("closed_registration", "Closed registration", userJohn)
		suite.Require().NoError(suite.app.Store.EventUpsert(context.Background(), event))

		gock.New(telegock.EditMessageText).
			Reply(200).
			Filter(func(res *http.Response) bool {
				body := suite.Decode(res.Request.Body)
				if body.Get("inline_message_id").String() != event.Post.InlineMessageID {
					return false
				}
				text := body.Get("text").String()
				markup := telegramReplyMarkup(body)
				return strings.HasPrefix(text, locale.IconPostClosed) &&
					strings.Contains(text, event.Caption) &&
					strings.Contains(markup.Get("inline_keyboard").Raw, "post_closed")
			}).
			JSON(telegock.Result(true))

		gock.New(telegock.EditMessageText).
			Reply(200).
			Filter(func(res *http.Response) bool {
				body := suite.Decode(res.Request.Body)
				if body.Get("chat_id").Int() != userJohn.ID {
					return false
				}
				text := body.Get("text").String()
				markup := telegramReplyMarkup(body)
				return strings.Contains(text, locale.EventSettingsClosed[true]) &&
					strings.Contains(markup.Get("inline_keyboard").Raw, locale.BtnEventSettingsClosed[true])
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
				Data:    "\fevt_set_close|" + event.ID + "|0|rand",
			}))

		suite.NoPending()
		suite.NoUnmatched()

		updated, err := suite.app.EventService.Get(context.Background(), event.ID)
		suite.Require().NoError(err)
		suite.Require().NotNil(updated)
		suite.True(updated.Settings.Closed)

		gock.New(telegock.SendMessage).
			Reply(200).
			Filter(func(res *http.Response) bool {
				body := suite.Decode(res.Request.Body)
				return body.Get("chat_id").Int() == userJane.ID &&
					body.Get("text").String() == locale.ResultEventClosed
			}).
			JSON(telegock.Result(tele.Message{}))

		gock.New(telegock.GetUpdates).
			Reply(200).
			JSON(telegock.Updates().Message(tele.Message{
				Sender: userJane,
				Chat:   privateChat(userJane),
				Text:   "/start rand-signup-" + event.ID + "-leader",
			}))

		suite.NoPending()
		suite.NoUnmatched()
	})

	suite.Run("closed post button returns closed notice", func() {
		gock.New(telegock.AnswerCallbackQuery).
			Reply(200).
			Filter(func(res *http.Response) bool {
				body := suite.Decode(res.Request.Body)
				return body.Get("text").String() == locale.ResultEventClosed
			}).
			JSON(telegock.Result(true))

		gock.New(telegock.GetUpdates).
			Reply(200).
			JSON(telegock.Updates().CallbackQuery(tele.Callback{
				Sender:    userJane,
				MessageID: "inline_closed_registration",
				Data:      "\fpost_closed|rand",
			}))

		suite.NoPending()
		suite.NoUnmatched()
	})
}

func (suite *AppTestSuite) TestEventRemovedAfterRenderFailures() {
	suite.Run("event is marked removed after repeated invalid inline message renders", func() {
		eventID := suite.eventDraftCreate(queryA)

		for i := 0; i < 3; i++ {
			gock.New(telegock.EditMessageText).
				Reply(200).
				JSON(map[string]any{
					"ok":          false,
					"error_code":  400,
					"description": "Bad Request: MESSAGE_ID_INVALID",
				})

			gock.New(telegock.GetUpdates).
				Reply(200).
				JSON(telegock.Updates().InlineResult(tele.InlineResult{
					Sender:    userJohn,
					ResultID:  eventID,
					Query:     queryA.Text,
					MessageID: "missing-inline-message",
				}))

			suite.NoPending()
			suite.NoUnmatched()
		}

		event, err := suite.app.EventService.Get(context.Background(), eventID)
		suite.Require().NoError(err)
		suite.Require().NotNil(event)
		suite.Equal(3, event.RenderFails)
		suite.True(event.Removed)
	})
}

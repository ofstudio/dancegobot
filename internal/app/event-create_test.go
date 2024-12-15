package app

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

func (suite *AppTestSuite) TestEventDraft() {
	suite.Run("empty inline query", func() {
		gock.New(telegock.GetUpdates).
			Reply(200).
			JSON(telegock.Updates().InlineQuery(tele.Query{
				Sender:   userJohn,
				Text:     "",
				ChatType: "supergroup",
			}))

		gock.New(telegock.AnswerInlineQuery).
			Reply(200).
			Filter(func(res *http.Response) bool {
				body := suite.Decode(res.Request.Body)
				suite.Len(body.Get("results").Array(), 1)
				result := body.Get("results").Array()[0]
				suite.Equal("article", result.Get("type").String())
				suite.Equal(locale.QueryTextEmpty, result.Get("title").String())
				suite.Equal(locale.QueryDescriptionEmpty, result.Get("description").String())
				suite.Equal(locale.QueryTextEmpty, result.Get("message_text").String())
				suite.False(result.Get("reply_markup.inline_keyboard").Exists())
				return true
			}).
			JSON(telegock.Result(true))

		suite.NoPending()
		suite.NoUnmatched()
	})

	suite.Run("non-empty inline query", func() {
		gock.New(telegock.GetUpdates).
			Reply(200).
			JSON(telegock.Updates().InlineQuery(tele.Query{
				Sender:   userJohn,
				Text:     "Test text",
				ChatType: "supergroup",
			}))

		var eventID string
		gock.New(telegock.AnswerInlineQuery).
			Reply(200).
			Filter(func(res *http.Response) bool {
				body := suite.Decode(res.Request.Body)
				suite.Len(body.Get("results").Array(), 1)
				result := body.Get("results").Array()[0]
				suite.Equal("article", result.Get("type").String())
				suite.Equal("Test text", result.Get("title").String())
				suite.Equal(locale.QueryDescription, result.Get("description").String())
				suite.Equal("Test text", result.Get("input_message_content.message_text").String())
				suite.Equal("HTML", result.Get("input_message_content.parse_mode").String())
				suite.True(result.Get("input_message_content.link_preview_options.is_disabled").Bool())
				suite.Len(result.Get("reply_markup.inline_keyboard.0").Array(), 2)
				eventID = result.Get("id").String()
				kbd := result.Get("reply_markup.inline_keyboard.0")
				suite.Equal(locale.RoleIcon[models.RoleLeader], kbd.Get("0.text").String())
				suite.Contains(kbd.Get("0.callback_data").String(), "\fsignup|"+eventID+"|leader")
				suite.Equal(locale.RoleIcon[models.RoleFollower], kbd.Get("1.text").String())
				suite.Contains(kbd.Get("1.callback_data").String(), "\fsignup|"+eventID+"|follower")
				return true
			}).JSON(telegock.Result(true))

		suite.NoPending()
		suite.NoUnmatched()
		event, err := suite.app.srv.Event.Get(context.Background(), eventID)
		suite.NoError(err)
		suite.NotNil(event)
		suite.Equal("Test text", event.Caption)
	})

	suite.Run("quite long inline query", func() {
		gock.New(telegock.GetUpdates).
			Reply(200).
			JSON(telegock.Updates().InlineQuery(tele.Query{
				Sender:   userJohn,
				Text:     strings.Repeat("A", 250),
				ChatType: "supergroup",
			}))

		gock.New(telegock.AnswerInlineQuery).
			Reply(200).
			Filter(func(res *http.Response) bool {
				body := suite.Decode(res.Request.Body)
				suite.Len(body.Get("results").Array(), 1)
				result := body.Get("results").Array()[0]
				suite.Equal("Осталось 5 символов", result.Get("description").String())
				return true
			}).JSON(telegock.Result(true))

		suite.NoPending()
		suite.NoUnmatched()
	})

	suite.Run("too long inline query", func() {
		gock.New(telegock.GetUpdates).
			Reply(200).
			JSON(telegock.Updates().InlineQuery(tele.Query{
				Sender:   userJohn,
				Text:     strings.Repeat("A", 300),
				ChatType: "supergroup",
			}))

		gock.New(telegock.AnswerInlineQuery).
			Reply(200).
			Filter(func(res *http.Response) bool {
				body := suite.Decode(res.Request.Body)
				suite.Len(body.Get("results").Array(), 1)
				result := body.Get("results").Array()[0]
				suite.Equal(locale.QueryOverflow, result.Get("description").String())
				return true
			}).JSON(telegock.Result(true))

		suite.NoPending()
		suite.NoUnmatched()
	})
}

func (suite *AppTestSuite) eventDraftCreate(query tele.Query) string {
	var eventID string

	gock.New(telegock.GetUpdates).
		Reply(200).
		JSON(telegock.Updates().InlineQuery(query))

	gock.New(telegock.AnswerInlineQuery).
		Reply(200).
		Filter(func(res *http.Response) bool {
			body := suite.Decode(res.Request.Body)
			eventID = body.Get("results.0.id").String()
			suite.Regexp(`^[a-zA-Z0-9]{12}$`, eventID)
			return true
		}).JSON(telegock.Result(true))

	suite.NoPending()
	return eventID
}

func (suite *AppTestSuite) TestEventPostAdd() {

	suite.Run("via ChosenInlineResult", func() {
		eventID := suite.eventDraftCreate(queryA)
		gock.New(telegock.GetUpdates).
			Reply(200).
			JSON(telegock.Updates().InlineResult(tele.InlineResult{
				Sender:    userJohn,
				ResultID:  eventID,
				Query:     queryA.Text,
				MessageID: "test-inline-message-ChosenInlineResult",
			}))

		// Should render the event post
		gock.New(telegock.EditMessageText).Reply(200).JSON(telegock.Result(true))

		suite.NoPending()
		suite.NoUnmatched()
		event, err := suite.app.srv.Event.Get(context.Background(), eventID)
		suite.NoError(err)
		suite.Require().NotNil(event)
		suite.Equal(eventID, event.ID)
		suite.Equal(queryA.Text, event.Caption)
		suite.Require().NotNil(event.Post)
		suite.Equal("test-inline-message-ChosenInlineResult", event.Post.InlineMessageID)
	})

	suite.Run("via CallbackQuery success", func() {
		eventID := suite.eventDraftCreate(queryB)
		gock.New(telegock.GetUpdates).
			Reply(200).
			JSON(telegock.Updates().CallbackQuery(tele.Callback{
				Sender:    userJohn,
				MessageID: "test-inline-message-CallbackQuery",
				Data:      "\fsignup|" + eventID + "|leader|rand-token",
			}))

		// Should reply with the signup URL
		gock.New(telegock.AnswerCallbackQuery).
			Reply(200).
			Filter(func(res *http.Response) bool {
				body := suite.Decode(res.Request.Body)
				suite.Regexp(rxUrlSignupLeader, body.Get("url").String())
				return true
			}).JSON(telegock.Result(true))

		// Should render the event post
		gock.New(telegock.EditMessageText).Reply(200).JSON(telegock.Result(true))

		suite.NoPending()
		suite.NoUnmatched()
		event, err := suite.app.srv.Event.Get(context.Background(), eventID)
		suite.NoError(err)
		suite.Require().NotNil(event)
		suite.Equal(eventID, event.ID)
		suite.Equal(queryB.Text, event.Caption)
		suite.Require().NotNil(event.Post)
		suite.Equal("test-inline-message-CallbackQuery", event.Post.InlineMessageID)
	})

	suite.Run("via CallbackQuery invalid", func() {
		eventID := suite.eventDraftCreate(queryB)
		gock.New(telegock.GetUpdates).
			Reply(200).
			JSON(telegock.Updates().CallbackQuery(tele.Callback{
				Sender:    userJohn,
				MessageID: "test-inline-message-CallbackQuery",
				Data:      "\fsignup|INVALID_PAYLOAD",
			}))

		// Should reply with the signup URL
		gock.New(telegock.AnswerCallbackQuery).
			Reply(200).
			Filter(func(res *http.Response) bool {
				body := suite.Decode(res.Request.Body)
				suite.Equal(locale.ErrSomethingWrong, body.Get("text").String())
				return true
			}).JSON(telegock.Result(true))

		suite.NoPending()
		suite.NoUnmatched()
		event, err := suite.app.srv.Event.Get(context.Background(), eventID)
		suite.Require().NoError(err)
		suite.Require().NotNil(event)
		suite.Equal(eventID, event.ID)
		suite.Require().Nil(event.Post)
	})
}

func (suite *AppTestSuite) TestPostChatAdd() {

	suite.Run("if before ChosenInlineResult", func() {
		eventID := suite.eventDraftCreate(queryA)
		gock.New(telegock.GetUpdates).
			Reply(200).
			JSON(telegock.Updates().Message(tele.Message{
				ID:     12345,
				Sender: userJohn,
				Chat:   chatSuperGroup,
				Text:   queryA.Text,
				Via:    botUser,
				ReplyMarkup: &tele.ReplyMarkup{
					InlineKeyboard: [][]tele.InlineButton{
						{
							{Text: locale.RoleIcon[models.RoleLeader], Data: "\fsignup|" + eventID + "|leader|rand-token"},
							{Text: locale.RoleIcon[models.RoleLeader], Data: "\fsignup|" + eventID + "|follower|rand-token"},
						},
					},
				}}))
		suite.NoPending()
		suite.NoUnmatched()
		event, err := suite.app.srv.Event.Get(context.Background(), eventID)
		suite.Require().NoError(err)
		suite.Require().NotNil(event)
		suite.Require().NotNil(event.Post)
		suite.Equal(models.NewChat(chatSuperGroup), *event.Post.Chat)
		suite.Equal(12345, event.Post.ChatMessageID)
	})

	suite.Run("if after ChosenInlineResult", func() {
		eventID := suite.eventDraftCreate(queryB)
		gock.New(telegock.GetUpdates).
			Reply(200).
			JSON(telegock.Updates().Message(tele.Message{
				ID:     67890,
				Sender: userJane,
				Chat:   chatSuperGroup,
				Text:   queryB.Text,
				Via:    botUser,
				ReplyMarkup: &tele.ReplyMarkup{
					InlineKeyboard: [][]tele.InlineButton{
						{
							{Text: locale.RoleIcon[models.RoleLeader], URL: "https://t.me/" + botUser.Username + "?start=rand_token-signup-" + eventID + "-leader"},
							{Text: locale.RoleIcon[models.RoleLeader], URL: "https://t.me/" + botUser.Username + "?start=rand_token-signup-" + eventID + "-follower"},
						},
					},
				}}))

		suite.NoPending()
		suite.NoUnmatched()
		event, err := suite.app.srv.Event.Get(context.Background(), eventID)
		suite.Require().NoError(err)
		suite.Require().NotNil(event)
		suite.Require().NotNil(event.Post)
		suite.Equal(models.NewChat(chatSuperGroup), *event.Post.Chat)
		suite.Equal(67890, event.Post.ChatMessageID)
	})
}

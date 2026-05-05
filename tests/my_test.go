package tests

import (
	"context"
	"net/http"
	"time"

	"github.com/h2non/gock"
	"github.com/tidwall/gjson"
	tele "gopkg.in/telebot.v4"

	"github.com/ofstudio/dancegobot/internal/locale"
	"github.com/ofstudio/dancegobot/internal/models"
	"github.com/ofstudio/dancegobot/pkg/telegock"
)

func (suite *AppTestSuite) TestMyCommandNoEvents() {
	suite.Run("no events", func() {
		gock.New(telegock.SendMessage).
			Reply(200).
			Filter(func(res *http.Response) bool {
				body := suite.Decode(res.Request.Body)
				markup := telegramReplyMarkup(body)
				suite.Equal(userJohn.ID, body.Get("chat_id").Int())
				suite.Equal(locale.MyNoEvents, body.Get("text").String())
				suite.Equal(locale.BtnTry, markup.Get("inline_keyboard.0.0.text").String())
				return true
			}).
			JSON(telegock.Result(tele.Message{}))

		gock.New(telegock.GetUpdates).
			Reply(200).
			JSON(telegock.Updates().Message(tele.Message{
				Sender: userJohn,
				Chat:   privateChat(userJohn),
				Text:   "/my",
			}))

		suite.NoPending()
		suite.NoUnmatched()
	})
}

func (suite *AppTestSuite) TestMyCommandOwnedEvent() {
	suite.Run("owned event", func() {
		event := testMyEvent("my_owned_event", "Owner event", userJohn)
		suite.Require().NoError(suite.app.Store.EventUpsert(context.Background(), event))

		gock.New(telegock.SendMessage).
			Reply(200).
			Filter(func(res *http.Response) bool {
				body := suite.Decode(res.Request.Body)
				markup := telegramReplyMarkup(body)
				suite.Contains(body.Get("text").String(), "Owner event")
				suite.Contains(markup.Get("inline_keyboard.0.0.url").String(), "signup-my_owned_event-leader")
				suite.Contains(markup.Get("inline_keyboard.0.1.url").String(), "signup-my_owned_event-follower")
				suite.Contains(markup.Get("inline_keyboard.0.2.callback_data").String(), "my_refresh|0")
				suite.Contains(markup.Get("inline_keyboard").Raw, "evt_set|my_owned_event|0")
				return true
			}).
			JSON(telegock.Result(tele.Message{}))

		gock.New(telegock.GetUpdates).
			Reply(200).
			JSON(telegock.Updates().Message(tele.Message{
				Sender: userJohn,
				Chat:   privateChat(userJohn),
				Text:   "/my",
			}))

		suite.NoPending()
		suite.NoUnmatched()
	})
}

func (suite *AppTestSuite) TestMyCommandJoinedEvent() {
	suite.Run("joined event", func() {
		event := testMyEvent("my_joined_event", "Joined event", userJane)
		john := models.NewProfile(*userJohn)
		event.Couples = []models.Couple{{
			Dancers: []models.Dancer{
				{Profile: &john, FullName: john.FullName(), Role: models.RoleLeader},
				{FullName: "Test Partner", Role: models.RoleFollower},
			},
			CreatedBy: john,
			CreatedAt: event.CreatedAt,
		}}
		suite.Require().NoError(suite.app.Store.EventUpsert(context.Background(), event))

		gock.New(telegock.SendMessage).
			Reply(200).
			Filter(func(res *http.Response) bool {
				body := suite.Decode(res.Request.Body)
				markup := telegramReplyMarkup(body)
				keyboard := markup.Get("inline_keyboard").Raw
				suite.Contains(body.Get("text").String(), "Joined event")
				suite.Contains(markup.Get("inline_keyboard.0.0.url").String(), "signup-my_joined_event-leader")
				suite.Contains(markup.Get("inline_keyboard.0.1.callback_data").String(), "my_refresh|0")
				suite.NotContains(keyboard, "evt_set|my_joined_event")
				return true
			}).
			JSON(telegock.Result(tele.Message{}))

		gock.New(telegock.GetUpdates).
			Reply(200).
			JSON(telegock.Updates().Message(tele.Message{
				Sender: userJohn,
				Chat:   privateChat(userJohn),
				Text:   "/my",
			}))

		suite.NoPending()
		suite.NoUnmatched()
	})
}

func (suite *AppTestSuite) TestMyCommandPagination() {
	suite.Run("pagination", func() {
		events := []*models.Event{
			testMyEvent("my_page_first", "Page first", userJohn),
			testMyEvent("my_page_second", "Page second", userJohn),
		}
		for _, event := range events {
			suite.Require().NoError(suite.app.Store.EventUpsert(context.Background(), event))
		}

		ids, err := suite.app.EventService.GetMy(context.Background(), &models.Profile{ID: userJohn.ID})
		suite.Require().NoError(err)
		suite.Require().Len(ids, 2)
		captions := map[string]string{
			"my_page_first":  "Page first",
			"my_page_second": "Page second",
		}

		gock.New(telegock.SendMessage).
			Reply(200).
			Filter(func(res *http.Response) bool {
				body := suite.Decode(res.Request.Body)
				markup := telegramReplyMarkup(body)
				suite.Contains(body.Get("text").String(), captions[ids[0]])
				suite.Contains(markup.Get("inline_keyboard").Raw, "my_turn_page|1")
				return true
			}).
			JSON(telegock.Result(tele.Message{ID: 1}))

		gock.New(telegock.GetUpdates).
			Reply(200).
			JSON(telegock.Updates().Message(tele.Message{
				Sender: userJohn,
				Chat:   privateChat(userJohn),
				Text:   "/my",
			}))

		suite.NoPending()
		suite.NoUnmatched()

		gock.New(telegock.EditMessageText).
			Reply(200).
			Filter(func(res *http.Response) bool {
				body := suite.Decode(res.Request.Body)
				markup := telegramReplyMarkup(body)
				suite.Contains(body.Get("text").String(), captions[ids[1]])
				suite.Contains(markup.Get("inline_keyboard").Raw, "my_turn_page|0")
				return true
			}).
			JSON(telegock.Result(tele.Message{}))

		gock.New(telegock.AnswerCallbackQuery).
			Reply(200).
			JSON(telegock.Result(true))

		gock.New(telegock.GetUpdates).
			Reply(200).
			JSON(telegock.Updates().CallbackQuery(tele.Callback{
				Sender:  userJohn,
				Message: &tele.Message{ID: 1, Chat: privateChat(userJohn)},
				Data:    "\fmy_turn_page|1|rand-token",
			}))

		suite.NoPending()
		suite.NoUnmatched()
	})
}

func privateChat(user *tele.User) *tele.Chat {
	return &tele.Chat{
		ID:        user.ID,
		Type:      tele.ChatPrivate,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Username:  user.Username,
	}
}

func testMyEvent(id string, caption string, owner *tele.User) *models.Event {
	profile := models.NewProfile(*owner)
	return &models.Event{
		ID:        id,
		Caption:   caption,
		Post:      &models.Post{InlineMessageID: "inline_" + id},
		Owner:     profile,
		CreatedAt: time.Date(2026, 5, 5, 12, 0, 0, 0, time.UTC),
	}
}

func telegramReplyMarkup(body gjson.Result) gjson.Result {
	markup := body.Get("reply_markup")
	if markup.IsObject() || markup.IsArray() {
		return markup
	}
	return gjson.Parse(markup.String())
}

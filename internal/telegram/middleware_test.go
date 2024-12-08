package telegram

import (
	"testing"

	"github.com/stretchr/testify/suite"
	tele "gopkg.in/telebot.v4"

	"github.com/ofstudio/dancegobot/internal/config"
)

func TestMiddleware(t *testing.T) {
	suite.Run(t, new(TestMiddlewareSuite))
}

type TestMiddlewareSuite struct {
	suite.Suite
}

func (suite *TestMiddlewareSuite) SetupSuite() {
	config.SetBotProfile(&tele.User{ID: 123, Username: "my_bot"})
}

func (suite *TestMiddlewareSuite) Test_isPost() {
	suite.Run("is post message", func() {
		m := NewMiddleware(config.Settings{}, nil)
		eventID, ok := m.isPost(&msgIsPost)
		suite.True(ok)
		suite.Equal("event123", eventID)
	})

	suite.Run("not via bot", func() {
		m := NewMiddleware(config.Settings{}, nil)
		_, ok := m.isPost(&msgNotViaBot)
		suite.False(ok)
	})

	suite.Run("via another bot", func() {
		m := NewMiddleware(config.Settings{}, nil)
		_, ok := m.isPost(&msgViaAnotherBot)
		suite.False(ok)
	})

	suite.Run("no reply markup", func() {
		m := NewMiddleware(config.Settings{}, nil)
		_, ok := m.isPost(&msgNoReplyMarkup)
		suite.False(ok)
	})

	suite.Run("no inline keyboard", func() {
		m := NewMiddleware(config.Settings{}, nil)
		_, ok := m.isPost(&msgNoInlineKeyboard)
		suite.False(ok)
	})

	suite.Run("no inline button", func() {
		m := NewMiddleware(config.Settings{}, nil)
		_, ok := m.isPost(&msgNoInlineButton)
		suite.False(ok)
	})

	suite.Run("no URL", func() {
		m := NewMiddleware(config.Settings{}, nil)
		_, ok := m.isPost(&msgNoURL)
		suite.False(ok)
	})

	suite.Run("unknown URL", func() {
		m := NewMiddleware(config.Settings{}, nil)
		_, ok := m.isPost(&msgUnknownURL)
		suite.False(ok)
	})

	suite.Run("not signup URL", func() {
		m := NewMiddleware(config.Settings{}, nil)
		_, ok := m.isPost(&msgNotSignupURL)
		suite.False(ok)
	})
}

var (
	msgIsPost = tele.Message{
		Via: &tele.User{ID: 123},
		ReplyMarkup: &tele.ReplyMarkup{InlineKeyboard: [][]tele.InlineButton{{
			{URL: "https://t.me/my_bot?start=1a2b-signup-event123-leader"},
		}}},
	}

	msgNotViaBot = tele.Message{
		Via: nil,
		ReplyMarkup: &tele.ReplyMarkup{InlineKeyboard: [][]tele.InlineButton{{
			{URL: "https://t.me/my_bot?start=1a2b-signup-event123-leader"},
		}}},
	}

	msgViaAnotherBot = tele.Message{
		Via: &tele.User{ID: 456},
		ReplyMarkup: &tele.ReplyMarkup{InlineKeyboard: [][]tele.InlineButton{{
			{URL: "https://t.me/my_bot?start=1a2b-signup-event123-leader"},
		}}},
	}

	msgNoReplyMarkup = tele.Message{
		Via:         &tele.User{ID: 123},
		ReplyMarkup: nil,
	}

	msgNoInlineKeyboard = tele.Message{
		Via:         &tele.User{ID: 123},
		ReplyMarkup: &tele.ReplyMarkup{InlineKeyboard: nil},
	}

	msgNoInlineButton = tele.Message{
		Via:         &tele.User{ID: 123},
		ReplyMarkup: &tele.ReplyMarkup{InlineKeyboard: [][]tele.InlineButton{{}}},
	}

	msgNoURL = tele.Message{
		Via: &tele.User{ID: 123},
		ReplyMarkup: &tele.ReplyMarkup{InlineKeyboard: [][]tele.InlineButton{{
			{URL: ""},
		}}},
	}

	msgUnknownURL = tele.Message{
		Via: &tele.User{ID: 123},
		ReplyMarkup: &tele.ReplyMarkup{InlineKeyboard: [][]tele.InlineButton{{
			{URL: "https://example.com"},
		}}},
	}

	msgNotSignupURL = tele.Message{
		Via: &tele.User{ID: 123},
		ReplyMarkup: &tele.ReplyMarkup{InlineKeyboard: [][]tele.InlineButton{{
			{URL: "https://t.me/my_bot?start=1a2b-another_action-event123-leader"},
		}}},
	}
)

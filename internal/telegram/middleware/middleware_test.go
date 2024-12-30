package middleware

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

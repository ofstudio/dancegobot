package app

import (
	"fmt"
	"net/http"

	"github.com/h2non/gock"
	tele "gopkg.in/telebot.v4"

	"github.com/ofstudio/dancegobot/internal/locale"
	"github.com/ofstudio/dancegobot/pkg/telegock"
)

func (suite *AppTestSuite) TestStart() {
	suite.Run("start command", func() {
		gock.New(telegock.GetUpdates).
			Reply(200).
			JSON(telegock.Updates().Message(tele.Message{
				ID:     123456,
				Sender: &tele.User{ID: 123456},
				Chat:   &tele.Chat{ID: 123456, Type: tele.ChatPrivate},
				Text:   "/start",
			}))

		gock.New(telegock.SendMessage).
			Reply(200).
			Filter(func(res *http.Response) bool {
				body := suite.Decode(res.Request.Body)
				suite.Equal(body.Get("text").String(), fmt.Sprintf(locale.Start, botUser.Username))
				return true
			})

		suite.NoPending()
		suite.NoUnmatched()
	})
}

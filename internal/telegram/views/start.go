package views

import (
	"fmt"

	tele "gopkg.in/telebot.v4"

	"github.com/ofstudio/dancegobot/internal/config"
	"github.com/ofstudio/dancegobot/internal/locale"
)

// SendStart sends a welcome message.
func SendStart(c tele.Context) error {
	rm := BtnTry()
	text := fmt.Sprintf(locale.Start, config.BotProfile().Username)
	return c.Send(text, rm, tele.ModeHTML, tele.NoPreview, tele.RemoveKeyboard)
}

// BtnTry creates a button for the "Try" option on the start message.
func BtnTry() *tele.ReplyMarkup {
	rm := &tele.ReplyMarkup{}
	rm.Inline(rm.Row(
		rm.Query(locale.BtnTry, " "),
	))
	return rm
}

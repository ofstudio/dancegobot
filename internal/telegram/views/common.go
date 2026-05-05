package views

import (
	"fmt"

	tele "gopkg.in/telebot.v4"

	"github.com/ofstudio/dancegobot/internal/locale"
)

// SendCloseOK sends a message on close reply button.
func SendCloseOK(c tele.Context) error {
	return c.Send(locale.Ok, tele.RemoveKeyboard)
}

// RemoveInlineKeyboard removes the inline keyboard from the message.
func RemoveInlineKeyboard(c tele.Context) error {
	if err := c.Edit(&tele.ReplyMarkup{}, tele.RemoveKeyboard, tele.ModeHTML, tele.NoPreview); err != nil {
		return fmt.Errorf("failed to remove inline keyboard: %w", err)
	}
	return nil
}

// btnTry creates a button to try the bot.
func btnTry() *tele.ReplyMarkup {
	rm := &tele.ReplyMarkup{}
	rm.Inline(rm.Row(
		rm.Query(locale.BtnTry, " "),
	))
	return rm
}

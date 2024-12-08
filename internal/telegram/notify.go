package telegram

import (
	"errors"
	"fmt"

	tele "gopkg.in/telebot.v4"

	"github.com/ofstudio/dancegobot/internal/locale"
	"github.com/ofstudio/dancegobot/internal/models"
)

// Notify sends notification to the recipient.
func Notify(api tele.API) func(n *models.Notification) error {
	return func(n *models.Notification) error {
		if n.Event != nil {
			n.EventID = &n.Event.ID
		}

		t, ok := locale.Notifications[n.TmplCode]
		if !ok {
			return fmt.Errorf("unknown notification template: %s", n.TmplCode)
		}
		user := &tele.User{ID: n.Recipient.ID}

		var initiatorName string
		if n.Initiator != nil {
			initiatorName = fmtName(n.Initiator.FullName(), n.Initiator)
		}
		text := fmt.Sprintf(t, n.Event.Caption, initiatorName)
		rm := btnChatLink(n.Event)
		_, err := api.Send(user, text, rm, tele.ModeHTML, tele.NoPreview, tele.RemoveKeyboard)
		if errors.Is(err, tele.ErrTrueResult) {
			return nil
		}
		return err
	}
}

package telegram

import (
	"errors"
	"fmt"
	"html/template"
	"strings"

	tele "gopkg.in/telebot.v4"

	"github.com/ofstudio/dancegobot/internal/locale"
	"github.com/ofstudio/dancegobot/internal/models"
)

// Notify sends notification to the recipient.
func Notify(api tele.API) func(n *models.Notification) error {
	return func(n *models.Notification) error {
		textSb, err := notifyText(n)
		if err != nil {
			return err
		}
		rm := btnChatLink(n.Payload.Event)

		// Send notification
		user := &tele.User{ID: n.Recipient.ID}
		_, err = api.Send(user, textSb.String(), rm, tele.ModeHTML, tele.NoPreview, tele.RemoveKeyboard)
		if errors.Is(err, tele.ErrTrueResult) {
			return nil
		}
		return err
	}
}

// notifyText returns [strings.Builder] with notification text for the given notification
func notifyText(n *models.Notification) (*strings.Builder, error) {
	sb := &strings.Builder{}
	err := notifyT.ExecuteTemplate(sb, n.TmplCode.String(), n.Payload)
	if err != nil {
		return nil, fmt.Errorf("failed to execute notification template '%s': %w", n.TmplCode, err)
	}
	return sb, nil
}

var notifyT *template.Template

// initialize notification templates
func init() {
	var err error

	// Parse notification base template
	notifyT, err = template.New("").Funcs(template.FuncMap{
		"urlTo": func(p *models.Profile) template.URL {
			return template.URL(fmtProfileURL(p))
		},
	}).Parse(locale.NotificationsBase)
	if err != nil {
		panic(fmt.Sprintf("failed to parse notification base template: %v", err))
	}

	// Parse notification templates
	for name, tmpl := range locale.Notifications {
		_, err = notifyT.New(name.String()).Parse(tmpl)
		if err != nil {
			panic(fmt.Sprintf("failed to parse notification template '%s': %v", name, err))
		}
	}
}

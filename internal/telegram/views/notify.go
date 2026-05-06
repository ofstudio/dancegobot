package views

import (
	"errors"
	"fmt"
	"html/template"
	"strings"

	tele "gopkg.in/telebot.v4"

	"github.com/ofstudio/dancegobot/internal/locale"
	"github.com/ofstudio/dancegobot/internal/models"
)

var notifyT *template.Template

// initialize notification templates
func init() {
	var err error

	// Parse notification base template
	notifyT, err = template.New("").Funcs(template.FuncMap{
		"fmtDancer": func(dancer *models.Dancer) template.HTML {
			if dancer == nil {
				return ""
			}
			return template.HTML(fmtDancer(*dancer))
		},
		"fmtProfile": func(p *models.Profile) template.HTML {
			return template.HTML(fmtProfile(p))
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

// Notify returns services.NotifyFunc function for services.NotifierService
// that sends notifications to the user.
func Notify(api tele.API) func(n *models.Notification) error {
	return func(n *models.Notification) error {
		textSb, err := notifyTextBuilder(n)
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

// notifyTextBuilder returns strings.Builder with the text for the given notification
func notifyTextBuilder(n *models.Notification) (*strings.Builder, error) {
	sb := &strings.Builder{}
	err := notifyT.ExecuteTemplate(sb, n.TmplCode.String(), n.Payload)
	if err != nil {
		return nil, fmt.Errorf("failed to execute notification template '%s': %w", n.TmplCode, err)
	}
	return sb, nil
}

// chatLink returns a link to the chat message.
//
// Known Telegram limitations:
//   - Only messages in supergroups or channels can be linked
//   - Supergroup or channel can be either public or private
//   - Bot should be a member of supergroup or an admin of the channel
//
// Link format:
//
//	https://t.me/c/{chat_link_id}/{message_id}
//
// Where {chat_link_id} = - {chat_id} - 1000000000000
//
// For example:
//
//	message_id:     1234
//	chat_id:       -1001234567890 (supergroup or channel)
//	chat_link_id:  -(-1001234567890) - 1000000000000 = 1234567890
//
// Which gives us the link: https://t.me/c/1234567890/1234
func chatLink(event *models.Event) (string, bool) {
	if event == nil ||
		event.Removed ||
		event.Post == nil ||
		event.Post.Chat == nil ||
		event.Post.ChatMessageID == 0 ||
		(event.Post.Chat.Type != models.ChatSuper && event.Post.Chat.Type != models.ChatChannel) {
		return "", false
	}

	chatLinkId := -event.Post.Chat.ID - 1000000000000
	return fmt.Sprintf("https://t.me/c/%d/%d", chatLinkId, event.Post.ChatMessageID), true
}

// btnChatLink creates a button to the event post in chat.
func btnChatLink(event *models.Event) *tele.ReplyMarkup {
	rm := &tele.ReplyMarkup{}
	link, ok := chatLink(event)
	if !ok {
		return rm
	}
	rm.Inline(rm.Row(rm.URL(locale.BtnChatLink, link)))
	return rm
}

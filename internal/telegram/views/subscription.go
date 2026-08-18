package views

import (
	"fmt"
	"html/template"

	tele "gopkg.in/telebot.v4"

	"github.com/ofstudio/dancegobot/internal/locale"
	"github.com/ofstudio/dancegobot/internal/models"
	"github.com/ofstudio/dancegobot/internal/telegram"
)

var (
	BtnSubscriptionSubscribe = tele.Btn{Unique: "subscription_subscribe"}
	BtnSubscriptionClose     = tele.Btn{Unique: "subscription_close"}
)

// EventSubscribeURL returns a deep link that subscribes the user to the event chat.
func EventSubscribeURL(eventID string) string {
	return telegram.Deeplink{
		Action:  models.SessionSubscribe,
		EventID: eventID,
	}.String()
}

// SubscriptionPrompt renders a confirmation before subscribing to an event chat.
func SubscriptionPrompt(c tele.Context, chat *models.Chat, eventID string) error {
	return c.Send(subscriptionPromptText(chat), btnSubscriptionPrompt(eventID), tele.ModeHTML, tele.NoPreview, tele.RemoveKeyboard)
}

func btnSubscriptionPrompt(eventID string) *tele.ReplyMarkup {
	rm := &tele.ReplyMarkup{}
	rm.Inline(
		rm.Row(rm.Data(locale.BtnSubscribe, BtnSubscriptionSubscribe.Unique, eventID)),
		rm.Row(rm.Data(locale.BtnClose, BtnSubscriptionClose.Unique)),
	)
	return rm
}

// SubscriptionSubscribed edits a confirmation after a successful subscription.
func SubscriptionSubscribed(c tele.Context, chat *models.Chat) error {
	return editSubscriptionMessage(c, template.HTMLEscapeString(SubscriptionSubscribedText(chat)))
}

// SubscriptionUnavailable edits a confirmation when subscription is unavailable.
func SubscriptionUnavailable(c tele.Context) error {
	return editSubscriptionMessage(c, locale.SubscriptionUnavailable)
}

// SubscriptionUnsubscribed edits a new-event notification after unsubscription.
func SubscriptionUnsubscribed(c tele.Context, chat *models.Chat, subscribeURL string) error {
	return editSubscriptionMessage(c, subscriptionUnsubscribedMessage(chat, subscribeURL))
}

func subscriptionUnsubscribedMessage(chat *models.Chat, subscribeURL string) string {
	text := template.HTMLEscapeString(SubscriptionUnsubscribedText(chat))
	if subscribeURL != "" {
		text += "\n\n" + `<a href="` + template.HTMLEscapeString(subscribeURL) + `">` +
			locale.LinkResubscribe + `</a>`
	}
	return text
}

// SubscriptionSubscribedText returns a successful subscription message.
func SubscriptionSubscribedText(chat *models.Chat) string {
	name := subscriptionChatName(chat)
	if name == "" {
		return locale.SubscriptionSubscribedChat
	}
	return fmt.Sprintf(locale.SubscriptionSubscribed, name)
}

// SubscriptionUnsubscribedText returns a successful unsubscription message.
func SubscriptionUnsubscribedText(chat *models.Chat) string {
	name := subscriptionChatName(chat)
	if name == "" {
		return locale.SubscriptionUnsubscribedChat
	}
	return fmt.Sprintf(locale.SubscriptionUnsubscribed, name)
}

func subscriptionPromptText(chat *models.Chat) string {
	name := subscriptionChatName(chat)
	if name == "" {
		return locale.SubscriptionPromptChat
	}
	return fmt.Sprintf(locale.SubscriptionPrompt, template.HTMLEscapeString(name))
}

func subscriptionChatName(chat *models.Chat) string {
	if chat == nil {
		return ""
	}
	if chat.Title != "" {
		return chat.Title
	}
	if chat.Username != "" {
		return "@" + chat.Username
	}
	return ""
}

func editSubscriptionMessage(c tele.Context, text string) error {
	err := c.Edit(text, &tele.ReplyMarkup{}, tele.ModeHTML, tele.NoPreview)
	if editErrorIsSuccess(err) {
		return nil
	}
	return err
}

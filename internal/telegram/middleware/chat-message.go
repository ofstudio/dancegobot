package middleware

import (
	"strings"

	tele "gopkg.in/telebot.v4"

	"github.com/ofstudio/dancegobot/internal/config"
	"github.com/ofstudio/dancegobot/internal/models"
	"github.com/ofstudio/dancegobot/internal/telegram"
	"github.com/ofstudio/dancegobot/internal/telegram/views"
	"github.com/ofstudio/dancegobot/pkg/telelog"
)

// ChatMessage is a middleware that adds to the event
// a message id and a chat where the event post is published.
func (m *Middleware) ChatMessage() tele.MiddlewareFunc {
	return func(next tele.HandlerFunc) tele.HandlerFunc {
		return func(c tele.Context) error {
			// Check if the message is an event post
			eventID, ok := m.isEventPost(c.Message())
			if ok {
				chatMessageID := c.Message().ID
				chat := models.NewChat(c.Message().Chat)
				event, err := m.eventService.PostChatAdd(m.ctx(c), eventID, &chat, chatMessageID)
				if err != nil {
					m.log.Error("[middleware] failed to add chat to the event post: "+err.Error(),
						"event_id", eventID,
						"chat", chat.LogValue(),
						"chat_message_id", chatMessageID,
						telelog.Trace(c))
				} else {
					m.log.Info("[middleware] chat added to the event post",
						"event", event.LogValue(),
						"post", event.Post.LogValue(),
						telelog.Trace(c))
				}
			}
			return next(c)
		}
	}
}

// isEventPost checks if Telegram message is an event post.
// If the message is an event post, it returns the event ID and true.
func (m *Middleware) isEventPost(msg *tele.Message) (string, bool) {
	if msg == nil ||
		msg.Via == nil ||
		msg.Via.ID != config.BotProfile().ID ||
		msg.ReplyMarkup == nil ||
		len(msg.ReplyMarkup.InlineKeyboard) == 0 ||
		len(msg.ReplyMarkup.InlineKeyboard[0]) == 0 {
		return "", false
	}
	btn := msg.ReplyMarkup.InlineKeyboard[0][0]
	eventID, ok := m.parseBtnEventSignupCb(btn)
	if ok {
		return eventID, true
	}

	eventID, ok = m.parseBtnEventSignupURL(btn)
	if ok {
		return eventID, true
	}

	return "", false
}

// parseBtnEventSignupCb parses the inline button data and returns the event ID.
// If the button data is not an event signup callback button, it returns an empty string and false.
func (m *Middleware) parseBtnEventSignupCb(btn tele.InlineButton) (string, bool) {
	args := strings.Split(btn.Data, "|")
	if len(args) < 3 {
		return "", false
	}
	if args[0] != "\f"+views.BtnEventSignupCb.Unique {
		return "", false
	}

	return args[1], true
}

// parseBtnEventSignupURL parses the inline button URL and returns the event ID.
// If the button URL is not an event signup url, it returns an empty string and false.
func (m *Middleware) parseBtnEventSignupURL(btn tele.InlineButton) (string, bool) {
	dl, err := telegram.DeeplinkParse(btn.URL)
	if err != nil || dl.Action != models.SessionSignup {
		return "", false
	}
	return dl.EventID, true
}

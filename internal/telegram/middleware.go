package telegram

import (
	"context"
	"log/slog"

	tele "gopkg.in/telebot.v4"

	"github.com/ofstudio/dancegobot/pkg/noplog"
	"github.com/ofstudio/dancegobot/pkg/randtoken"
	"github.com/ofstudio/dancegobot/pkg/telelog"
	"github.com/ofstudio/dancegobot/pkg/trace"
)

// Middleware is a collection of middlewares.
type Middleware struct {
	botUser *tele.User
	log     *slog.Logger
}

// NewMiddleware creates a new middleware collection.
func NewMiddleware(botUser *tele.User) *Middleware {
	return &Middleware{
		botUser: botUser,
		log:     noplog.Logger(),
	}
}

func (m *Middleware) WithLogger(l *slog.Logger) *Middleware {
	m.log = l
	return m
}

// ctx returns the context from the telebot context.
// If the context is not set, it returns a new context.Background().
func (m *Middleware) ctx(c tele.Context) context.Context {
	ctx, ok := c.Get("ctx").(context.Context)
	if !ok {
		ctx = context.Background()
	}
	return ctx
}

// Context is a middleware that sets the context for the request.
func (m *Middleware) Context(ctx context.Context) tele.MiddlewareFunc {
	return func(next tele.HandlerFunc) tele.HandlerFunc {
		return func(c tele.Context) error {
			c.Set("ctx", ctx)
			return next(c)
		}
	}
}

// Trace is a middleware that sets trace context for the request.
func (m *Middleware) Trace() tele.MiddlewareFunc {
	return func(next tele.HandlerFunc) tele.HandlerFunc {
		return func(c tele.Context) error {
			ctx := m.ctx(c)
			ctx = trace.Context(ctx, randtoken.New(8))
			c.Set("ctx", ctx)
			return next(c)
		}
	}
}

// PassPrivateMessages is a middleware that passes message updates only from private chats.
func (m *Middleware) PassPrivateMessages() tele.MiddlewareFunc {
	return func(next tele.HandlerFunc) tele.HandlerFunc {
		return func(c tele.Context) error {
			msg := c.Message()
			if msg != nil && msg.Chat != nil && msg.Chat.Type != tele.ChatPrivate {
				m.log.Warn("[bot] skipping non-private message", telelog.Attr(msg))
				return nil
			}
			return next(c)
		}
	}
}

func (m *Middleware) Logger() tele.MiddlewareFunc {
	return func(next tele.HandlerFunc) tele.HandlerFunc {
		return func(c tele.Context) error {
			// skip logging for inline queries
			if c.Query() != nil {
				return next(c)
			}
			defer func() { m.log.Info("[bot] update handled", telelog.Trace(c)) }()
			return next(c)
		}
	}
}

// ChatLink is a middleware that adds a chat where
// the event announcement was posted to the event.
//
// Known Telegram limitations:
//   - Only supergroups and channels can be linked
//   - Supergroup or channel can be either public or private
//   - Bot should be a member of supergroup or an admin in the channel
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
//
// See also:
//
//   - https://stackoverflow.com/questions/51065460/link-message-by-message-id-via-telegram-bot
//   - https://core.telegram.org/api/links
//   - https://core.telegram.org/bots/api#chat
//   - https://core.telegram.org/bots/api#message
func (m *Middleware) ChatLink() tele.MiddlewareFunc {
	// skip if the bot user is not set
	if m.botUser == nil {
		return func(next tele.HandlerFunc) tele.HandlerFunc {
			return next
		}
	}
	return func(next tele.HandlerFunc) tele.HandlerFunc {
		return func(c tele.Context) error {
			// skip non relevant messages
			if c.Message() == nil || c.Message().Text == "" ||
				c.Message().Via == nil || c.Message().Via.ID != m.botUser.ID {
				return next(c)
			}
			// todo
			// 1. Find the an appropriate event by creation date
			// 2. Add the chat to the event if event found
			return next(c)
		}
	}
}

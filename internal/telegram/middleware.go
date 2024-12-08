package telegram

import (
	"context"
	"log/slog"
	"strings"

	tele "gopkg.in/telebot.v4"

	"github.com/ofstudio/dancegobot/internal/config"
	"github.com/ofstudio/dancegobot/internal/models"
	"github.com/ofstudio/dancegobot/pkg/noplog"
	"github.com/ofstudio/dancegobot/pkg/randtoken"
	"github.com/ofstudio/dancegobot/pkg/telelog"
	"github.com/ofstudio/dancegobot/pkg/trace"
)

// Middleware is a collection of middlewares.
type Middleware struct {
	cfg    config.Settings
	events EventService
	log    *slog.Logger
}

// NewMiddleware creates a new middleware collection.
func NewMiddleware(cfg config.Settings, es EventService) *Middleware {
	return &Middleware{
		cfg:    cfg,
		events: es,
		log:    noplog.Logger(),
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

// ChatMessage is a middleware that adds to the event
// a message id and a chat where the event post is published.
func (m *Middleware) ChatMessage() tele.MiddlewareFunc {
	return func(next tele.HandlerFunc) tele.HandlerFunc {
		return func(c tele.Context) error {
			// Check if the message is an event post
			eventID, ok := m.isPost(c.Message())
			if ok {
				chatMessageID := c.Message().ID
				chat := models.NewChat(c.Message().Chat)
				upd, err := m.events.PostChatAdd(m.ctx(c), eventID, &chat, chatMessageID)
				if err != nil {
					m.log.Error("[middleware] failed to add chat to the event post: "+err.Error(), telelog.Trace(c))
				} else {
					m.log.Info("[middleware] chat added to the event post", "", upd, telelog.Trace(c))
				}
			}
			return next(c)
		}
	}
}

// isPost checks if Telegram message is an event post.
// If the message is an event post, it returns the event ID and true.
func (m *Middleware) isPost(msg *tele.Message) (string, bool) {
	if msg == nil ||
		msg.Via == nil ||
		msg.Via.ID != config.BotProfile().ID ||
		msg.ReplyMarkup == nil ||
		len(msg.ReplyMarkup.InlineKeyboard) == 0 ||
		len(msg.ReplyMarkup.InlineKeyboard[0]) == 0 {
		return "", false
	}
	btn := msg.ReplyMarkup.InlineKeyboard[0][0]
	eventID, ok := m.btnCbSignupParse(btn)
	if ok {
		return eventID, true
	}

	eventID, ok = m.btnDeeplinkParse(btn)
	if ok {
		return eventID, true
	}

	return "", false
}

func (m *Middleware) btnCbSignupParse(btn tele.InlineButton) (string, bool) {
	args := strings.Split(btn.Data, "|")
	if len(args) < 3 {
		return "", false
	}
	if args[0] != "\f"+BtnCbSignup.Unique {
		return "", false
	}

	return args[1], true
}

func (m *Middleware) btnDeeplinkParse(btn tele.InlineButton) (string, bool) {
	action, params, err := deeplinkParse(btn.URL)
	if err != nil || action != models.SessionSignup || len(params) == 0 {
		return "", false
	}
	return params[0], true
}

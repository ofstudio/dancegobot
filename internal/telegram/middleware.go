package telegram

import (
	"context"
	"log/slog"

	tele "gopkg.in/telebot.v4"

	"github.com/ofstudio/dancegobot/helpers"
	"github.com/ofstudio/dancegobot/pkg/randtoken"
	"github.com/ofstudio/dancegobot/pkg/telelog"
	"github.com/ofstudio/dancegobot/pkg/trace"
)

// Middleware is a collection of middlewares.
type Middleware struct {
	log *slog.Logger
}

// NewMiddleware creates a new middleware collection.
func NewMiddleware() *Middleware {
	return &Middleware{
		log: helpers.NopLogger(),
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

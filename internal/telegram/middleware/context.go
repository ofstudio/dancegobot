package middleware

import (
	"context"

	tele "gopkg.in/telebot.v4"

	"github.com/ofstudio/dancegobot/pkg/randtoken"
	"github.com/ofstudio/dancegobot/pkg/trace"
)

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

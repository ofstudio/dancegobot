package middleware

import (
	"fmt"

	tele "gopkg.in/telebot.v4"

	"github.com/ofstudio/dancegobot/internal/models"
)

// User is a middleware that adds the user to the context.
func (m *Middleware) User() tele.MiddlewareFunc {
	return func(next tele.HandlerFunc) tele.HandlerFunc {
		return func(c tele.Context) error {
			if c.Sender() != nil {
				user, err := m.userService.Get(m.ctx(c), models.NewProfile(*c.Sender()))
				if err != nil {
					return fmt.Errorf("failed to get user: %w", err)
				}
				c.Set("user", user)
			}
			return next(c)
		}
	}
}

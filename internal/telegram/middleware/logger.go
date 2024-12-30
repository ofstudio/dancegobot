package middleware

import (
	tele "gopkg.in/telebot.v4"

	"github.com/ofstudio/dancegobot/pkg/telelog"
)

// Logger is a middleware that logs the update.
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

package telelog

import (
	"context"
	"log/slog"

	tele "gopkg.in/telebot.v4"

	"github.com/ofstudio/dancegobot/pkg/trace"
)

// Trace returns a slog.Attr with the trace information.
func Trace(c tele.Context) slog.Attr {
	if c != nil {
		ctx, ok := c.Get("ctx").(context.Context)
		if ok {
			return trace.Attr(ctx)
		}
	}
	return slog.Group("")
}

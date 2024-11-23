package telelog

import (
	"log/slog"

	tele "gopkg.in/telebot.v4"
)

// QueryValue returns a [slog.Value] for the given [tele.Query].
func QueryValue(v tele.Query) slog.Value {
	attrs := []slog.Attr{
		slog.String("id", v.ID),
		slog.String("chat_type", v.ChatType),
	}

	if v.Sender != nil {
		attrs = append(attrs, slog.Any("from", UserValue(*v.Sender)))
	}

	return slog.GroupValue(attrs...)
}

package telelog

import (
	"log/slog"

	tele "gopkg.in/telebot.v4"
)

func InlineResultValue(v tele.InlineResult) slog.Value {
	attrs := []slog.Attr{
		slog.String("id", v.ResultID),
		slog.String("message_id", v.MessageID),
	}

	if v.Sender != nil {
		attrs = append(attrs, slog.Any("sender", UserValue(*v.Sender)))
	}

	return slog.GroupValue(attrs...)
}

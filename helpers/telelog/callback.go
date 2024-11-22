package telelog

import (
	"log/slog"

	tele "gopkg.in/telebot.v4"
)

// CallbackValue returns a [slog.Value] for the given [tele.Callback].
func CallbackValue(v tele.Callback) slog.Value {
	attrs := []slog.Attr{
		slog.String("id", v.ID),
	}

	if v.Message != nil {
		msgAttrs := []slog.Attr{
			slog.Int("id", v.Message.ID),
		}
		if v.Message.Chat != nil {
			msgAttrs = append(msgAttrs, slog.Any("chat", ChatValue(*v.Message.Chat)))
		}
		attrs = append(attrs, slog.Any("message", slog.GroupValue(msgAttrs...)))
	}

	if v.Sender != nil {
		attrs = append(attrs, slog.Any("sender", UserValue(*v.Sender)))
	}

	if v.MessageID != "" {
		attrs = append(attrs, slog.String("inline_message_id", v.MessageID))
	}

	if v.Data != "" {
		attrs = append(attrs, slog.String("data", v.Data))
	}

	return slog.GroupValue(attrs...)
}

package telelog

import (
	"log/slog"

	tele "gopkg.in/telebot.v4"
)

func ChatValue(v tele.Chat) slog.Value {
	attrs := []slog.Attr{
		slog.Int64("id", v.ID),
		slog.String("type", string(v.Type)),
	}

	if v.Title != "" {
		attrs = append(attrs, slog.String("title", v.Title))
	}

	if v.FirstName != "" {
		attrs = append(attrs, slog.String("first_name", v.FirstName))
	}

	if v.LastName != "" {
		attrs = append(attrs, slog.String("last_name", v.LastName))
	}

	if v.Username != "" {
		attrs = append(attrs, slog.String("username", v.Username))
	}

	return slog.GroupValue(attrs...)
}

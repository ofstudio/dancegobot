package telelog

import (
	"log/slog"

	tele "gopkg.in/telebot.v4"
)

// UserValue returns a [slog.Value] for the given [tele.User].
func UserValue(v tele.User) slog.Value {
	attrs := []slog.Attr{
		slog.Int64("id", v.ID),
		slog.String("first_name", v.FirstName),
	}

	if v.LastName != "" {
		attrs = append(attrs, slog.String("last_name", v.LastName))
	}

	if v.Username != "" {
		attrs = append(attrs, slog.String("username", v.Username))
	}

	return slog.GroupValue(attrs...)
}

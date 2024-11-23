package telelog

import (
	"log/slog"

	tele "gopkg.in/telebot.v4"
)

// UpdateAttr returns a [slog.Attr] for the given [tele.Update].
// Supported update types:
//   - [tele.Message]
//   - [tele.Callback]
//   - [tele.Query]
//   - [tele.InlineResult]
func UpdateAttr(v tele.Update) slog.Attr {
	var t string
	var payload slog.Attr

	switch {
	case v.Message != nil:
		t = "message"
		payload = slog.Any("message", MessageValue(*v.Message))
	case v.Callback != nil:
		t = "callback"
		payload = slog.Any("callback", CallbackValue(*v.Callback))
	case v.Query != nil:
		t = "query"
		payload = slog.Any("query", QueryValue(*v.Query))
	case v.InlineResult != nil:
		t = "inline_result"
		payload = slog.Any("inline_result", InlineResultValue(*v.InlineResult))
	default:
		t = "unsupported"
	}

	return slog.Group("",
		slog.Group("update",
			slog.Int("id", v.ID),
			slog.String("type", t),
		), payload)
}

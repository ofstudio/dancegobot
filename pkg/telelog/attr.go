package telelog

import (
	"log/slog"
	"reflect"

	tele "gopkg.in/telebot.v4"
)

// Attr returns a slog.Attr for the given value.
// Values supported:
//   - [tele.Context]
//   - [tele.Update]
//   - [tele.Message]
//   - [tele.Callback]
//   - [tele.Query]
//   - [tele.InlineResult]
//   - [tele.User]
//   - [tele.Chat]
func Attr(val any) slog.Attr {
	switch v := val.(type) {
	case tele.Context:
		return slog.Group("", UpdateAttr(v.Update()), Trace(v))
	case tele.Update:
		return UpdateAttr(v)
	case *tele.Update:
		return UpdateAttr(*v)
	case tele.Message:
		return slog.Any("message", MessageValue(v))
	case *tele.Message:
		return slog.Any("message", MessageValue(*v))
	case tele.Callback:
		return slog.Any("callback", CallbackValue(v))
	case *tele.Callback:
		return slog.Any("callback", CallbackValue(*v))
	case tele.Query:
		return slog.Any("query", QueryValue(v))
	case *tele.Query:
		return slog.Any("query", QueryValue(*v))
	case tele.InlineResult:
		return slog.Any("inline_result", InlineResultValue(v))
	case *tele.InlineResult:
		return slog.Any("inline_result", InlineResultValue(*v))
	case tele.User:
		return slog.Any("user", UserValue(v))
	case *tele.User:
		return slog.Any("user", UserValue(*v))
	case tele.Chat:
		return slog.Any("chat", ChatValue(v))
	case *tele.Chat:
		return slog.Any("chat", ChatValue(*v))
	default:
		return slog.String("unsupported_type", reflect.TypeOf(val).String())
	}
}

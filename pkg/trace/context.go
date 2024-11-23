package trace

import (
	"context"
	"log/slog"
	"time"
)

type callIDKeyType struct{}
type startTimeKeyType struct{}

var (
	callIDKey    = callIDKeyType{}
	startTimeKey = startTimeKeyType{}
)

// Context returns a new context with the call ID
func Context(ctx context.Context, callID string) context.Context {
	ctx = context.WithValue(ctx, callIDKey, callID)
	return context.WithValue(ctx, startTimeKey, time.Now().UTC())
}

// Attr returns a [slog.Attr] with the call ID from given context
func Attr(ctx context.Context) slog.Attr {
	var attrs []any
	if callID, ok := ctx.Value(callIDKey).(string); ok {
		attrs = append(attrs, slog.String("call_id", callID))
	}
	if startTime, ok := ctx.Value(startTimeKey).(time.Time); ok {
		d := time.Since(startTime).Truncate(time.Millisecond)
		attrs = append(attrs, slog.Duration("elapsed", d))
	}
	return slog.Group("", attrs...)
}

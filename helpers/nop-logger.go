package helpers

import "log/slog"

type nopWriter struct{}

func (nopWriter) Write([]byte) (int, error) { return 0, nil }

func NopLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(nopWriter{}, nil))
}

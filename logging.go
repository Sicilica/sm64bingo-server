package bingo

import (
	"context"
	"log/slog"
)

type loggerKeyType int

var loggerKey loggerKeyType

func Logger(ctx context.Context) *slog.Logger {
	l, ok := ctx.Value(loggerKey).(*slog.Logger)
	if ok {
		return l
	}
	return slog.Default()
}

func WithLogger(ctx context.Context, logger *slog.Logger) context.Context {
	return context.WithValue(ctx, loggerKey, logger)
}

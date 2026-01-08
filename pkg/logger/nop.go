package logger

import "context"

var _ Logger = (*NopLogger)(nil)

type NopLogger struct {
}

func NewNopLogger() Logger {
	return &NopLogger{}
}

func (l *NopLogger) Debug(ctx context.Context, msg string, args ...Field) {
}

func (l *NopLogger) Info(ctx context.Context, msg string, args ...Field) {}

func (l *NopLogger) Warn(ctx context.Context, msg string, args ...Field) {}

func (l *NopLogger) Error(ctx context.Context, msg string, args ...Field) {}

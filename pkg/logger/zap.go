package logger

import (
	"context"

	"go.uber.org/zap"
)

var _ Logger = (*ZapLogger)(nil)

type ZapLogger struct {
	logger *zap.Logger
}

func NewZapLogger(logger *zap.Logger) Logger {
	return &ZapLogger{logger: logger}
}

func (l *ZapLogger) Info(ctx context.Context, msg string, args ...Field) {
	l.logger.Info(msg, l.toArgs(args)...)
}

func (l *ZapLogger) Debug(ctx context.Context, msg string, args ...Field) {
	l.logger.Debug(msg, l.toArgs(args)...)
}

func (l *ZapLogger) Warn(ctx context.Context, msg string, args ...Field) {
	l.logger.Warn(msg, l.toArgs(args)...)
}

func (l *ZapLogger) Error(ctx context.Context, msg string, args ...Field) {
	l.logger.Error(msg, l.toArgs(args)...)
}

func (l *ZapLogger) toArgs(args []Field) []zap.Field {
	res := make([]zap.Field, 0, len(args))
	for _, arg := range args {
		res = append(res, zap.Any(arg.Key, arg.Val))
	}
	return res
}

package logger

import (
	"context"

	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"go.uber.org/zap"
)

// OtelZapLogger 是 Logger 接口的具体实现
type OtelZapLogger struct {
	l *otelzap.Logger
}

// NewOtelZapLogger 构造函数
func NewOtelZapLogger(l *otelzap.Logger) Logger {
	return &OtelZapLogger{
		l: l,
	}
}

func (o *OtelZapLogger) Debug(ctx context.Context, msg string, args ...Field) {
	// 核心魔法：otelzap.Ctx(ctx) 会自动从 context 提取 TraceID
	o.l.Ctx(ctx).Debug(msg, o.toZapFields(args)...)
}

func (o *OtelZapLogger) Info(ctx context.Context, msg string, args ...Field) {
	o.l.Ctx(ctx).Info(msg, o.toZapFields(args)...)
}

func (o *OtelZapLogger) Warn(ctx context.Context, msg string, args ...Field) {
	o.l.Ctx(ctx).Warn(msg, o.toZapFields(args)...)
}

func (o *OtelZapLogger) Error(ctx context.Context, msg string, args ...Field) {
	o.l.Ctx(ctx).Error(msg, o.toZapFields(args)...)
}

// toZapFields 将我们自定义的 Field 转换为 zap.Field
func (o *OtelZapLogger) toZapFields(args []Field) []zap.Field {
	res := make([]zap.Field, 0, len(args))
	for _, arg := range args {
		res = append(res, zap.Any(arg.Key, arg.Val))
	}
	return res
}

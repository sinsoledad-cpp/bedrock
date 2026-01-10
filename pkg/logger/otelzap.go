package logger

import (
	"context"

	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"go.opentelemetry.io/otel/trace"
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
	args = o.injectTraceContext(ctx, args)
	o.l.Ctx(ctx).Debug(msg, o.toZapFields(args)...)
}

func (o *OtelZapLogger) Info(ctx context.Context, msg string, args ...Field) {
	args = o.injectTraceContext(ctx, args)
	o.l.Ctx(ctx).Info(msg, o.toZapFields(args)...)
}

func (o *OtelZapLogger) Warn(ctx context.Context, msg string, args ...Field) {
	args = o.injectTraceContext(ctx, args)
	o.l.Ctx(ctx).Warn(msg, o.toZapFields(args)...)
}

func (o *OtelZapLogger) Error(ctx context.Context, msg string, args ...Field) {
	args = o.injectTraceContext(ctx, args)
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
func (o *OtelZapLogger) injectTraceContext(ctx context.Context, args []Field) []Field {
	if span := trace.SpanFromContext(ctx); span.SpanContext().IsValid() {
		// 将 trace_id 和 span_id 拼接到 args 切片中
		args = append(args,
			Field{Key: "trace_id", Val: span.SpanContext().TraceID().String()},
			Field{Key: "span_id", Val: span.SpanContext().SpanID().String()},
		)
	}
	return args
}

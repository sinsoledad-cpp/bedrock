package ioc

import (
	"bedrock/pkg/logger"

	"github.com/spf13/viper"
	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"go.uber.org/zap"
)

func InitLogger() logger.Logger {

	mode := viper.GetString("server.mode")
	if mode == "" {
		mode = "debug"
	}

	if err := InitZap(mode); err != nil {
		panic(err)
	}

	baseLogger := zap.L()
	otelLogger := otelzap.New(
		baseLogger,
		//otelzap.WithTraceIDField(true), // 开启自动注入 trace_id
		otelzap.WithMinLevel(zap.DebugLevel),
	)

	otelzap.ReplaceGlobals(otelLogger)
	return logger.NewOtelZapLogger(otelLogger)
}

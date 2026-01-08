package ioc

import (
	"bedrock/pkg/logger"

	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// InitLogger 初始化 Logger 组件
// 注意：返回值是接口 logger.Logger，实现了依赖倒置
func InitLogger() logger.Logger {
	// 1. 配置原生 Zap
	cfg := zap.NewDevelopmentConfig()
	cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	baseLogger, err := cfg.Build()
	if err != nil {
		panic(err)
	}

	// 2. 包装为 otelzap
	// WithTraceIDField(true) 确保 trace_id 字段被自动加入日志
	otelLogger := otelzap.New(
		baseLogger,
		otelzap.WithMinLevel(zap.DebugLevel),
	)

	// (可选) 替换全局 Logger，方便第三方库使用
	otelzap.ReplaceGlobals(otelLogger)

	// 3. 返回我们的适配器
	return logger.NewOtelZapLogger(otelLogger)
}

// func InitLogger() logger.Logger {
// 	//type LogConfig struct {
// 	//	Level      string `mapstructure:"level"`
// 	//	Filename   string `mapstructure:"filename"`
// 	//	MaxSize    int    `mapstructure:"max_size"`
// 	//	MaxAge     int    `mapstructure:"max_age"`
// 	//	MaxBackups int    `mapstructure:"max_backups"`
// 	//}
// 	//var customCfg LogConfig
// 	//viper.UnmarshalKey("log", &customCfg)

// 	//// 这里我们用一个小技巧，
// 	//// 就是直接使用 zap 本身的配置结构体来处理
// 	//cfg := zap.NewDevelopmentConfig()
// 	////err := viper.UnmarshalKey("log", &cfg)
// 	////if err != nil {
// 	////	panic(err)
// 	////}
// 	//l, err := cfg.Build()
// 	//if err != nil {
// 	//	panic(err)
// 	//}

// 	if err := InitZap("debug"); err != nil {
// 		panic(err)
// 	}

// 	return logger.NewZapLogger(lg)
// }

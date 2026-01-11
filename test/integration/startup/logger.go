package startup

import "bedrock/pkg/logger"

func InitLogger() logger.Logger {
	return logger.NewNopLogger()
}

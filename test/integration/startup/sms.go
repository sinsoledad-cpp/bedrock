package startup

import (
	"bedrock/internal/service"
	"bedrock/internal/service/sms/memory"
)

func InitCodeService() service.CodeService {
	// 验证码服务使用内存实现，方便测试
	return service.NewCodeService(nil, memory.NewService())
}

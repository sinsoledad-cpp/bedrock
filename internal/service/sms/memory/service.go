package memory

import (
	"bedrock/internal/service/sms"
	"context"
	"fmt"
)

type Service struct {
}

func NewService() sms.Service {
	return &Service{}
}

func (s *Service) Send(ctx context.Context, tplId string, args []string, numbers ...string) error {
	fmt.Println("验证码是", args)
	return nil
}

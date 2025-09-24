package service

import (
	"bedrock/internal/domain"
	"bedrock/internal/repository"
	"bedrock/pkg/logger"
	"context"
	"errors"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrDuplicateEmail        = repository.ErrDuplicateUser
	ErrInvalidUserOrPassword = errors.New("用户不存在或者密码不对")
)

//go:generate mockgen -source=./user.go -package=mocks -destination=./mocks/user.mock.go UserService
type UserService interface {
	Signup(ctx context.Context, user domain.User) error
	Login(ctx context.Context, email string, password string) (domain.User, error)
	//UpdateNonSensitiveInfo(ctx context.Context, user domain.User) error
	//FindById(ctx context.Context, uid int64) (domain.User, error)
	//FindOrCreate(ctx context.Context, phone string) (domain.User, error)
	//FindOrCreateByWechat(ctx context.Context, wechatInfo domain.WechatInfo) (domain.User, error)
}
type DefaultUserService struct {
	logger logger.Logger
	repo   repository.UserRepository
}

func NewUserService(logger logger.Logger, repo repository.UserRepository) UserService {
	return &DefaultUserService{
		logger: logger,
		repo:   repo,
	}
}

func (svc *DefaultUserService) Signup(ctx context.Context, u domain.User) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost) // bcrypt.DefaultCost 表示加密的复杂度，默认值为 10。
	if err != nil {
		return err
	}
	u.Password = string(hash)
	return svc.repo.Create(ctx, u)
}
func (svc *DefaultUserService) Login(ctx context.Context, email string, password string) (domain.User, error) {
	u, err := svc.repo.FindByEmail(ctx, email)
	if errors.Is(err, repository.ErrUserNotFound) {
		return domain.User{}, ErrInvalidUserOrPassword
	}
	if err != nil {
		return domain.User{}, err
	}
	// 检查密码对不对
	err = bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	if err != nil {
		return domain.User{}, ErrInvalidUserOrPassword
	}
	return u, nil
}

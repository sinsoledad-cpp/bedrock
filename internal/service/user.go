package service

import (
	"bedrock/internal/domain"
	"bedrock/internal/repository"
	"bedrock/pkg/logger"
	"context"
	"errors"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
)

var (
	ErrDuplicateEmail        = repository.ErrDuplicateUser
	ErrInvalidUserOrPassword = errors.New("用户不存在或者密码不对")
)

//go:generate mockgen -source=./user.go -package=mocks -destination=./mocks/user.mock.go UserService
type UserService interface {
	Signup(ctx context.Context, user domain.User) error
	Login(ctx context.Context, email string, password string) (domain.User, error)
	UploadAvatar(ctx context.Context, uid int64, file *multipart.FileHeader) (string, error)
	UpdateNonSensitiveInfo(ctx context.Context, user domain.User) error
	FindById(ctx context.Context, uid int64) (domain.User, error)
	//FindOrCreate(ctx context.Context, phone string) (domain.User, error)
	//FindOrCreateByWechat(ctx context.Context, wechatInfo domain.WechatInfo) (domain.User, error)
}
type DefaultUserService struct {
	l    logger.Logger
	repo repository.UserRepository
}

func NewUserService(log logger.Logger, repo repository.UserRepository) UserService {
	return &DefaultUserService{
		l:    log,
		repo: repo,
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
func (svc *DefaultUserService) UploadAvatar(ctx context.Context, uid int64, file *multipart.FileHeader) (string, error) {
	// 0. (新增) 从数据库中获取旧头像的路径，用于后续删除
	// 注意：这里我们假设 svc.repo 有一个方法可以获取用户信息，从而得到旧头像路径
	// 实际项目中，你需要根据你的 repo 设计来实现
	oldUser, err := svc.repo.FindById(ctx, uid) // 假设有这么一个方法
	if err != nil {
		return "", err // 获取用户信息失败，直接返回
	}
	oldAvatarPath := oldUser.Avatar // 假设用户结构体中有 Avatar 字段

	// 1. 生成新文件的路径和文件名
	ext := filepath.Ext(file.Filename)
	newPath := filepath.Join("uploads", "avatars", uuid.New().String()+ext)

	// 2. 创建目录 (逻辑不变)
	if err := os.MkdirAll(filepath.Dir(newPath), os.ModePerm); err != nil {
		return "", err
	}

	// 3. 保存新文件 (逻辑不变)
	src, err := file.Open()
	if err != nil {
		return "", err
	}
	defer func(src multipart.File) {
		if err := src.Close(); err != nil {
			svc.l.Warn("关闭新头像文件失败", logger.Error(err))
		}
	}(src)

	dst, err := os.Create(newPath) //创建或者清空文件
	if err != nil {
		return "", err
	}
	defer func(dst *os.File) {
		if err := dst.Close(); err != nil {
			svc.l.Warn("关闭老头像文件失败", logger.Error(err))
		}
	}(dst)

	if _, err = io.Copy(dst, src); err != nil {
		return "", err
	}

	// 4. 更新数据库中的用户头像为新路径
	err = svc.repo.UpdateAvatar(ctx, uid, newPath)
	if err != nil {
		// 如果数据库更新失败，我们需要删除刚刚保存的新文件，进行“回滚”
		// 同时，旧文件没有被动过，保证了数据状态的回退
		if err := os.Remove(newPath); err != nil {
			svc.l.Warn("数据库更新失败进行回滚操作,但是删除新头像失败", logger.Error(err), logger.String("new_avatar_path", newPath))
		}
		return "", err
	}

	// 5. (新增) 数据库更新成功后，删除旧的头像文件
	// 检查 oldAvatarPath 是否为空或默认值，避免删除默认头像
	if oldAvatarPath != "" && oldAvatarPath != "path/to/default/avatar.png" {
		// 删除旧文件失败是一个可以容忍的错误，不应该影响主流程的成功返回
		// 但我们应该记录日志，方便后续手动清理或进行监控
		if err := os.Remove(oldAvatarPath); err != nil {
			// 在这里记录日志，但不要返回错误
			svc.l.Warn("数据库更新成功,删除旧头像失败", logger.Error(err), logger.String("old_avatar_path", oldAvatarPath))
		}
	}

	// 6. 返回新文件的路径
	return newPath, nil
}
func (svc *DefaultUserService) UpdateNonSensitiveInfo(ctx context.Context, user domain.User) error {
	// UpdateNicknameAndXXAnd
	return svc.repo.UpdateNonZeroFields(ctx, user)
}
func (svc *DefaultUserService) FindById(ctx context.Context, uid int64) (domain.User, error) {
	return svc.repo.FindById(ctx, uid)
}

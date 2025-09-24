package repository

import (
	"bedrock/internal/domain"
	"bedrock/internal/repository/cache"
	"bedrock/internal/repository/dao"
	"context"
	"database/sql"
)

var (
	ErrDuplicateUser = dao.ErrDuplicateEmail
	ErrUserNotFound  = dao.ErrRecordNotFound
)

type UserRepository interface {
	Create(ctx context.Context, user domain.User) error
	//FindByEmail(ctx context.Context, email string) (domain.User, error)
	//UpdateNonZeroFields(ctx context.Context, user domain.User) error
	//FindByPhone(ctx context.Context, phone string) (domain.User, error)
	//FindById(ctx context.Context, uID int64) (domain.User, error)
	//FindByWechat(ctx context.Context, openID string) (domain.User, error)
}

type CachedUserRepository struct {
	userDAO   dao.UserDAO
	userCache cache.UserCache
}

func NewCachedUserRepository(userDAO dao.UserDAO, userCache cache.UserCache) UserRepository {
	return &CachedUserRepository{
		userDAO:   userDAO,
		userCache: userCache,
	}
}

func (u *CachedUserRepository) Create(ctx context.Context, user domain.User) error {
	return u.userDAO.Insert(ctx, u.toEntity(user))
}

func (u *CachedUserRepository) toEntity(user domain.User) dao.User {
	return dao.User{
		ID: user.ID,
		Email: sql.NullString{
			String: user.Email,
			Valid:  user.Email != "",
		},
		Phone: sql.NullString{
			String: user.Phone,
			Valid:  user.Phone != "",
		},
		Password: user.Password,
		Birthday: user.Birthday.UnixMilli(),
		WechatUnionId: sql.NullString{
			String: user.WechatInfo.UnionID,
			Valid:  user.WechatInfo.UnionID != "",
		},
		WechatOpenId: sql.NullString{
			String: user.WechatInfo.OpenID,
			Valid:  user.WechatInfo.OpenID != "",
		},
		AboutMe:  user.AboutMe,
		Nickname: user.Nickname,
	}
}

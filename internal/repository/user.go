package repository

import (
	"bedrock/internal/domain"
	"bedrock/internal/repository/cache"
	"bedrock/internal/repository/dao"
	"context"
	"database/sql"
	"time"
)

var (
	ErrDuplicateUser = dao.ErrDuplicateEmail
	ErrUserNotFound  = dao.ErrRecordNotFound
)

type UserRepository interface {
	Create(ctx context.Context, user domain.User) error
	FindByEmail(ctx context.Context, email string) (domain.User, error)
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

func (c *CachedUserRepository) Create(ctx context.Context, user domain.User) error {
	return c.userDAO.Insert(ctx, c.toEntity(user))
}
func (c *CachedUserRepository) FindByEmail(ctx context.Context, email string) (domain.User, error) {
	u, err := c.userDAO.FindByEmail(ctx, email)
	if err != nil {
		return domain.User{}, err
	}
	return c.toDomain(u), nil
}
func (c *CachedUserRepository) toEntity(user domain.User) dao.User {
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
func (c *CachedUserRepository) toDomain(u dao.User) domain.User {
	return domain.User{
		ID:       u.ID,
		Email:    u.Email.String,
		Phone:    u.Phone.String,
		Password: u.Password,
		AboutMe:  u.AboutMe,
		Nickname: u.Nickname,
		Birthday: time.UnixMilli(u.Birthday),
		Ctime:    time.UnixMilli(u.Ctime),
		WechatInfo: domain.WechatInfo{
			OpenID:  u.WechatOpenId.String,
			UnionID: u.WechatUnionId.String,
		},
	}
}

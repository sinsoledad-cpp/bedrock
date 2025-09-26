package repository

import (
	"bedrock/internal/domain"
	"bedrock/internal/repository/cache"
	"bedrock/internal/repository/dao"
	"context"
	"database/sql"
	"errors"
	"time"
)

var (
	ErrDuplicateUser = dao.ErrDuplicateEmail
	ErrUserNotFound  = dao.ErrRecordNotFound
)

type UserRepository interface {
	Create(ctx context.Context, user domain.User) error
	FindByEmail(ctx context.Context, email string) (domain.User, error)
	UpdateAvatar(ctx context.Context, id int64, avatar string) error
	UpdateNonZeroFields(ctx context.Context, user domain.User) error
	//FindByPhone(ctx context.Context, phone string) (domain.User, error)

	FindById(ctx context.Context, uID int64) (domain.User, error)
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

func (c *CachedUserRepository) UpdateAvatar(ctx context.Context, id int64, avatar string) error {
	// 更新数据库
	err := c.userDAO.UpdateAvatar(ctx, id, avatar)
	//if err != nil {
	//	return err
	//}
	//// 操作缓存：这里选择直接删除缓存，让下一次查询重新加载
	//return c.userCache.Delete(ctx, id)
	return err
}

func (c *CachedUserRepository) UpdateNonZeroFields(ctx context.Context, user domain.User) error {
	// 更新 DB 之后，删除
	err := c.userDAO.UpdateById(ctx, c.toEntity(user))
	if err != nil {
		return err
	}
	// 延迟一秒
	time.AfterFunc(time.Second, func() {
		_ = c.userCache.Delete(ctx, user.ID)
	})
	return c.userCache.Delete(ctx, user.ID)
}

func (c *CachedUserRepository) FindById(ctx context.Context, uid int64) (domain.User, error) {
	du, err := c.userCache.Get(ctx, uid)
	switch {
	case err == nil: // 只要 err 为 nil，就返回
		return du, nil
	case errors.Is(err, cache.ErrKeyNotExist):
		u, err := c.userDAO.FindById(ctx, uid)
		if err != nil {
			return domain.User{}, err
		}
		du = c.toDomain(u)
		//go func() {
		//	err = repo.cache.Set(ctx, du)
		//	if err != nil {
		//		log.Println(err)
		//	}
		//}()
		err = c.userCache.Set(ctx, du)
		if err != nil {
			// 网络崩了，也可能是 redis 崩了
		}
		return du, nil
	default:
		// 接近降级的写法
		return domain.User{}, err
	}
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

package dao

import (
	"context"
	"database/sql"
	"errors"
	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
	"strings"
	"time"
)

type User struct {
	ID       int64          `gorm:"primaryKey,autoIncrement"`
	Email    sql.NullString `gorm:"unique"` //Email    *string // 代表这是一个可以为 NULL 的列
	Password string
	Nickname string         `gorm:"type=varchar(128)"`
	Birthday int64          // YYYY-MM-DD
	AboutMe  string         `gorm:"type=varchar(4096)"`
	Phone    sql.NullString `gorm:"unique"` // 代表这是一个可以为 NULL 的列
	// 1 如果查询要求同时使用 openid 和 unionid，就要创建联合唯一索引
	// 2 如果查询只用 openid，那么就在 openid 上创建唯一索引，或者 <openid, unionId> 联合索引
	// 3 如果查询只用 unionid，那么就在 unionid 上创建唯一索引，或者 <unionid, openid> 联合索引
	WechatOpenId  sql.NullString `gorm:"unique"`
	WechatUnionId sql.NullString
	Ctime         int64 // 创建时间 // 时区，UTC 0 的毫秒数
	Utime         int64 // 更新时间
	// json 存储
	//Addr string
}

var (
	ErrDuplicateEmail = errors.New("邮箱冲突")
	ErrDuplicatePhone = errors.New("手机号冲突")
	ErrRecordNotFound = gorm.ErrRecordNotFound
)

type UserDAO interface {
	Insert(ctx context.Context, user User) error
	FindByEmail(ctx context.Context, email string) (User, error)
	//UpdateById(ctx context.Context, entity User) error
	//FindById(ctx context.Context, uid int64) (User, error)
	//FindByPhone(ctx context.Context, phone string) (User, error)
	//FindByWechat(ctx context.Context, openId string) (User, error)
}

type GORMUserDAO struct {
	db *gorm.DB
}

func NewGORMUserDAO(db *gorm.DB) UserDAO {
	return &GORMUserDAO{
		db: db,
	}
}
func (g *GORMUserDAO) Insert(ctx context.Context, user User) error {
	now := time.Now().UnixMilli()
	user.Ctime = now
	user.Utime = now
	err := g.db.WithContext(ctx).Create(&user).Error
	var e *mysql.MySQLError
	if errors.As(err, &e) {

		const uniqueIndexErrNo uint16 = 1062
		if e.Number == uniqueIndexErrNo {

			if strings.Contains(e.Message, "email") {
				return ErrDuplicateEmail
			}
			if strings.Contains(e.Message, "phone") {
				return ErrDuplicatePhone
			}

			return ErrDuplicateEmail
		}

	}
	return err
}

func (g *GORMUserDAO) FindByEmail(ctx context.Context, email string) (User, error) {
	var u User
	err := g.db.WithContext(ctx).Where("email=?", email).First(&u).Error
	return u, err
}

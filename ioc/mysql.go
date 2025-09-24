package ioc

import (
	"bedrock/internal/repository/dao"
	"bedrock/pkg/logger"
	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	glogger "gorm.io/gorm/logger"
)

func InitMySQL(l logger.Logger) *gorm.DB {
	type Config struct {
		DSN string `yaml:"dsn"`
	}
	var cfg Config = Config{
		DSN: "root:root@tcp(localhost:3306)/bedrock?charset=utf8mb4&parseTime=True&loc=Local",
	}
	err := viper.UnmarshalKey("mysql", &cfg)
	if err != nil {
		panic(err)
	}
	db, err := gorm.Open(mysql.Open(cfg.DSN),
		&gorm.Config{
			Logger: glogger.New(gormLoggerFunc(l.Debug), glogger.Config{
				// 慢查询
				SlowThreshold: 0,
				LogLevel:      glogger.Info,
			}),
		},
	)
	if err != nil {
		panic(err)
	}

	err = dao.InitTables(db)
	if err != nil {
		panic(err)
	}
	return db
}

type gormLoggerFunc func(msg string, fields ...logger.Field)

func (g gormLoggerFunc) Printf(s string, i ...interface{}) {
	g(s, logger.Field{Key: "args", Val: i})
}

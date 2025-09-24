//go:build wireinject

package main

import (
	"bedrock/internal/repository"
	"bedrock/internal/repository/cache"
	"bedrock/internal/repository/dao"
	"bedrock/internal/service"
	"bedrock/internal/web"
	"bedrock/ioc"
	"github.com/google/wire"
)

var thirdParty = wire.NewSet(
	ioc.InitLogger,
	ioc.InitMySQL,
	ioc.InitRedis,
)

var userHdl = wire.NewSet(
	cache.NewRedisUserCache,
	dao.NewGORMUserDAO,
	repository.NewCachedUserRepository,
	service.NewUserService,
	web.NewUserHandler,
)

func InitApp() *App {
	wire.Build(
		thirdParty,
		userHdl,
		ioc.InitWebEngine,
		wire.Struct(new(App), "*"),
	)
	return new(App)
}

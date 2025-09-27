//go:build wireinject

package main

import (
	"bedrock/internal/repository"
	"bedrock/internal/repository/cache"
	"bedrock/internal/repository/dao"
	"bedrock/internal/service"
	"bedrock/internal/web"
	"bedrock/internal/web/middleware/jwt"
	"bedrock/ioc"
	"github.com/google/wire"
)

var thirdParty = wire.NewSet(
	ioc.InitLogger,
	ioc.InitMySQL,
	ioc.InitRedis,
)

var userSvc = wire.NewSet(
	cache.NewRedisUserCache,
	dao.NewGORMUserDAO,
	repository.NewCachedUserRepository,
	service.NewUserService,
)

var codeSvc = wire.NewSet(
	cache.NewRedisCodeCache,
	repository.NewCachedCodeRepository,
	ioc.InitSMSService,
	service.NewCodeService,
)

func InitApp() *App {
	wire.Build(
		thirdParty,
		userSvc,
		codeSvc,

		jwt.NewRedisJWTHandler,
		web.NewUserHandler,

		ioc.InitWebEngine,
		wire.Struct(new(App), "*"),
	)
	return new(App)
}

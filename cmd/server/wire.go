//go:build wireinject

package main

import (
	ioc2 "bedrock/cmd/server/ioc"
	"bedrock/internal/repository"
	"bedrock/internal/repository/cache"
	"bedrock/internal/repository/dao"
	"bedrock/internal/service"
	"bedrock/internal/web"
	"bedrock/internal/web/middleware/jwt"

	"github.com/google/wire"
)

var thirdParty = wire.NewSet(
	ioc2.InitLogger,
	ioc2.InitMySQL,
	ioc2.InitRedis,
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
	ioc2.InitSMSService,
	service.NewCodeService,
)

//var wechatSvc = wire.NewSet(
//	ioc.InitWechatService,
//)

func InitApp() *App {
	wire.Build(
		thirdParty,

		userSvc,
		codeSvc,
		//wechatSvc,

		jwt.NewRedisJWTHandler,
		web.NewUserHandler,
		//web.NewOAuth2WechatHandler,

		ioc2.InitWebEngine,
		ioc2.InitGinMiddlewares,
		wire.Struct(new(App), "*"),
	)
	return new(App)
}

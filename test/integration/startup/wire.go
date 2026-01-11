//go:build wireinject

package startup

import (
	"bedrock/internal/repository"
	"bedrock/internal/repository/cache"
	"bedrock/internal/repository/dao"
	"bedrock/internal/service"
	"bedrock/internal/web"
	"bedrock/internal/web/middleware/jwt"

	"github.com/gin-gonic/gin"
	"github.com/google/wire"
)

var thirdParty = wire.NewSet(
	InitLogger,
	InitMySQL,
	InitRedis,
	InitStorageService,
	InitCodeService,
)

var userSvc = wire.NewSet(
	cache.NewRedisUserCache,
	dao.NewGORMUserDAO,
	repository.NewCachedUserRepository,
	service.NewUserService,
)

func InitUserHandler() *web.UserHandler {
	wire.Build(
		thirdParty,
		userSvc,
		jwt.NewRedisJWTHandler,
		web.NewUserHandler,
	)
	return new(web.UserHandler)
}

func InitWebServer() *gin.Engine {
	wire.Build(
		thirdParty,
		userSvc,
		jwt.NewRedisJWTHandler,
		web.NewUserHandler,
		InitGinServer,
	)
	return new(gin.Engine)
}

package ioc

import (
	"bedrock/internal/web"
	"bedrock/internal/web/middleware"
	"bedrock/internal/web/middleware/jwt"
	"bedrock/pkg/ginx"
	"bedrock/pkg/logger"
	"github.com/gin-gonic/gin"
)

func InitWebEngine(middlewares []gin.HandlerFunc, l logger.Logger, userHdl *web.UserHandler) *gin.Engine {
	gin.ForceConsoleColor()
	engine := gin.Default()
	ginx.SetLogger(l)
	engine.Use(middlewares...)
	userHdl.RegisterRoutes(engine)
	//wechatHdl.RegisterRoutes(engine)//, wechatHdl *web.OAuth2WechatHandler
	return engine
}
func InitGinMiddlewares(jwtHdl jwt.Handler) []gin.HandlerFunc {

	return []gin.HandlerFunc{
		middleware.NewJWTAuth(jwtHdl).Middleware(),
	}
}

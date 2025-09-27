package ioc

import (
	"bedrock/internal/web"
	"bedrock/pkg/ginx"
	"bedrock/pkg/logger"
	"github.com/gin-gonic/gin"
)

func InitWebEngine(l logger.Logger, userHdl *web.UserHandler) *gin.Engine {
	engine := gin.Default()
	ginx.SetLogger(l)
	userHdl.RegisterRoutes(engine)
	//wechatHdl.RegisterRoutes(engine)//, wechatHdl *web.OAuth2WechatHandler
	return engine
}

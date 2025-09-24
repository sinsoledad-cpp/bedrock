package ioc

import (
	"bedrock/internal/web"
	"github.com/gin-gonic/gin"
)

func InitWebEngine(userHdl *web.UserHandler) *gin.Engine {
	engine := gin.Default()
	userHdl.RegisterRoutes(engine)
	return engine
}

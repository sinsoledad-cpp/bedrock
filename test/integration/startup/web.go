package startup

import (
	"bedrock/internal/web"

	"github.com/gin-gonic/gin"
)

func InitGinServer(hdl *web.UserHandler) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	server := gin.Default()
	hdl.RegisterRoutes(server)
	return server
}

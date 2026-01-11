package startup

import (
	"bedrock/internal/web"
	"bedrock/internal/web/middleware"
	jwtware "bedrock/internal/web/middleware/jwt"

	"github.com/gin-gonic/gin"
)

func InitGinServer(hdl *web.UserHandler, jwtHdl jwtware.Handler) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	server := gin.Default()
	m := middleware.NewJWTAuth(jwtHdl)
	server.Use(m.Middleware())
	hdl.RegisterRoutes(server)
	return server
}

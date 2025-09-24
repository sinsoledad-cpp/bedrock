package ginx

import (
	"bedrock/pkg/logger"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"net/http"
)

var log = logger.NewNopLogger()

func SetLogger(l logger.Logger) {
	log = l
}

// WrapBodyAndClaims bizFn 就是你的业务逻辑
func WrapBodyAndClaims[Req any, Claims jwt.Claims](bizFn func(ctx *gin.Context, req Req, uc Claims) (Result, error)) gin.HandlerFunc {
	return func(ctx *gin.Context) {

		var req Req
		if err := ctx.Bind(&req); err != nil {
			log.Error("输入错误", logger.Error(err))
			return
		}
		log.Debug("输入参数", logger.Field{Key: "req:=", Val: req})

		val, ok := ctx.Get("user")
		if !ok {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		uc, ok := val.(Claims)
		if !ok {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		res, err := bizFn(ctx, req, uc)
		if err != nil {
			log.Error("执行业务逻辑失败", logger.Error(err))
		}
		log.Debug("返回响应", logger.Field{Key: "res:=", Val: res})

		ctx.JSON(http.StatusOK, res)
	}
}

func Wrap(bizFn func(ctx *gin.Context) (Result, error)) gin.HandlerFunc {
	return func(ctx *gin.Context) {

		res, err := bizFn(ctx)
		if err != nil {
			log.Error("执行业务逻辑失败", logger.Error(err))
		}
		log.Debug("返回响应", logger.Field{Key: "res", Val: res})

		ctx.JSON(http.StatusOK, res)
	}
}

func WrapBody[Req any](bizFn func(ctx *gin.Context, req Req) (Result, error)) gin.HandlerFunc {
	return func(ctx *gin.Context) {

		var req Req
		if err := ctx.Bind(&req); err != nil {
			log.Error("输入错误", logger.Error(err))
			return
		}
		log.Debug("输入参数", logger.Field{Key: "req:=", Val: req})

		res, err := bizFn(ctx, req)
		if err != nil {
			log.Error("执行业务逻辑失败", logger.Error(err))
		}
		log.Debug("返回响应", logger.Field{Key: "res:=", Val: res})

		ctx.JSON(http.StatusOK, res)
	}
}

func WrapClaims[Claims any](bizFn func(ctx *gin.Context, uc Claims) (Result, error)) gin.HandlerFunc {
	return func(ctx *gin.Context) {

		val, ok := ctx.Get("user")
		if !ok {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		uc, ok := val.(Claims)
		if !ok {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		res, err := bizFn(ctx, uc)
		if err != nil {
			log.Error("执行业务逻辑失败", logger.Error(err))
		}

		ctx.JSON(http.StatusOK, res)
	}
}

package web

import (
	"bedrock/internal/domain"
	"bedrock/internal/service"
	"bedrock/internal/web/errs"
	jwtware "bedrock/internal/web/middleware/jwt"
	"bedrock/pkg/ginx"
	"bedrock/pkg/logger"
	"errors"
	regexp "github.com/dlclark/regexp2"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"net/http"
)

var _ Handler = (*UserHandler)(nil)

const (
	emailRegexPattern    = "(?i)^[A-Z0-9_!#$%&'*+/=?`{|}~^.-]+@[A-Z0-9.-]+$"
	passwordRegexPattern = `^(?=.*[a-z])(?=.*[A-Z])(?=.*[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?~]).{8,}$`
	bizLogin             = "login"
)

type UserHandler struct {
	log              logger.Logger
	userSvc          service.UserService
	jwtware          jwtware.Handler
	emailRegexExp    *regexp.Regexp
	passwordRegexExp *regexp.Regexp
}

func NewUserHandler(log logger.Logger, userSvc service.UserService) *UserHandler {
	return &UserHandler{
		log:              log,
		userSvc:          userSvc,
		emailRegexExp:    regexp.MustCompile(emailRegexPattern, regexp.None),
		passwordRegexExp: regexp.MustCompile(passwordRegexPattern, regexp.None),
	}
}

func (u *UserHandler) RegisterRoutes(e *gin.Engine) {
	g := e.Group("/user")
	g.POST("/ping", u.Ping)
	g.POST("/signup", ginx.WrapBody(u.SignUp))
	g.POST("/login", ginx.WrapBody(u.LoginJWT))
	g.POST("/refresh_token", ginx.Wrap(u.RefreshToken))
}
func (u *UserHandler) Ping(ctx *gin.Context) {
	ctx.String(http.StatusOK, "ping pong")
}

type SignUpReq struct {
	Email           string `json:"email"`
	Password        string `json:"password"`
	ConfirmPassword string `json:"confirmPassword"`
}

func (u *UserHandler) SignUp(ctx *gin.Context, req SignUpReq) (ginx.Result, error) {
	// 校验客户端输入
	isEmail, err := u.emailRegexExp.MatchString(req.Email)
	if err != nil {
		return ginx.Result{
			Code: errs.UserInternalServerError,
			Msg:  "系统错误",
		}, err
	}
	if !isEmail {
		return ginx.Result{
			Code: errs.UserInvalidInput,
			Msg:  "邮箱格式错误",
		}, nil
	}
	if req.Password != req.ConfirmPassword {
		return ginx.Result{
			Code: errs.UserInvalidInput,
			Msg:  "两次输入密码不对",
		}, nil
	}
	isPassword, err := u.passwordRegexExp.MatchString(req.Password)
	if err != nil {
		return ginx.Result{
			Code: errs.UserInvalidInput,
			Msg:  "系统错误",
		}, err
	}
	if !isPassword {
		return ginx.Result{
			Code: errs.UserInvalidInput,
			Msg:  "密码必须包含数字、特殊字符、大小字母，并且长度不能小于 8 位",
		}, nil
	}

	// 业务逻辑
	err = u.userSvc.Signup(ctx.Request.Context(), domain.User{Email: req.Email, Password: req.ConfirmPassword})
	if errors.Is(err, service.ErrDuplicateEmail) {
		return ginx.Result{
			Code: errs.UserDuplicateEmail,
			Msg:  "邮箱冲突",
		}, err
	}
	if err != nil {
		return ginx.Result{
			Code: errs.UserInternalServerError,
			Msg:  "系统错误",
		}, err
	}

	return ginx.Result{
		Code: http.StatusCreated,
		Msg:  "注册成功",
	}, nil
}

type LoginJWTReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (u *UserHandler) LoginJWT(ctx *gin.Context, req LoginJWTReq) (ginx.Result, error) {
	user, err := u.userSvc.Login(ctx, req.Email, req.Password)
	switch {
	case err == nil:
		err = u.jwtware.SetLoginToken(ctx, user.ID)
		if err != nil {
			return ginx.Result{
				Code: errs.UserInternalServerError,
				Msg:  "系统错误",
			}, err
		}
		return ginx.Result{
			Code: http.StatusOK,
			Msg:  "登录成功",
		}, nil
	case errors.Is(err, service.ErrInvalidUserOrPassword):
		return ginx.Result{
			Code: errs.UserInvalidOrPassword,
			Msg:  "用户名或者密码错误",
		}, err
	default:
		return ginx.Result{
			Code: errs.UserInternalServerError,
			Msg:  "系统错误",
		}, err
	}
}

func (u *UserHandler) RefreshToken(ctx *gin.Context) (ginx.Result, error) {
	// 假定长 token 也放在这里
	tokenStr := ctx.GetHeader("X-Refresh-Token")

	var rc jwtware.RefreshClaims
	token, err := jwt.ParseWithClaims(tokenStr, &rc, func(token *jwt.Token) (interface{}, error) {
		return jwtware.RefreshTokenKey, nil
	})
	// 这边要保持和登录校验一直的逻辑，即返回 401 响应
	if err != nil || token == nil || !token.Valid {
		return ginx.Result{
			Code: http.StatusUnauthorized,
			Msg:  "登录已过期，请重新登录",
		}, err
	}

	// 校验 ssid
	err = u.jwtware.CheckSession(ctx, rc.Ssid)
	if err != nil {
		// 如果是会话不存在的业务错误，返回 401
		if errors.Is(err, jwtware.ErrSessionNotFound) {
			return ginx.Result{
				Code: http.StatusUnauthorized,
				Msg:  "会话已过期，请重新登录",
			}, err
		}
		// 系统错误或者用户已经主动退出登录了
		// 这里也可以考虑说，如果在 Redis 已经崩溃的时候，
		// 就不要去校验是不是已经主动退出登录了。
		//ctx.AbortWithStatus(http.StatusUnauthorized)
		return ginx.Result{
			Code: http.StatusInternalServerError,
			Msg:  "系统错误",
		}, err
	}

	err = u.jwtware.SetJWTToken(ctx, rc.Uid, rc.Ssid)
	if err != nil {
		return ginx.Result{
			Code: http.StatusUnauthorized,
			Msg:  "系统内部错误",
		}, err
	}
	return ginx.Result{
		Code: http.StatusOK,
		Msg:  "刷新成功",
	}, nil
}

package web

import (
	"bedrock/internal/domain"
	"bedrock/internal/service"
	"bedrock/internal/web/errs"
	"bedrock/pkg/ginx"
	"bedrock/pkg/logger"
	"errors"
	regexp "github.com/dlclark/regexp2"
	"github.com/gin-gonic/gin"
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

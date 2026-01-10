package web

import (
	"bedrock/internal/domain"
	"bedrock/internal/service"
	svcmocks "bedrock/internal/service/mocks"
	"bedrock/internal/web/errs"
	jwtware "bedrock/internal/web/middleware/jwt"
	"bedrock/pkg/ginx"
	"bedrock/pkg/logger"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestUserHandler_SignUp(t *testing.T) {
	testCases := []struct {
		name string
		mock func(ctrl *gomock.Controller) service.UserService
		req  SignUpReq

		wantResult ginx.Result
		wantErr    error
	}{
		{
			name: "注册成功",
			mock: func(ctrl *gomock.Controller) service.UserService {
				svc := svcmocks.NewMockUserService(ctrl)
				svc.EXPECT().Signup(gomock.Any(), domain.User{
					Email:    "test@example.com",
					Password: "Password123!",
				}).Return(nil)
				return svc
			},
			req: SignUpReq{
				Email:           "test@example.com",
				Password:        "Password123!",
				ConfirmPassword: "Password123!",
			},
			wantResult: ginx.Result{
				Code: http.StatusCreated,
				Msg:  "注册成功",
			},
			wantErr: nil,
		},
		{
			name: "两次输入密码不一致",
			mock: func(ctrl *gomock.Controller) service.UserService {
				svc := svcmocks.NewMockUserService(ctrl)
				return svc
			},
			req: SignUpReq{
				Email:           "test@example.com",
				Password:        "Password123!",
				ConfirmPassword: "Password1234!",
			},
			wantResult: ginx.Result{
				Code: errs.UserInvalidInput,
				Msg:  "两次输入密码不同",
			},
			wantErr: nil,
		},
		{
			name: "邮箱格式错误",
			mock: func(ctrl *gomock.Controller) service.UserService {
				svc := svcmocks.NewMockUserService(ctrl)
				return svc
			},
			req: SignUpReq{
				Email:           "invalid-email",
				Password:        "Password123!",
				ConfirmPassword: "Password123!",
			},
			wantResult: ginx.Result{
				Code: errs.UserInvalidInput,
				Msg:  "邮箱格式错误",
			},
			wantErr: nil,
		},
		{
			name: "密码格式错误",
			mock: func(ctrl *gomock.Controller) service.UserService {
				svc := svcmocks.NewMockUserService(ctrl)
				return svc
			},
			req: SignUpReq{
				Email:           "test@example.com",
				Password:        "123",
				ConfirmPassword: "123",
			},
			wantResult: ginx.Result{
				Code: errs.UserInvalidInput,
				Msg:  "密码必须包含数字、特殊字符、大小字母，并且长度不能小于 8 位",
			},
			wantErr: nil,
		},
		{
			name: "邮箱冲突",
			mock: func(ctrl *gomock.Controller) service.UserService {
				svc := svcmocks.NewMockUserService(ctrl)
				svc.EXPECT().Signup(gomock.Any(), domain.User{
					Email:    "test@example.com",
					Password: "Password123!",
				}).Return(service.ErrDuplicateEmail)
				return svc
			},
			req: SignUpReq{
				Email:           "test@example.com",
				Password:        "Password123!",
				ConfirmPassword: "Password123!",
			},
			wantResult: ginx.Result{
				Code: errs.UserDuplicateEmail,
				Msg:  "邮箱冲突",
			},
			wantErr: service.ErrDuplicateEmail,
		},
		{
			name: "系统错误",
			mock: func(ctrl *gomock.Controller) service.UserService {
				svc := svcmocks.NewMockUserService(ctrl)
				svc.EXPECT().Signup(gomock.Any(), domain.User{
					Email:    "test@example.com",
					Password: "Password123!",
				}).Return(errors.New("service error"))
				return svc
			},
			req: SignUpReq{
				Email:           "test@example.com",
				Password:        "Password123!",
				ConfirmPassword: "Password123!",
			},
			wantResult: ginx.Result{
				Code: errs.UserInternalServerError,
				Msg:  "系统错误",
			},
			wantErr: errors.New("service error"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			svc := tc.mock(ctrl)
			// 使用 NewUserHandler 初始化，确保正则表达式等字段被正确初始化
			h := NewUserHandler(logger.NewNopLogger(), svc, nil, nil, nil)

			// 构造 gin.Context
			ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
			// 必须设置 Request，因为 h.SignUp 用到了 ctx.Request.Context()
			ctx.Request = httptest.NewRequest("POST", "/users/signup", nil)

			// 调用被测方法
			res, err := h.SignUp(ctx, tc.req)

			assert.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.wantResult, res)
		})
	}
}

func TestUserHandler_Profile(t *testing.T) {
	testCases := []struct {
		name string
		mock func(ctrl *gomock.Controller) service.UserService
		uc   jwtware.UserClaims

		wantResult ginx.Result
		wantErr    error
	}{
		{
			name: "查询成功",
			mock: func(ctrl *gomock.Controller) service.UserService {
				svc := svcmocks.NewMockUserService(ctrl)
				svc.EXPECT().FindById(gomock.Any(), int64(123)).Return(domain.User{
					Nickname: "test_user",
					Email:    "test@example.com",
					AboutMe:  "I am a tester",
					Birthday: time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
					Avatar:   "avatar.jpg",
				}, nil)
				return svc
			},
			uc: jwtware.UserClaims{Uid: 123},
			wantResult: ginx.Result{
				Code: http.StatusOK,
				Msg:  "获取用户信息成功",
				Data: ProfileVO{
					Nickname: "test_user",
					Email:    "test@example.com",
					AboutMe:  "I am a tester",
					Birthday: "2000-01-01",
					Avatar:   "avatar.jpg",
				},
			},
			wantErr: nil,
		},
		{
			name: "查询失败",
			mock: func(ctrl *gomock.Controller) service.UserService {
				svc := svcmocks.NewMockUserService(ctrl)
				svc.EXPECT().FindById(gomock.Any(), int64(123)).Return(domain.User{}, errors.New("db error"))
				return svc
			},
			uc: jwtware.UserClaims{Uid: 123},
			wantResult: ginx.Result{
				Code: errs.UserInternalServerError,
				Msg:  "系统错误",
			},
			wantErr: errors.New("db error"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			svc := tc.mock(ctrl)
			h := NewUserHandler(logger.NewNopLogger(), svc, nil, nil, nil)

			ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
			ctx.Request = httptest.NewRequest("GET", "/users/profile", nil)

			res, err := h.Profile(ctx, tc.uc)

			assert.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.wantResult, res)
		})
	}
}

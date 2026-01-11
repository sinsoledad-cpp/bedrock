package integration

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"bedrock/internal/repository/dao"
	"bedrock/internal/web"
	"bedrock/internal/web/errs"
	"bedrock/test/integration/startup"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

type UserTestSuite struct {
	suite.Suite
	db     *gorm.DB
	rdb    redis.Cmdable
	server *gin.Engine
}

func (s *UserTestSuite) SetupSuite() {
	s.db = startup.InitMySQL()
	s.rdb = startup.InitRedis()
	s.server = startup.InitWebServer()

	// 初始化表结构
	if err := dao.InitTables(s.db); err != nil {
		s.T().Fatal(err)
	}
}

func (s *UserTestSuite) TearDownTest() {
	// 每个测试用例执行后清空数据
	s.db.Exec("TRUNCATE TABLE users")
	s.rdb.FlushDB(context.Background())
}

func (s *UserTestSuite) TestSignUp() {
	testCases := []struct {
		name     string
		before   func(t *testing.T)
		after    func(t *testing.T)
		req      web.SignUpReq
		wantCode int
		wantMsg  string
	}{
		{
			name: "注册成功",
			req: web.SignUpReq{
				Email:           "test@example.com",
				Password:        "Password123!",
				ConfirmPassword: "Password123!",
			},
			wantCode: 201,
			wantMsg:  "注册成功",
		},
		{
			name: "邮箱格式错误",
			req: web.SignUpReq{
				Email:           "invalid-email",
				Password:        "Password123!",
				ConfirmPassword: "Password123!",
			},
			wantCode: 400,
			wantMsg:  "输入参数有误，请检查",
		},
		{
			name: "两次密码不一致",
			req: web.SignUpReq{
				Email:           "test2@example.com",
				Password:        "Password123!",
				ConfirmPassword: "Password1234!",
			},
			wantCode: 400,
			wantMsg:  "输入参数有误，请检查",
		},
		{
			name: "重复注册",
			before: func(t *testing.T) {
				// 先插入一条记录
				user := dao.User{
					Email:    sql.NullString{String: "duplicate@example.com", Valid: true},
					Password: "hashed_password",
					Ctime:    time.Now().UnixMilli(),
					Utime:    time.Now().UnixMilli(),
				}
				err := s.db.Create(&user).Error
				assert.NoError(t, err)
			},
			req: web.SignUpReq{
				Email:           "duplicate@example.com",
				Password:        "Password123!",
				ConfirmPassword: "Password123!",
			},
			wantCode: errs.UserDuplicateEmail,
			wantMsg:  "邮箱冲突",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			if tc.before != nil {
				tc.before(s.T())
			}

			body, _ := json.Marshal(tc.req)
			req, _ := http.NewRequest(http.MethodPost, "/users/signup", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			recorder := httptest.NewRecorder()

			s.server.ServeHTTP(recorder, req)

			// 解析响应
			// 假设响应结构是 ginx.Result
			var res struct {
				Code int    `json:"code"`
				Msg  string `json:"msg"`
				Data any    `json:"data"`
			}
			err := json.Unmarshal(recorder.Body.Bytes(), &res)
			assert.NoError(s.T(), err)

			assert.Equal(s.T(), tc.wantCode, res.Code)
			if tc.wantMsg != "" {
				assert.Equal(s.T(), tc.wantMsg, res.Msg)
			}

			if tc.after != nil {
				tc.after(s.T())
			}
		})
	}
}

func TestUser(t *testing.T) {
	suite.Run(t, new(UserTestSuite))
}

package integration

import (
	"bytes"
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
	// 注意：并行测试时，不能直接 TRUNCATE TABLE 或 FlushDB，否则会影响其他正在运行的测试
	// 应该在每个 case 的 after 中清理自己的数据
	// s.db.Exec("TRUNCATE TABLE users")
	// s.rdb.FlushDB(context.Background())
}

func (s *UserTestSuite) TestSignUp() {
	t := s.T()
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
			before: func(t *testing.T) {
				// 确保数据不存在
				s.db.Where("email = ?", "test_signup_success@example.com").Delete(&dao.User{})
			},
			after: func(t *testing.T) {
				var user dao.User
				if err := s.db.Where("email = ?", "test_signup_success@example.com").First(&user).Error; err != nil {
					t.Fatal(err)
				}
				assert.Equal(t, "test_signup_success@example.com", user.Email.String)
				// 清理数据
				s.db.Where("email = ?", "test_signup_success@example.com").Delete(&dao.User{})
			},
			req: web.SignUpReq{
				Email:           "test_signup_success@example.com",
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
			after: func(t *testing.T) {
				// 清理数据
				s.db.Where("email = ?", "duplicate@example.com").Delete(&dao.User{})
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
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if tc.before != nil {
				tc.before(t)
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
			assert.NoError(t, err)

			assert.Equal(t, tc.wantCode, res.Code)
			if tc.wantMsg != "" {
				assert.Equal(t, tc.wantMsg, res.Msg)
			}

			if tc.after != nil {
				tc.after(t)
			}
		})
	}
}

func (s *UserTestSuite) TestLogin() {
	t := s.T()

	testCases := []struct {
		name     string
		before   func(t *testing.T)
		after    func(t *testing.T)
		req      web.LoginJWTReq
		wantCode int
		wantMsg  string
	}{
		{
			name: "登录成功",
			before: func(t *testing.T) {
				// 先创建一个用户
				// 密码是 Password123!
				password := "$2a$10$cP8aKcJCsuzC1aM7gZ26IuD80XS7Viol0yx2UpASiY0afl10n3ZQe"
				now := time.Now().UnixMilli()
				err := s.db.Create(&dao.User{
					Email:    sql.NullString{String: "test_login_success@example.com", Valid: true},
					Password: password,
					Ctime:    now,
					Utime:    now,
				}).Error
				if err != nil {
					t.Fatal(err)
				}
			},
			after: func(t *testing.T) {
				s.db.Where("email = ?", "test_login_success@example.com").Delete(&dao.User{})
			},
			req: web.LoginJWTReq{
				Email:    "test_login_success@example.com",
				Password: "Password123!",
			},
			wantCode: 200,
			wantMsg:  "登录成功",
		},
		{
			name: "用户不存在",
			before: func(t *testing.T) {
			},
			after: func(t *testing.T) {
			},
			req: web.LoginJWTReq{
				Email:    "test_login_not_found@example.com",
				Password: "Password123!",
			},
			wantCode: errs.UserInvalidOrPassword,
			wantMsg:  "用户名或者密码错误",
		},
		{
			name: "密码错误",
			before: func(t *testing.T) {
				password := "$2a$10$cP8aKcJCsuzC1aM7gZ26IuD80XS7Viol0yx2UpASiY0afl10n3ZQe"
				now := time.Now().UnixMilli()
				err := s.db.Create(&dao.User{
					Email:    sql.NullString{String: "test_login_wrong_password@example.com", Valid: true},
					Password: password,
					Ctime:    now,
					Utime:    now,
				}).Error
				if err != nil {
					t.Fatal(err)
				}
			},
			after: func(t *testing.T) {
				s.db.Where("email = ?", "test_login_wrong_password@example.com").Delete(&dao.User{})
			},
			req: web.LoginJWTReq{
				Email:    "test_login_wrong_password@example.com",
				Password: "WrongPassword!",
			},
			wantCode: errs.UserInvalidOrPassword,
			wantMsg:  "用户名或者密码错误",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if tc.before != nil {
				tc.before(t)
			}

			reqBody, err := json.Marshal(tc.req)
			if err != nil {
				t.Fatal(err)
			}
			req, err := http.NewRequest(http.MethodPost, "/users/login", bytes.NewBuffer(reqBody))
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			s.server.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)
			var res map[string]interface{}
			err = json.Unmarshal(w.Body.Bytes(), &res)
			if err != nil {
				t.Fatal(err)
			}
			assert.Equal(t, float64(tc.wantCode), res["code"])
			assert.Equal(t, tc.wantMsg, res["msg"])

			if tc.after != nil {
				tc.after(t)
			}
		})
	}
}

func (s *UserTestSuite) TestEdit() {
	t := s.T()

	testCases := []struct {
		name     string
		before   func(t *testing.T) (string, int64) // 返回 token 和 uid
		after    func(t *testing.T, uid int64)
		req      web.UserEditReq
		wantCode int
		wantMsg  string
	}{
		{
			name: "修改成功",
			before: func(t *testing.T) (string, int64) {
				// 确保数据不存在
				s.db.Where("email = ?", "test_edit_success@example.com").Delete(&dao.User{})

				// 创建用户并登录获取 token
				email := "test_edit_success@example.com"
				uid := s.createUser(t, email)
				token := s.login(t, uid)
				return token, uid
			},
			after: func(t *testing.T, uid int64) {
				var user dao.User
				if err := s.db.Where("id = ?", uid).First(&user).Error; err != nil {
					t.Fatal(err)
				}
				assert.Equal(t, "new_nickname", user.Nickname)
				assert.Equal(t, "new_about_me", user.AboutMe)
				// 验证生日是否正确更新（注意时区问题，这里简化验证）

				// 清理数据
				s.db.Where("email = ?", "test_edit_success@example.com").Delete(&dao.User{})
			},
			req: web.UserEditReq{
				Nickname: "new_nickname",
				AboutMe:  "new_about_me",
				Birthday: "2000-01-01",
			},
			wantCode: 200,
			wantMsg:  "上传成功",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var token string
			var uid int64
			if tc.before != nil {
				token, uid = tc.before(t)
			}

			reqBody, err := json.Marshal(tc.req)
			if err != nil {
				t.Fatal(err)
			}
			req, err := http.NewRequest(http.MethodPost, "/users/edit", bytes.NewBuffer(reqBody))
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", token)

			w := httptest.NewRecorder()
			s.server.ServeHTTP(w, req)

			// 调试日志
			if w.Code != tc.wantCode {
				t.Logf("Expected code %d, got %d. Body: %s", tc.wantCode, w.Code, w.Body.String())
			}

			assert.Equal(t, http.StatusOK, w.Code)
			var res map[string]interface{}
			err = json.Unmarshal(w.Body.Bytes(), &res)
			if err != nil {
				t.Fatal(err)
			}
			assert.Equal(t, float64(tc.wantCode), res["code"])
			assert.Equal(t, tc.wantMsg, res["msg"])

			if tc.after != nil {
				tc.after(t, uid)
			}
		})
	}
}

func (s *UserTestSuite) TestProfile() {
	t := s.T()

	testCases := []struct {
		name     string
		before   func(t *testing.T) (string, int64)
		after    func(t *testing.T, uid int64)
		wantCode int
		wantMsg  string
		wantData web.ProfileVO
	}{
		{
			name: "获取成功",
			before: func(t *testing.T) (string, int64) {
				email := "test_profile_success@example.com"
				uid := s.createUser(t, email)
				// 更新一下其他信息以便验证
				s.db.Model(&dao.User{}).Where("id = ?", uid).Updates(map[string]interface{}{
					"nickname": "test_nick",
					"about_me": "test_about",
				})
				token := s.login(t, uid)
				return token, uid
			},
			after: func(t *testing.T, uid int64) {
			},
			wantCode: 200,
			wantMsg:  "获取用户信息成功",
			wantData: web.ProfileVO{
				Email:    "test_profile_success@example.com",
				Nickname: "test_nick",
				AboutMe:  "test_about",
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var token string
			var uid int64
			if tc.before != nil {
				token, uid = tc.before(t)
			}

			req, err := http.NewRequest(http.MethodGet, "/users/profile", nil)
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Set("Authorization", token)

			w := httptest.NewRecorder()
			s.server.ServeHTTP(w, req)

			// 调试日志
			if w.Code != tc.wantCode {
				t.Logf("Expected code %d, got %d. Body: %s", tc.wantCode, w.Code, w.Body.String())
			}

			assert.Equal(t, http.StatusOK, w.Code)
			var res struct {
				Code int           `json:"code"`
				Msg  string        `json:"msg"`
				Data web.ProfileVO `json:"data"`
			}
			err = json.Unmarshal(w.Body.Bytes(), &res)
			if err != nil {
				t.Fatal(err)
			}
			assert.Equal(t, tc.wantCode, res.Code)
			assert.Equal(t, tc.wantMsg, res.Msg)
			// 验证部分字段
			assert.Equal(t, tc.wantData.Email, res.Data.Email)
			assert.Equal(t, tc.wantData.Nickname, res.Data.Nickname)
			assert.Equal(t, tc.wantData.AboutMe, res.Data.AboutMe)

			if tc.after != nil {
				tc.after(t, uid)
			}
		})
	}
}

// 辅助方法：创建用户
func (s *UserTestSuite) createUser(t *testing.T, email string) int64 {
	// 确保数据不存在
	s.db.Where("email = ?", email).Delete(&dao.User{})

	user := dao.User{
		Email:    sql.NullString{String: email, Valid: true},
		Password: "$2a$10$cP8aKcJCsuzC1aM7gZ26IuD80XS7Viol0yx2UpASiY0afl10n3ZQe", // Password123!
		Ctime:    time.Now().UnixMilli(),
		Utime:    time.Now().UnixMilli(),
	}
	if err := s.db.Create(&user).Error; err != nil {
		t.Fatal(err)
	}
	return user.ID
}

// 辅助方法：登录获取 Token
func (s *UserTestSuite) login(t *testing.T, uid int64) string {
	// 获取用户的 email
	var user dao.User
	if err := s.db.First(&user, uid).Error; err != nil {
		t.Fatal(err)
	}

	reqBody := map[string]string{
		"email":    user.Email.String,
		"password": "Password123!",
	}
	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest(http.MethodPost, "/users/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.server.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Fatal("Login failed", w.Body.String())
	}
	token := w.Header().Get("x-jwt-token")
	if token == "" {
		t.Fatal("Token not found in header")
	}
	// 返回 Bearer token
	return "Bearer " + token
}

func TestUser(t *testing.T) {
	suite.Run(t, new(UserTestSuite))
}

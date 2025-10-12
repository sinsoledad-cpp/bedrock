# Bedrock - 基于Gin的单体通用后端开发框架

[![Go Version](https://img.shields.io/badge/Go-1.25+-blue.svg)](https://golang.org)
[![Gin Framework](https://img.shields.io/badge/Gin-v1.10.1-green.svg)](https://gin-gonic.com)
[![License](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

## 📖 项目简介

Bedrock 是一个基于 Gin 框架构建的单体通用后端开发框架，采用现代化的架构设计，集成了用户认证、短信服务、缓存、数据库操作等常用功能模块。项目采用清晰的分层架构，支持依赖注入，适合快速开发企业级应用。

## ✨ 特性

- 🚀 **高性能** - 基于 Gin 框架，提供高性能的 HTTP 服务
- 🔐 **完整认证** - JWT 认证、会话管理、Token 刷新机制
- 📱 **短信服务** - 支持阿里云、腾讯云短信服务，可扩展其他服务商
- 💾 **数据持久化** - MySQL + GORM，Redis 缓存支持
- 🔄 **依赖注入** - 使用 Google Wire 实现依赖注入
- 📊 **监控支持** - 集成 Prometheus 监控指标
- 🌐 **跨域支持** - 内置 CORS 中间件
- 🛡️ **安全防护** - 参数验证、限流、防重放攻击
- 📝 **日志管理** - 结构化日志，支持文件轮转
- 🔧 **配置管理** - 基于 Viper 的灵活配置管理

## 🏗️ 项目架构

```
bedrock/
├── cmd/                 # 应用入口
├── configs/            # 配置文件
├── internal/           # 内部模块（不对外暴露）
│   ├── domain/         # 领域模型
│   ├── repository/     # 数据访问层
│   │   ├── cache/      # 缓存实现
│   │   ├── dao/        # 数据访问对象
│   │   └── repository/ # 仓储接口
│   ├── service/        # 业务逻辑层
│   │   ├── sms/        # 短信服务
│   │   ├── oauth2/     # OAuth2 认证
│   │   └── code/       # 验证码服务
│   └── web/            # Web 层
│       ├── middleware/  # 中间件
│       └── handler/    # 请求处理器
├── ioc/                # 依赖注入容器
├── pkg/                # 可复用的公共包
│   ├── ginx/           # Gin 扩展
│   ├── logger/         # 日志组件
│   ├── validate/       # 参数验证
│   └── captcha/        # 验证码组件
└── setting/            # 配置初始化
```

## 🚀 快速开始

### 环境要求

- Go 1.25+
- MySQL 8.0+
- Redis 6.0+
- Docker & Docker Compose (可选)

### 安装步骤

1. **克隆项目**
```bash
git clone <repository-url>
cd bedrock
```

2. **安装依赖**
```bash
go mod tidy
```

3. **启动基础设施**
```bash
# 使用 Docker Compose 启动 MySQL 和 Redis
docker-compose up -d
```

4. **配置应用**
```bash
# 复制配置文件模板
cp configs/dev.yaml.example configs/dev.yaml

# 编辑配置文件，设置数据库连接等信息
vim configs/dev.yaml
```

5. **运行应用**
```bash
# 开发模式运行
go run main.go

# 或者构建后运行
go build -o bedrock
./bedrock
```

### 配置文件示例

```yaml
# configs/dev.yaml
server:
  port: 8080

database:
  mysql:
    dsn: "root:root@tcp(localhost:3306)/bedrock?charset=utf8mb4&parseTime=True&loc=Local"
  redis:
    addr: "localhost:6379"

sms:
  provider: "tencent"  # 或 "aliyun"
  tencent:
    secret_id: "your-secret-id"
    secret_key: "your-secret-key"
    app_id: "your-app-id"
    sign_name: "your-sign-name"

jwt:
  access_token_key: "k6CswdUm77WKcbM68UQUuxVsHSpTCwgK"
  refresh_token_key: "k6CswdUm77WKcbM68UQUuxVsHSpTCwgA"

log:
  level: "info"
  path: "./logs/app.log"
```

## 📚 API 文档

### 用户相关接口

#### 用户注册
```http
POST /user/signup
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "Password123!",
  "confirmPassword": "Password123!"
}
```

#### 用户登录
```http
POST /user/login
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "Password123!"
}
```

#### 短信登录
```http
# 发送验证码
POST /user/login_sms/code/send
Content-Type: application/json

{
  "phone": "13800138000"
}

# 验证码登录
POST /user/login_sms
Content-Type: application/json

{
  "phone": "13800138000",
  "code": "123456"
}
```

#### 获取用户信息
```http
GET /user/profile
Authorization: Bearer <jwt-token>
```

#### 更新用户信息
```http
POST /user/edit
Authorization: Bearer <jwt-token>
Content-Type: application/json

{
  "nickname": "新昵称",
  "birthday": "1990-01-01",
  "aboutMe": "个人简介"
}
```

#### Token 刷新
```http
POST /user/refresh_token
X-Refresh-Token: <refresh-token>
```

#### 用户退出
```http
POST /user/logout
Authorization: Bearer <jwt-token>
```

### 响应格式

所有接口返回统一的 JSON 格式：

```json
{
  "code": 200,
  "msg": "操作成功",
  "data": {
    // 具体数据
  }
}
```

## 🔧 核心组件

### 1. 依赖注入 (IOC)

项目使用 Google Wire 实现依赖注入，确保组件间的松耦合：

```go
// wire.go 中定义依赖关系
var userSvc = wire.NewSet(
    cache.NewRedisUserCache,
    dao.NewGORMUserDAO,
    repository.NewCachedUserRepository,
    service.NewUserService,
)
```

### 2. 数据访问层

采用 Repository 模式，支持缓存：

```go
type UserRepository interface {
    Create(ctx context.Context, u domain.User) error
    FindByEmail(ctx context.Context, email string) (domain.User, error)
    FindById(ctx context.Context, id int64) (domain.User, error)
}
```

### 3. 业务逻辑层

清晰的业务服务划分：

- `UserService` - 用户相关业务逻辑
- `CodeService` - 验证码服务
- `SMSService` - 短信发送服务

### 4. Web 层

基于 Gin 的 Web 框架扩展：

```go
// 统一的响应包装器
func WrapBody[Req any](bizFn func(ctx *gin.Context, req Req) (Result, error)) gin.HandlerFunc
```

### 5. 中间件

- JWT 认证中间件
- CORS 跨域中间件
- 限流中间件
- 日志中间件

## 🛠️ 开发指南

### 添加新的 API 接口

1. **在 `internal/web/` 中添加处理器**
2. **在 `internal/service/` 中添加业务逻辑**
3. **在 `internal/repository/` 中添加数据访问**
4. **在 `wire.go` 中注册依赖**
5. **在路由中注册接口**

### 自定义短信服务商

实现 `sms.Service` 接口：

```go
type Service interface {
    Send(ctx context.Context, tpl string, args []string, numbers ...string) error
}
```

### 添加新的数据库表

1. **在 `internal/domain/` 中定义领域模型**
2. **在 `internal/repository/dao/` 中定义 DAO**
3. **在 `internal/repository/` 中定义 Repository**

## 📊 监控和日志

### 日志配置

项目使用 zap 日志库，支持结构化日志和文件轮转：

```go
// 初始化日志
logger.Init(&conf.LogConf{
    Level:      "info",
    Path:       "./logs/app.log",
    MaxSize:    100, // MB
    MaxBackups: 10,
    MaxAge:     30,  // days
}, "debug")
```

### Prometheus 监控

集成 Prometheus 监控指标，可通过 `/metrics` 端点访问。

## 🔒 安全特性

- JWT Token 自动刷新机制
- 会话管理，支持主动退出
- 短信验证码防刷机制
- 密码强度验证
- SQL 注入防护（GORM 参数化查询）
- XSS 防护

## 🐳 Docker 部署

### 使用 Docker Compose

```bash
# 启动所有服务
docker-compose up -d

# 查看服务状态
docker-compose ps

# 停止服务
docker-compose down
```

### 构建应用镜像

```dockerfile
FROM golang:1.25-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o bedrock .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/bedrock .
COPY configs/prod.yaml ./configs/

EXPOSE 8080
CMD ["./bedrock", "--config", "configs/prod.yaml"]
```

## 🤝 贡献指南

1. Fork 本项目
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 开启 Pull Request

## 📄 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情。

## 🙏 致谢

- [Gin](https://gin-gonic.com/) - 高性能 HTTP 框架
- [GORM](https://gorm.io/) - 优雅的 ORM 库
- [Wire](https://github.com/google/wire) - Go 依赖注入工具
- [Viper](https://github.com/spf13/viper) - 配置管理
- [Zap](https://github.com/uber-go/zap) - 高性能日志库

## 📞 联系方式

如有问题或建议，请通过以下方式联系：

- 提交 Issue
- 发送邮件

---

**Bedrock** - 为您的下一个项目奠定坚实的基础！ 🚀
# Sinx 用户认证系统

基于 Gin 框架的用户注册登录后端系统，采用分层架构设计。

## 项目架构

```text
sinx/
├── main.go                     # 应用入口
├── go.mod                      # Go模块依赖
├── .env                        # 环境变量配置
├── pkg/                        # 公共库
│   ├── config/                 # 配置管理
│   ├── logger/                 # 日志组件
│   ├── errorx/                 # 错误处理
│   ├── response/               # 统一响应
│   ├── auth/                   # JWT认证
│   └── utils/                  # 工具函数
├── domain/                     # 领域层
│   └── user/
│       ├── entity/             # 实体
│       ├── repository/         # 仓储接口
│       └── service/            # 领域服务
├── infra/                      # 基础设施层
│   ├── database/               # 数据库连接
│   ├── repository/             # 仓储实现
│   └── migration/              # 数据库迁移
├── application/                # 应用服务层
│   ├── user/
│   │   ├── dto/               # 数据传输对象
│   │   └── service/           # 应用服务
│   └── application.go          # 应用初始化
└── api/                        # 传输层
    ├── handler/                # 处理器
    ├── middleware/             # 中间件
    └── router/                 # 路由
```

## 技术栈

- **Go 1.24**
- **Gin** - HTTP框架
- **GORM** - ORM框架
- **PostgreSQL** - 数据库
- **JWT** - 身份认证
- **Zap** - 日志框架
- **bcrypt** - 密码加密

## 功能特性

- 用户注册
- 用户登录
- JWT身份认证
- 用户资料查询
- 分层架构设计
- 统一错误处理
- 结构化日志
- 数据库自动迁移
- Swagger API 文档（/swagger/index.html）

## 快速开始

### 1. 环境准备

确保已安装：

- Go 1.24+
- PostgreSQL

### 2. 数据库配置

创建PostgreSQL数据库：

```sql
CREATE DATABASE sinx;
```

### 3. 环境变量配置

复制并修改 `.env` 文件中的数据库配置（注意：DB_NAME 是数据库名称，与 Go Module 路径 github.com/sine-io/sinx 无关）：

```bash
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=123456
DB_NAME=sinx
JWT_ISSUER=github.com/sine-io/sinx
```

### 4. 安装依赖

```bash
go mod tidy
```

### 5. 运行应用

```bash
go run main.go
```

应用将在 `http://localhost:8080` 启动。

### 6. 生成并查看 Swagger 文档

第一次需要安装 swag CLI：

```bash
go install github.com/swaggo/swag/cmd/swag@latest
```

在项目根目录生成文档：

```bash
swag init -g main.go -o docs
```

启动后访问：

```text
http://localhost:8080/swagger/index.html
```

## API接口

### 用户注册

```http
POST /api/auth/register
Content-Type: application/json

{
    "username": "john_doe",
    "email": "john@example.com",
    "password": "password123"
}
```

### 用户登录

```http
POST /api/auth/login
Content-Type: application/json

{
    "username": "john_doe",
    "password": "password123"
}
```

### 获取用户资料

```http
GET /api/user/profile
Authorization: Bearer <token>
```

### 健康检查

```http
GET /health
```

## 错误码

| 错误码 | 说明 |
|-------|------|
| 0 | 成功 |
| 10001 | 内部服务器错误 |
| 10002 | 参数错误 |
| 10003 | 未认证 |
| 20001 | 用户不存在 |
| 20002 | 用户已存在 |
| 20003 | 密码错误 |
| 20004 | 令牌无效 |

## 开发指南

### 分层原则

1. **API层**：处理HTTP请求，参数验证，调用应用服务
2. **应用服务层**：用例编排，事务边界，调用领域服务
3. **领域层**：业务逻辑，实体，领域服务
4. **基础设施层**：数据持久化，外部接口

### 添加新功能

1. 在 `domain/` 中定义实体和领域服务
2. 在 `infra/` 中实现仓储
3. 在 `application/` 中编写应用服务
4. 在 `api/` 中添加处理器和路由

### 数据库迁移

使用GORM的AutoMigrate功能：

```go
// 在 infra/migration/migration.go 中添加新的模型
err := db.AutoMigrate(
    &entity.User{},
    &entity.NewEntity{}, // 新增实体
)
```

## 配置说明

主要环境变量：

| 变量名 | 说明 | 默认值 |
|--------|------|--------|
| APP_ENV | 运行环境 | development |
| LISTEN_ADDR | 监听地址 | :8080 |
| LOG_LEVEL | 日志级别 | info |
| DB_* | 数据库配置 | - |
| JWT_SECRET | JWT密钥 | - |
| JWT_EXPIRE_HOURS | JWT过期时间(小时) | 24 |
| JWT_ISSUER | JWT签发者 | github.com/sine-io/sinx |

## 部署

### Docker部署

```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod tidy
RUN go build -o sinx main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/sinx .
COPY --from=builder /app/.env .
CMD ["./sinx"]
```

### 生产环境注意事项

1. 修改JWT密钥
2. 设置APP_ENV=production
3. 配置HTTPS
4. 使用连接池
5. 监控日志

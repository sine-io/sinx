# Sinx 用户认证 / RBAC 权限管理系统

基于 Gin + GORM + PostgreSQL 构建的用户认证与 **RBAC(基于角色的访问控制)** 系统，采用整洁分层 / DDD 风格（传输层 / 应用层 / 领域层 / 基础设施层），内置统一响应、结构化日志、权限点集中管理、菜单-角色-用户关联、Swagger 文档以及简单的权限中间件。

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

- **Go 1.21+ / 1.24 兼容**
- **Gin** - HTTP框架
- **GORM** - ORM框架
- **PostgreSQL** - 数据库
- **JWT** - 身份认证
- **Zap** - 日志框架
- **bcrypt** - 密码加密

## 功能特性

- 用户注册 / 登录 / 个人资料
- JWT 身份认证（HS256）
- RBAC：用户-角色-菜单-权限点
- 角色绑定菜单、用户绑定角色
- 动态权限校验中间件（按权限点）
- 菜单树 / 角色菜单树 / 用户菜单树
- 全量权限导出接口（便于前端动态渲染）
- 分页列表：用户 / 角色 / 菜单
- 统一错误码与响应包装
- 结构化 Zap 日志、恢复 & CORS 中间件
- 数据库自动迁移
- Swagger API 文档（/swagger/index.html）
- Docker / docker-compose 一键启动

## 快速开始

### 1. 环境准备

确保已安装：

- Go 1.24+
- PostgreSQL

### 2. 数据库配置（本地）

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

### 4. （可选）使用 Docker Compose 快速运行

```bash
docker compose up -d --build
```

启动后：

- 应用：<http://localhost:8080>
- PostgreSQL：localhost:5432 （用户 postgres / 123456）

### 5. 安装依赖（本地开发）

```bash
go mod tidy
```

### 6. 运行应用

```bash
go run main.go
```

应用将在 `http://localhost:8080` 启动。

### 7. 生成并查看 Swagger 文档

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

## 权限模型简述

核心关系：

```text
User <-> UserRole(关联表) <-> Role <-> RoleMenu(关联表) <-> Menu (含 perms 权限点)
```

权限校验流程：

1. 用户登录获取 JWT
2. 请求受保护接口时 `AuthMiddleware` 解析用户 ID
3. `PermissionMiddleware` 根据用户 ID 计算并缓存其权限集合（当前实现为实时查询，可扩展 Redis）
4. 判断是否包含所需权限字符串（如 `user:list`）

`pkg/permissions/perms.go` 集中定义全部权限常量，并由 `/api/perms/all` 对外返回，便于前端生成动态路由或按钮显隐。

## 主要权限点

| 类别 | 权限点 |
| ---- | ------ |
| 用户 | user:create / user:list / user:update / user:delete / user:bindRole / user:unbindRole / user:roles |
| 角色 | role:create / role:list / role:update / role:delete / role:bindMenu / role:unbindMenu / role:menus / role:users |
| 菜单 | menu:create / menu:list / menu:update / menu:delete / menu:roles / menu:roleMenuTree |

## API 接口（节选）

### 认证相关

#### 用户注册

```http
POST /api/auth/register
Content-Type: application/json

{
    "username": "john_doe",
    "email": "john@example.com",
    "password": "password123"
}
```

#### 用户登录

```http
POST /api/auth/login
Content-Type: application/json

{
    "username": "john_doe",
    "password": "password123"
}
```

#### 获取用户资料

```http
GET /api/user/profile
Authorization: Bearer <token>
```

#### 健康检查

```http
GET /health
```

### 用户 / 角色 / 菜单 / 绑定相关（更多示例见 `API_TEST.md`）

| 功能 | 方法 | 路径 | 权限 | 说明 |
| ---- | ---- | ---- | ---- | ---- |
| 创建用户 | POST | /api/user/create | user:create | 管理员创建后台用户 |
| 用户列表 | GET | /api/user/list | user:list | 分页查询 |
| 更新用户 | POST | /api/user/update | user:update | 修改昵称/状态等 |
| 删除用户 | POST | /api/user/delete | user:delete | 逻辑删除 |
| 修改密码 | POST | /api/user/changePassword | 需登录 | 用户自改密码 |
| 绑定角色 | POST | /api/user/bindRole | user:bindRole | 批量绑定 |
| 解绑角色 | POST | /api/user/unbindRole | user:unbindRole | 批量解绑 |
| 用户角色 | GET | /api/user/roles?id=1 | user:roles | 列出角色 |
| 用户菜单树 | GET | /api/user/menus | 登录 | 动态菜单 |
| 创建角色 | POST | /api/role/create | role:create | 新增或更新 |
| 角色列表 | GET | /api/role/list | role:list | 分页查询 |
| 删除角色 | POST | /api/role/delete | role:delete | 删除 |
| 绑定菜单 | POST | /api/role/bindMenu | role:bindMenu | 批量 |
| 角色菜单 | GET | /api/role/menus?id=1 | role:menus | 列表 |
| 角色菜单树ID | GET | /api/menu/roleMenuTree?roleId=1 | menu:roleMenuTree | ID集合 |
| 创建菜单 | POST | /api/menu/create | menu:create | 支持目录/按钮 |
| 菜单列表 | GET | /api/menu/list | menu:list | 支持模糊/状态 |
| 菜单树 | GET | /api/menu/tree | 登录 | 全量树 |
| 菜单角色 | GET | /api/menu/roles?menuId=1 | menu:roles | 反查角色 |
| 所有权限 | GET | /api/perms/all | 登录 | 全部权限点 |

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

## Curl 示例（简略）

```bash
# 登录获取 token
curl -X POST http://localhost:8080/api/auth/login -H "Content-Type: application/json" -d '{"username":"admin","password":"admin123"}'

# 携带 token 创建角色
curl -X POST http://localhost:8080/api/role/create -H "Authorization: Bearer <TOKEN>" -H "Content-Type: application/json" -d '{"name":"管理员","status":1}'
```

更多详尽请求与响应示例见 `API_TEST.md`。

## 部署

### Docker 部署

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

1. 修改 JWT_SECRET；设置足够熵
2. 设置 APP_ENV=production, LOG_LEVEL=info 或 warn
3. 前置反向代理 (Nginx / Traefik) + HTTPS
4. 数据库连接池与慢查询监控
5. 配置集中日志（ELK / Loki）
6. 结合 Redis 做权限/菜单缓存可显著降低查询压力
7. 定期滚动刷新用户权限缓存（或在绑定/解绑后主动失效）

## Roadmap / 建议增强

- 权限缓存实现（Redis）
- 审计日志（谁修改了角色/菜单）
- 乐观锁 / 软删除标志
- OpenTelemetry 链路追踪
- 单元与集成测试覆盖扩展
- CLI 初始化超管账号脚本

---

欢迎贡献 PR 或 Issue。

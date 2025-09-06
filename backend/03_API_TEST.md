# Sinx API 测试示例（含 RBAC）

## 环境要求

1. 确保PostgreSQL已安装并运行
2. 创建数据库：`CREATE DATABASE sinx;`
3. 修改 `.env` 文件中的数据库配置

## 启动服务

```bash
go run main.go
```

服务将在 `http://localhost:8080` 启动

> 建议：先注册或准备一个初始管理员账号，再通过数据库手动授予其必要的角色/菜单/权限（或直接在迁移阶段初始化）。

## API测试示例

### 1. 健康检查

```bash
curl -X GET http://localhost:8080/health
```

预期响应：

```json
{
  "status": "ok",
  "message": "Service is running"
}
```

### 2. 用户注册

```bash
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "john_doe",
    "email": "john@example.com",
    "password": "password123"
  }'
```

预期响应：

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 1,
    "username": "john_doe",
    "email": "john@example.com",
    "is_active": true
  }
}
```

### 3. 用户登录

```bash
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "john_doe",
    "password": "password123"
  }'
```

预期响应：

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "user": {
      "id": 1,
      "username": "john_doe",
      "email": "john@example.com",
      "is_active": true
    }
  }
}
```

### 4. 获取用户资料（需要认证）

```bash
curl -X GET http://localhost:8080/api/user/profile \
  -H "Authorization: Bearer YOUR_JWT_TOKEN_HERE"
```

预期响应：

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 1,
    "username": "john_doe",
    "email": "john@example.com",
    "is_active": true
  }
}
```

### 5. 用户管理 / 角色绑定

#### 5.1 创建后台用户 (需要 user:create)

```bash
curl -X POST http://localhost:8080/api/user/create \
  -H "Authorization: Bearer <TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "ops01",
    "password": "Password@123",
    "nickname": "运维01",
    "email": "ops01@example.com"
  }'
```

#### 5.2 用户列表 (user:list)

```bash
curl -X GET 'http://localhost:8080/api/user/list?pageNum=1&pageSize=10' \
  -H "Authorization: Bearer <TOKEN>"
```

#### 5.3 更新用户 (user:update)

```bash
curl -X POST http://localhost:8080/api/user/update \
  -H "Authorization: Bearer <TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{"id":2,"nickname":"新昵称","email":"new@example.com"}'
```

#### 5.4 修改密码（本人或具备权限）

```bash
curl -X POST http://localhost:8080/api/user/changePassword \
  -H "Authorization: Bearer <TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{"userId":2,"oldPassword":"Password@123","newPassword":"Password@456"}'
```

#### 5.5 绑定用户角色 (user:bindRole)

```bash
curl -X POST http://localhost:8080/api/user/bindRole \
  -H "Authorization: Bearer <TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{"userId":2,"roleIds":[1,2]}'
```

#### 5.6 用户角色 (user:roles)

```bash
curl -X GET 'http://localhost:8080/api/user/roles?id=2' -H "Authorization: Bearer <TOKEN>"
```

### 6. 角色管理

#### 6.1 创建角色 (role:create)

```bash
curl -X POST http://localhost:8080/api/role/create \
  -H "Authorization: Bearer <TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{"name":"管理员","status":1,"remark":"系统最高权限"}'
```

#### 6.2 角色列表 (role:list)

```bash
curl -X GET 'http://localhost:8080/api/role/list?pageNum=1&pageSize=10' -H "Authorization: Bearer <TOKEN>"
```

#### 6.3 绑定角色菜单 (role:bindMenu)

```bash
curl -X POST http://localhost:8080/api/role/bindMenu \
  -H "Authorization: Bearer <TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{"roleId":1,"menuIds":[1,2,3]}'
```

#### 6.4 角色菜单列表 (role:menus)

```bash
curl -X GET 'http://localhost:8080/api/role/menus?id=1' -H "Authorization: Bearer <TOKEN>"
```

### 7. 菜单管理

#### 7.1 创建菜单 (menu:create)

菜单类型示例：目录(menuType=CATALOG)、菜单(menuType=MENU)、按钮(menuType=BUTTON)。

```bash
curl -X POST http://localhost:8080/api/menu/create \
  -H "Authorization: Bearer <TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{
    "name":"系统管理",
    "parentId":0,
    "orderNum":1,
    "isFrame":0,
    "menuType":"CATALOG",
    "isCatch":0,
    "isHidden":0,
    "status":1,
    "remark":"系统模块根"
  }'
```

#### 7.2 菜单列表 (menu:list)

```bash
curl -X GET 'http://localhost:8080/api/menu/list?pageNum=1&pageSize=20' -H "Authorization: Bearer <TOKEN>"
```

#### 7.3 菜单树 (登录即可)

```bash
curl -X GET http://localhost:8080/api/menu/tree -H "Authorization: Bearer <TOKEN>"
```

#### 7.4 角色菜单树 ID 集合 (menu:roleMenuTree)

```bash
curl -X GET 'http://localhost:8080/api/menu/roleMenuTree?roleId=1' -H "Authorization: Bearer <TOKEN>"
```

### 8. 所有权限点

```bash
curl -X GET http://localhost:8080/api/perms/all -H "Authorization: Bearer <TOKEN>"
```

## 错误处理示例

### 用户已存在

```bash
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "john_doe",
    "email": "john@example.com",
    "password": "password123"
  }'
```

响应：

```json
{
  "code": 20002,
  "message": "user already exists"
}
```

### 无效密码

```bash
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "john_doe",
    "password": "wrongpassword"
  }'
```

响应：

```json
{
  "code": 20003,
  "message": "invalid password"
}
```

### 未认证访问

```bash
curl -X GET http://localhost:8080/api/user/profile
```

响应：

```json
{
  "code": 10003,
  "message": "unauthorized"
}
```

## 使用Postman测试

1. 导入以下环境变量：
   - `base_url`: `http://localhost:8080`
   - `token`: (通过登录接口获取)

2. 设置请求头：
   - `Content-Type`: `application/json`
   - `Authorization`: `Bearer {{token}}`

3. 测试建议流程：

- 注册或初始化管理员账户
- 登录获取 token
- 创建角色 -> 创建菜单 -> 绑定角色菜单 -> 创建普通用户 -> 绑定角色 -> 登录普通用户验证权限

## 常见排查

| 问题 | 可能原因 | 解决 |
| ---- | -------- | ---- |
| 401 unauthorized | Header 缺失或 token 过期 | 重新登录，确认使用 `Authorization: Bearer <token>` |
| 403 (自定义 code) | 没有绑定相应角色/菜单/权限点 | 确认角色已绑定菜单，菜单包含 perms 字段，且用户已绑定角色 |
| 列表为空 | 尚未创建数据或分页超出 | 检查 pageNum/pageSize 或数据库记录 |

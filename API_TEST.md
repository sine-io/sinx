# Sinx API 测试示例

## 环境要求

1. 确保PostgreSQL已安装并运行
2. 创建数据库：`CREATE DATABASE sinx;`
3. 修改 `.env` 文件中的数据库配置

## 启动服务

```bash
go run main.go
```

服务将在 `http://localhost:8080` 启动

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

3. 测试流程：
   - 先调用注册接口
   - 再调用登录接口获取token
   - 使用token访问需要认证的接口

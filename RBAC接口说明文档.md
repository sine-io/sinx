# **RBAC权限管理系统接口说明文档**

## **1. 概述**

本系统采用RBAC（Role-Based Access Control）基于角色的访问控制模型，实现了用户、角色、菜单（权限）的多对多关系管理。

## **2. 数据库表设计**

### **2.1 核心表结构**

#### **用户表 (user)**

```SQL
CREATE TABLE user (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    username VARCHAR(50) COMMENT '账号',
    password VARCHAR(100) NOT NULL COMMENT '密码',
    avatar VARCHAR(100) COMMENT '头像',
    nickname VARCHAR(50) COMMENT '昵称',
    user_type SMALLINT COMMENT '账号类型:0普通账号,1是超管',
    email VARCHAR(50) COMMENT '电邮地址',
    mobile VARCHAR(30) COMMENT '手机号码',
    sort INT DEFAULT 1 COMMENT '排序',
    status SMALLINT COMMENT '状态0是正常,1是禁用',
    last_login_ip VARCHAR(30) COMMENT '最后登录ip地址',
    last_login_nation VARCHAR(100) COMMENT '最后登录国家',
    last_login_province VARCHAR(100) COMMENT '最后登录省份',
    last_login_city VARCHAR(100) COMMENT '最后登录城市',
    last_login_date TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) COMMENT '最后登录时间',
    salt VARCHAR(30) COMMENT '密码盐',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL
);
```

#### **角色表 (role)**

```SQL
CREATE TABLE role (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(50) COMMENT '名称',
    remark VARCHAR(100) COMMENT '备注',
    status SMALLINT COMMENT '状态 0正常 1禁用',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL
);
```

#### **菜单表 (menu)**

```SQL
CREATE TABLE menu (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(50) COMMENT '菜单名称',
    parent_id INT COMMENT '父级菜单ID',
    order_num INT COMMENT '排序',
    path VARCHAR(100) COMMENT '路径',
    component VARCHAR(100) COMMENT '组件',
    query VARCHAR(100) COMMENT '参数',
    is_frame SMALLINT COMMENT '是否外链',
    menu_type VARCHAR(100) COMMENT '菜单类型 C目录 M菜单 B按钮',
    is_catch SMALLINT COMMENT '是否缓存',
    is_hidden SMALLINT COMMENT '是否可见',
    perms VARCHAR(100) COMMENT '权限标识',
    icon VARCHAR(100) COMMENT '图标',
    status SMALLINT COMMENT '状态 0正常 1禁用',
    remark VARCHAR(100) COMMENT '备注',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL
);
```

#### **用户角色关联表 (user_role)**

```SQL
CREATE TABLE user_role (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    user_id INT COMMENT '用户ID',
    role_id INT COMMENT '角色ID',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL
);
```

#### **角色菜单关联表 (role_menu)**

```SQL
CREATE TABLE role_menu (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    role_id INT COMMENT '角色ID',
    menu_id INT COMMENT '菜单ID',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL
);
```

### **2.2 关系说明**

- **用户与角色**: 多对多关系，通过 `user_role` 表关联
- **角色与菜单**: 多对多关系，通过 `role_menu` 表关联
- **用户与菜单**: 通过角色间接关联，用户通过拥有的角色获得菜单权限

## **3. 接口文档**

### **3.1 用户管理接口**

#### **3.1.1 创建用户**

- **接口路径**: `POST /api/user/create`
- **请求参数**:

```JSON
{
    "username": "string",     // 用户名，必填
    "password": "string",     // 密码，必填
    "nickname": "string",     // 昵称，必填
    "email": "string",        // 邮箱，可选
    "mobile": "string",       // 手机号，可选
    "avatar": "string"        // 头像URL，可选
}
```

- **响应示例**:

```JSON
{
    "code": 0,
    "message": "操作成功",
    "data": null
}
```

#### **3.1.2 获取用户信息**

- **接口路径**: `GET /api/user`
- **请求参数**:

```Plain Text
username: string  // 用户名，可选。不传则获取当前登录用户信息
```

- **响应示例**:

```JSON
{
    "code": 0,
    "message": "操作成功",
    "data": {
        "id": 1,
        "username": "admin",
        "nickname": "管理员",
        "email": "admin@example.com",
        "mobile": "13800138000",
        "avatar": "http://example.com/avatar.jpg",
        "status": 0,
        "userType": 1,
        "sort": 1,
        "lastLoginIp": "127.0.0.1",
        "lastLoginDate": "2024-01-01T12:00:00Z",
        "createdAt": "2024-01-01T12:00:00Z",
        "updatedAt": "2024-01-01T12:00:00Z"
    }
}
```

#### **3.1.3 获取用户列表**

- **接口路径**: `GET /api/user/list`
- **请求参数**:

```Plain Text
pageNum: int      // 页码，默认1
pageSize: int     // 每页大小，默认10
```

- **响应示例**:

```JSON
{
    "code": 0,
    "message": "操作成功",
    "data": {
        "total": 100,
        "data": [
            {
                "id": 1,
                "username": "admin",
                "nickname": "管理员",
                "email": "admin@example.com",
                "status": 0
            }
        ]
    }
}
```

#### **3.1.4 更新用户信息**

- **接口路径**: `POST /api/user/update`
- **请求参数**:

```JSON
{
    "id": 1,                  // 用户ID，必填
    "username": "string",     // 用户名，可选
    "nickname": "string",     // 昵称，可选
    "email": "string",        // 邮箱，可选
    "mobile": "string",       // 手机号，可选
    "status": 0               // 状态，可选 0正常 1禁用
}
```

#### **3.1.5 删除用户**

- **接口路径**: `POST /api/user/delete`
- **请求参数**:

```JSON
{
    "id": 1                   // 用户ID，必填
}
```

#### **3.1.6 修改密码**

- **接口路径**: `POST /api/user/changePassword`
- **请求参数**:

```JSON
{
    "userId": 1,              // 用户ID，必填
    "oldPassword": "string",  // 旧密码，必填
    "newPassword": "string"   // 新密码，必填
}
```

#### **3.1.7 绑定角色**

- **接口路径**: `POST /api/user/bindRole`
- **请求参数**:

```JSON
{
    "userId": 1,              // 用户ID，必填
    "roleIds": [1, 2, 3]      // 角色ID数组，必填
}
```

#### **3.1.8 解绑角色**

- **接口路径**: `POST /api/user/unbindRole`
- **请求参数**:

```JSON
{
    "userId": 1,              // 用户ID，必填
    "roleIds": [1, 2]         // 要解绑的角色ID数组，必填
}
```

#### **3.1.9 获取用户角色列表**

- **接口路径**: `GET /api/user/roles`
- **请求参数**:

```Plain Text
id: int                       // 用户ID，必填
```

- **响应示例**:

```JSON
{
    "code": 0,
    "message": "操作成功",
    "data": [
        {
            "id": 1,
            "name": "管理员",
            "remark": "系统管理员",
            "status": 0
        }
    ]
}
```

#### **3.1.10 获取用户菜单**

- **接口路径**: `GET /api/user/menus`
- **请求参数**:

```Plain Text
userId: int                   // 用户ID，可选。不传则获取当前登录用户菜单
```

- **响应示例**:

```JSON
{
    "code": 0,
    "message": "操作成功",
    "data": [
        {
            "id": 1,
            "name": "系统管理",
            "path": "/system",
            "component": "Layout",
            "icon": "system",
            "children": [
                {
                    "id": 2,
                    "name": "用户管理",
                    "path": "/system/user",
                    "component": "system/user/index",
                    "icon": "user"
                }
            ]
        }
    ]
}
```

### **3.2 角色管理接口**

#### **3.2.1 创建角色**

- **接口路径**: `POST /api/role/create`
- **请求参数**:

```JSON
{
    "id": 1,                  // 角色ID，可选
    "name": "string",         // 角色名称，必填
    "remark": "string",       // 备注，可选
    "status": 0               // 状态，必填 0正常 1禁用
}
```

#### **3.2.2 获取角色信息**

- **接口路径**: `GET /api/role`
- **说明**: 该接口暂未实现具体功能

#### **3.2.3 获取角色列表**

- **接口路径**: `GET /api/role/list`
- **请求参数**:

```Plain Text
pageNum: int                  // 页码，默认1
pageSize: int                 // 每页大小，默认10
```

- **响应示例**:

```JSON
{
    "code": 0,
    "message": "操作成功",
    "data": {
        "total": 10,
        "data": [
            {
                "id": 1,
                "name": "管理员",
                "remark": "系统管理员",
                "status": 0,
                "createdAt": "2024-01-01T12:00:00Z"
            }
        ]
    }
}
```

#### **3.2.4 更新角色**

- **接口路径**: `POST /api/role/update`
- **请求参数**:

```JSON
{
    "id": 1,                  // 角色ID，必填
    "name": "string",         // 角色名称，必填
    "remark": "string",       // 备注，可选
    "status": 0               // 状态，必填
}
```

#### **3.2.5 删除角色**

- **接口路径**: `POST /api/role/delete`
- **请求参数**:

```JSON
{
    "id": 1                   // 角色ID，必填
}
```

#### **3.2.6 绑定菜单**

- **接口路径**: `POST /api/role/bindMenu`
- **请求参数**:

```JSON
{
    "roleId": 1,              // 角色ID，必填
    "menuIds": [1, 2, 3]      // 菜单ID数组，必填
}
```

#### **3.2.7 解绑菜单**

- **接口路径**: `POST /api/role/unbindMenu`
- **请求参数**:

```JSON
{
    "roleId": 1,              // 角色ID，必填
    "menuIds": [1, 2]         // 要解绑的菜单ID数组，必填
}
```

#### **3.2.8 获取角色菜单列表**

- **接口路径**: `GET /api/role/menus`
- **请求参数**:

```Plain Text
id: int                       // 角色ID，必填
```

- **响应示例**:

```JSON
{
    "code": 0,
    "message": "操作成功",
    "data": [
        {
            "id": 1,
            "name": "系统管理",
            "path": "/system",
            "component": "Layout",
            "perms": "system:view",
            "status": 0
        }
    ]
}
```

#### **3.2.9 获取拥有该角色的用户列表**

- **接口路径**: `GET /api/role/users`
- **请求参数**:

```Plain Text
roleId: int                   // 角色ID，必填
```

- **响应示例**:

```JSON
{
    "code": 0,
    "message": "操作成功",
    "data": [
        {
            "id": 1,
            "username": "admin",
            "nickname": "管理员",
            "status": 0
        }
    ]
}
```

### **3.3 菜单管理接口**

#### **3.3.1 创建菜单**

- **接口路径**: `POST /api/menu/create`
- **请求参数**:

```JSON
{
    "name": "string",         // 菜单名称，必填
    "parentId": 0,            // 父菜单ID，必填，0表示根菜单
    "orderNum": 1,            // 排序号，必填
    "path": "string",         // 路由地址，可选
    "component": "string",    // 组件路径，可选
    "query": "string",        // 请求参数，可选
    "isFrame": 0,             // 是否外链，必填 0否 1是
    "menuType": "C",          // 菜单类型，必填 C目录 M菜单 B按钮
    "isCatch": 0,             // 是否缓存，必填 0否 1是
    "isHidden": 0,            // 是否隐藏，必填 0否 1是
    "perms": "string",        // 权限标识，可选
    "icon": "string",         // 图标，可选
    "status": 0,              // 状态，必填 0正常 1禁用
    "remark": "string"        // 备注，可选
}
```

#### **3.3.2 获取菜单信息**

- **接口路径**: `GET /api/menu`
- **请求参数**:

```Plain Text
id: int                       // 菜单ID，必填
```

- **响应示例**:

```JSON
{
    "code": 0,
    "message": "操作成功",
    "data": {
        "id": 1,
        "name": "系统管理",
        "parentId": 0,
        "orderNum": 1,
        "path": "/system",
        "component": "Layout",
        "menuType": "C",
        "icon": "system",
        "status": 0
    }
}
```

#### **3.3.3 获取菜单列表**

- **接口路径**: `GET /api/menu/list`
- **请求参数**:

```Plain Text
pageNum: int                  // 页码，默认1
pageSize: int                 // 每页大小，默认10
name: string                  // 菜单名称，可选（模糊查询）
status: int                   // 状态，可选 0正常 1禁用
```

- **响应示例**:

```JSON
{
    "code": 0,
    "message": "操作成功",
    "data": {
        "total": 20,
        "data": [
            {
                "id": 1,
                "name": "系统管理",
                "parentId": 0,
                "orderNum": 1,
                "path": "/system",
                "component": "Layout",
                "menuType": "C",
                "status": 0,
                "createdAt": "2024-01-01T12:00:00Z"
            }
        ]
    }
}
```

#### **3.3.4 更新菜单**

- **接口路径**: `POST /api/menu/update`
- **请求参数**:

```JSON
{
    "id": 1,                  // 菜单ID，必填
    "name": "string",         // 菜单名称，必填
    "parentId": 0,            // 父菜单ID，必填
    "orderNum": 1,            // 排序号，必填
    "path": "string",         // 路由地址，可选
    "component": "string",    // 组件路径，可选
    "query": "string",        // 请求参数，可选
    "isFrame": 0,             // 是否外链，必填
    "menuType": "C",          // 菜单类型，必填
    "isCatch": 0,             // 是否缓存，必填
    "isHidden": 0,            // 是否隐藏，必填
    "perms": "string",        // 权限标识，可选
    "icon": "string",         // 图标，可选
    "status": 0,              // 状态，必填
    "remark": "string"        // 备注，可选
}
```

#### **3.3.5 删除菜单**

- **接口路径**: `POST /api/menu/delete`
- **请求参数**:

```JSON
{
    "id": 1                   // 菜单ID，必填
}
```

- **说明**: 删除前会检查是否存在子菜单，如存在子菜单则不允许删除

#### **3.3.6 获取拥有该菜单的角色列表**

- **接口路径**: `GET /api/menu/roles`
- **请求参数**:

```Plain Text
menuId: int                   // 菜单ID，必填
```

- **响应示例**:

```JSON
{
    "code": 0,
    "message": "操作成功",
    "data": [
        {
            "id": 1,
            "name": "管理员",
            "remark": "系统管理员",
            "status": 0
        }
    ]
}
```

#### **3.3.7 获取菜单树形结构**

- **接口路径**: `GET /api/menu/tree`
- **响应示例**:

```JSON
{
    "code": 0,
    "message": "操作成功",
    "data": [
        {
            "id": 1,
            "name": "系统管理",
            "parentId": 0,
            "children": [
                {
                    "id": 2,
                    "name": "用户管理",
                    "parentId": 1,
                    "children": []
                }
            ]
        }
    ]
}
```

#### **3.3.8 获取角色菜单树**

- **接口路径**: `GET /api/menu/roleMenuTree`
- **请求参数**:

```Plain Text
roleId: int                   // 角色ID，必填
```

- **响应示例**:

```JSON
{
    "code": 0,
    "message": "操作成功",
    "data": {
        "menuIds": [1, 2, 3, 4, 5]  // 该角色拥有的菜单ID数组
    }
}
```

## **4. 状态码说明**

- **0**: 操作成功
- **1**: 操作失败
- **401**: 未授权
- **403**: 禁止访问
- **404**: 资源不存在
- **422**: 参数验证失败
- **500**: 服务器内部错误

## **5. 认证说明**

所有接口都需要通过JWT认证中间件验证，请求头需要携带有效的token：

```Plain Text
Authorization: Bearer <token>
```

## **6. 权限控制流程**

1. **用户登录**: 验证用户名密码，返回JWT token
2. **接口访问**: 携带token访问接口，中间件验证token有效性
3. **权限检查**: 根据用户拥有的角色，查询对应的菜单权限
4. **权限验证**: 检查用户是否有访问该接口的权限
5. **业务处理**: 权限验证通过后，执行具体的业务逻辑

## **7. 数据关系图**

```Plain Text
用户(User) ←→ 用户角色(UserRole) ←→ 角色(Role) ←→ 角色菜单(RoleMenu) ←→ 菜单(Menu)
```

- 一个用户可以拥有多个角色
- 一个角色可以被多个用户拥有
- 一个角色可以拥有多个菜单权限
- 一个菜单权限可以被多个角色拥有
- 用户最终权限 = 所有拥有角色的菜单权限的并集

## **8. 注意事项**

1. **事务处理**: 角色绑定/解绑操作使用数据库事务确保数据一致性
2. **软删除**: 部分表支持软删除，删除的记录不会物理删除
3. **密码安全**: 密码使用MD5加密存储（建议升级为更安全的加密方式）
4. **参数验证**: 所有接口都有完善的参数验证机制
5. **错误处理**: 统一的错误响应格式，便于前端处理
6. **日志记录**: 重要操作都有日志记录，便于问题追踪

## **9. 扩展建议**

1. **权限细粒度控制**: 可以在菜单基础上增加操作权限（增删改查）
2. **数据权限**: 可以基于部门、区域等维度实现数据权限控制
3. **角色继承**: 支持角色之间的继承关系
4. **权限缓存**: 使用Redis缓存用户权限信息，提高性能
5. **审计日志**: 记录所有权限变更操作的审计日志

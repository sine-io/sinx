# 前端调用链路说明（登录 → 资料 → 角色 → 菜单）

本文基于后端路由与处理器实现（Gin），梳理前端从登录到获取用户资料、角色、菜单的标准调用顺序与返回数据要点。

- 基础地址：`http://<host>:<port>`（本地默认 `http://localhost:8080`）
- 统一返回包裹：`{ code: number, message: string, data: any }`
- 鉴权方式：除开放接口外，均需在请求头携带 `Authorization: Bearer <token>`

---

## 调用总览

1. 登录获取 token

   - POST `/api/auth/login`
   - 入参：`{ username: string, password: string }`
   - 返回：`data.token`（JWT），`data.user`（简要用户信息）
   - 用途：后续所有需要鉴权的接口在 Header 中携带此 token

1. 获取当前用户资料（可选但推荐）

   - GET `/api/user/profile`（需要登录）
   - 返回：`data` 为当前登录用户资料（`id, username, email, is_active`）
   - 用途：展示用户信息；若未登录会返回 401

1. 获取用户角色（仅当页面需要或具备权限时）

   - GET `/api/user/roles?id=<userId>`（需要登录 + 权限 `user:roles`）
   - 返回：`data` 为角色列表，元素形如 `{ id, name, remark, status }`
   - 用途：管理端或角色可视化页面显示；普通业务页面一般可跳过

1. 获取菜单（前端常用）

   - GET `/api/user/menus[?userId=<id>]`（需要登录）
   - 返回：`data` 为“用户可见菜单树”，节点形如：
     - `MenuTreeNode = { id, name, parentId, path?, component?, icon?, children: MenuTreeNode[] }`
   - 用途：根据登录用户权限动态渲染侧边栏/路由

---

## 详细说明与示例

### 1. 登录（公共接口）

- 路径：`POST /api/auth/login`
- Body：

```json
{
  "username": "john_doe",
  "password": "password123"
}
```

- 成功响应（节选）：

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "token": "<JWT>",
    "user": { "id": 1, "username": "john_doe", "email": "john@example.com", "is_active": true }
  }
}
```

- 下一步：将 `data.token` 写入本地（例如 LocalStorage）用于后续请求头。

### 2. 用户资料（需要登录）

- 路径：`GET /api/user/profile`
- Header：`Authorization: Bearer <token>`
- 成功响应（节选）：

```json
{
  "code": 0,
  "data": { "id": 1, "username": "john_doe", "email": "john@example.com", "is_active": true }
}
```

- 失败：未带 token 或 token 失效 → `code=10003（unauthorized）`
- 下一步：若页面需要展示头像/昵称等，可在此步完成初始化。

### 3. 用户角色（需要登录且具备接口权限）

- 注意：该接口路由受权限中间件保护，需要 `user:roles` 权限。
- 路径：`GET /api/user/roles?id=<userId>`
- Header：`Authorization: Bearer <token>`
- 成功响应（示例）：

```json
{
  "code": 0,
  "data": [
    { "id": 1, "name": "管理员", "remark": "系统最高权限", "status": 0 },
    { "id": 2, "name": "运营", "remark": "后台运营", "status": 0 }
  ]
}
```

- 用途：
  - 管理端用户详情页、分配角色页展示
  - 前端如需根据“角色名”做 UI 逻辑（不推荐强依赖）
- 普通业务页面：可跳过此步，直接调用第 4 步获取“用户菜单”。

### 4. 用户菜单（常用）

- 路径：`GET /api/user/menus[?userId=<id>]`
- Header：`Authorization: Bearer <token>`
- 返回结构：菜单树（仅包含当前用户有权访问的菜单）

```json
{
  "code": 0,
  "data": [
    {
      "id": 1,
      "name": "系统管理",
      "parentId": 0,
      "path": "/system",
      "component": "Layout",
      "icon": "setting",
      "children": [
        { "id": 2, "name": "用户管理", "parentId": 1, "path": "/system/user", "component": "system/user/index", "children": [] }
      ]
    }
  ]
}
```

- 用途：渲染侧边栏、生成动态路由；`children` 为递归树。

---

## 相关接口补充（管理场景）

- 角色菜单树（返回菜单 ID 集合）：
  - GET `/api/menu/roleMenuTree?roleId=<id>`（需要登录+权限 `menu:roleMenuTree`）
  - 响应：`{ code: 0, data: { menuIds: number[] } }`
  - 用途：角色分配菜单时回显。

- 全量权限点（给前端做可视化/灰显控制）：
  - GET `/api/perms/all`（需要登录）
  - 响应：`{ code: 0, data: { "user:create": true, "user:list": true, ... } }`

- 全量菜单树（不做用户过滤）：
  - GET `/api/menu/tree`（需要登录）
  - 用途：菜单管理页展示整棵树。业务端渲染请使用 `/api/user/menus`。

---

## 推荐前端流程（伪代码）

```ts
async function boot() {
  // 1) 登录
  const loginRes = await api.post('/api/auth/login', { username, password });
  setToken(loginRes.data.token);

  // 2) 拉取资料（可选）
  const profile = await api.get('/api/user/profile');
  setCurrentUser(profile.data);

  // 3) 直接拉用户菜单（常用做法）
  const menus = await api.get('/api/user/menus');
  renderRoutes(menus.data);

  // 若是管理页且具备权限，可再拉角色：
  // const roles = await api.get(`/api/user/roles?id=${profile.data.id}`);
}
```

---

## 常见错误与处理

- 401 未认证（`code=10003`）：未带 `Authorization` 或 token 过期 → 重新登录
- 403 无权限（`code=10004`）：访问了受权限保护的接口 → 确认账号已分配角色且角色拥有对应菜单/权限点
- 列表/树为空：尚未初始化数据或账号未绑定角色/菜单 → 按管理流程创建角色、菜单并绑定

---

## 字段参考（来自后端实现）

- User（简）：`{ id, username, email, is_active }`
- Role（简）：`{ id, name, remark, status }`
- MenuTreeNode：`{ id, name, parentId, path?, component?, icon?, children: MenuTreeNode[] }`

以上与 `docs/swagger.yaml`、`api/handler/*`、`application/rbac/dto/*` 一致。

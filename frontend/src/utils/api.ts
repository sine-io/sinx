import request from './request'

export interface LoginPayload { username: string; password: string }

export async function login(data: LoginPayload) {
  return request.post('/auth/login', data)
}

export async function getProfile() {
  return request.get('/user/profile')
}

export async function getUserRoles(userId: number | string) {
  return request.get('/user/roles', { params: { id: userId } })
}

export async function getUserMenus(userId?: number | string) {
  if (userId !== undefined) {
    return request.get('/user/menus', { params: { userId } })
  }
  return request.get('/user/menus')
}

export async function getAllPerms() {
  return request.get('/perms/all')
}

// 当前登录用户的权限集合
export async function getMyPerms() {
  return request.get('/perms/me')
}

// 用户列表（用于获取 total 统计）
export async function getUserList(pageNum = 1, pageSize = 10) {
  return request.get('/user/list', { params: { pageNum, pageSize } })
}

// 角色列表（用于获取 total 统计）
export async function getRoleList(pageNum = 1, pageSize = 10) {
  return request.get('/role/list', { params: { pageNum, pageSize } })
}

// 菜单列表（用于获取 total 统计）
export async function getMenuList(pageNum = 1, pageSize = 10) {
  return request.get('/menu/list', { params: { pageNum, pageSize } })
}

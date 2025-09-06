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

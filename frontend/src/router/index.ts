import { createRouter, createWebHistory, type RouteRecordRaw, type NavigationGuardNext, type RouteLocationNormalized } from 'vue-router'
import { getFirstMenuPath } from './dynamic'
import { getToken } from '../utils/auth'
import { loadPerms, hasPerm } from '../utils/perms'

const Login = () => import('../views/Login.vue')
const Dashboard = () => import('../views/Dashboard.vue')

const routes: RouteRecordRaw[] = [
  { path: '/login', name: 'login', component: Login, meta: { public: true, title: '登录' } },
  { path: '/', name: 'dashboard', component: Dashboard, meta: { title: '仪表盘' } },
  { path: '/403', name: 'forbidden', component: () => import('../views/Forbidden.vue'), meta: { public: true, title: '无权访问' } },
  // 兜底 404 路由，需放在最后
  { path: '/:pathMatch(.*)*', name: 'not-found', component: () => import('../views/NotFound.vue'), meta: { public: true, title: '404' } },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

router.beforeEach((to: RouteLocationNormalized, _from: RouteLocationNormalized, next: NavigationGuardNext) => {
  const token = getToken()
  if (!to.meta.public && !token) {
    next({ name: 'login', query: { redirect: to.fullPath } })
    return
  }
  // 登录后访问根路径时，自动跳转到第一个菜单
  if (to.path === '/' && !to.meta.public) {
    const first = getFirstMenuPath()
    if (first && first !== '/') {
      next({ path: first, replace: true })
      return
    }
  }
  // 权限校验：基于路由 meta.perms 与本地缓存的权限集合
  const required = to.meta?.perms as string | undefined
  if (!to.meta.public && required) {
    const perms = loadPerms()
    if (!hasPerm(perms, required)) {
      next({ name: 'forbidden', query: { from: to.fullPath } })
      return
    }
  }
  next()
})

export default router

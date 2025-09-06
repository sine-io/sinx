import { createRouter, createWebHistory, type RouteRecordRaw, type NavigationGuardNext, type RouteLocationNormalized } from 'vue-router'
import { getToken } from '../utils/auth'

const Login = () => import('../views/Login.vue')
const Dashboard = () => import('../views/Dashboard.vue')

const routes: RouteRecordRaw[] = [
  { path: '/login', name: 'login', component: Login, meta: { public: true, title: '登录' } },
  { path: '/', name: 'dashboard', component: Dashboard, meta: { title: '仪表盘' } },
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
  next()
})

export default router

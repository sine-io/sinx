import type { Router, RouteRecordRaw } from 'vue-router'

interface RawNode {
  id: number | string
  name: string
  path?: string
  component?: string
  children?: RawNode[]
  isHidden?: number
  menuType?: 'C' | 'M' | 'B'
  status?: number
  perms?: string
}

// 预扫描 views 目录（Vite 的动态导入）
const viewModules = import.meta.glob('../views/**/*.vue')

// 懒加载真实页面组件；若不存在则回退到占位页
function resolveComponent(component?: string) {
  if (component) {
    const key = `../${component}.vue`
    const mod = (viewModules as Record<string, any>)[key]
    if (mod) return mod
  }
  return () => import('../views/PlaceholderPage.vue')
}

// 父级占位（仅承载 children 的 <router-view/>）
function parentHolder() {
  return () => import('../layouts/BlankRouterView.vue')
}

function visible(nodes: RawNode[] = []): RawNode[] {
  return nodes.filter((n) => (n.status ?? 0) === 0 && (n.isHidden ?? 0) === 0 && n.menuType !== 'B')
}

function firstLeafPath(node: RawNode): string | undefined {
  if (!node.children || node.children.length === 0) return node.path
  for (const c of visible(node.children)) {
    const p = firstLeafPath(c)
    if (p) return p
  }
  return node.path
}

function normalizeName(path?: string, id?: string | number) {
  if (path) {
    const name = path.replace(/^\/+|\/+$/g, '').replace(/\//g, '-') || 'root'
    return name
  }
  return `menu-${id}`
}

function toRoute(node: RawNode): RouteRecordRaw | null {
  if (!node.path) return null
  const name = normalizeName(node.path, node.id)
  const children = visible(node.children || [])
  if (children.length > 0) {
    const childRoutes = children.map(toRoute).filter(Boolean) as RouteRecordRaw[]
    const redirect = firstLeafPath(node)
    return {
      path: node.path!,
      name,
      component: parentHolder(),
      redirect,
      meta: { title: node.name, perms: node.perms },
      children: childRoutes,
    }
  }
  return {
    path: node.path!,
    name,
    component: resolveComponent(node.component),
    meta: { title: node.name, perms: node.perms },
  }
}

export async function initDynamicRoutes(router: Router) {
  try {
    const res = await fetch('/static/config/tree.json', { cache: 'no-store' })
    const data = (await res.json()) as RawNode[]
    const roots = visible(data)
    const routes = roots.map(toRoute).filter(Boolean) as RouteRecordRaw[]
    routes.forEach((r) => router.addRoute(r))
    // 记录第一个可访问的叶子路径
    _firstMenuPath = roots.length ? firstLeafPath(roots[0]) : undefined
  } catch (e) {
    // 忽略，保底不影响启动
    console.warn('initDynamicRoutes failed:', e)
  }
}

// 缓存第一个菜单叶子路径，供登录后首页重定向使用
let _firstMenuPath: string | undefined
export function getFirstMenuPath() {
  return _firstMenuPath
}

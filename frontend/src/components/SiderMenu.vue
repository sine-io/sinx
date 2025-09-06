<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import MenuTree from './MenuTree.vue'

interface RawNode {
  id: number | string
  name: string
  path?: string
  children?: RawNode[]
  isHidden?: number
  menuType?: 'C' | 'M' | 'B'
  status?: number
}

interface MenuNode {
  id: number | string
  name: string
  path?: string
  children?: MenuNode[]
}

defineOptions({ name: 'SiderMenu' })

const router = useRouter()
const route = useRoute()
const loading = ref(true)
const raw = ref<RawNode[]>([])
const menu = ref<MenuNode[]>([])

function filterNodes(nodes: RawNode[] = []): MenuNode[] {
  const visible = nodes
    .filter((n) => (n.status ?? 0) === 0 && (n.isHidden ?? 0) === 0 && (n.menuType !== 'B'))
    .map((n) => ({
      id: n.id,
      name: n.name,
      path: n.path || undefined,
      children: n.children ? filterNodes(n.children) : undefined,
    }))
  // 清除空 children
  return visible.map((n) => ({ ...n, children: (n.children && n.children.length ? n.children : undefined) }))
}

async function loadTree() {
  loading.value = true
  try {
    const res = await fetch('/static/config/tree.json', { cache: 'no-store' })
    const data = (await res.json()) as RawNode[]
    raw.value = data
    menu.value = filterNodes(data)
  } finally {
    loading.value = false
  }
}

// 平铺 key->node 映射，key 使用 path 优先，否则 id
const keyNodeMap = computed<Record<string, MenuNode>>(() => {
  const map: Record<string, MenuNode> = {}
  const walk = (nodes?: MenuNode[]) => {
    nodes?.forEach((n) => {
      const key = n.path || String(n.id)
      map[key] = n
      walk(n.children)
    })
  }
  walk(menu.value)
  return map
})

// 计算选中/展开 keys
const selectedKeys = ref<string[]>([])
const openKeys = ref<string[]>([])

function findParents(targetKey: string): string[] {
  const parents: string[] = []
  function dfs(nodes: MenuNode[], path: string[]) {
    for (const n of nodes) {
      const key = n.path || String(n.id)
      const next = [...path, key]
      if (key === targetKey) {
        parents.push(...path)
        return true
      }
      if (n.children && dfs(n.children, next)) return true
    }
    return false
  }
  dfs(menu.value, [])
  return Array.from(new Set(parents))
}

function syncByRoutePath(path: string) {
  const key = path
  if (keyNodeMap.value[key]) {
    selectedKeys.value = [key]
    openKeys.value = findParents(key)
  } else {
    // 若路径不在菜单中，则仅保持展开不变
    selectedKeys.value = []
  }
}

function onMenuItemClick(info: string | { key?: string }) {
  // Arco Menu 会传递对象 { key, ... }，兼容字符串与对象两种形态
  const key = typeof info === 'string' ? info : info?.key
  if (!key) return
  // 仅在路由存在时跳转；避免跳到不存在的页面
  const exists = router.getRoutes().some((r) => r.path === key)
  if (exists) router.push(key)
  else selectedKeys.value = [key]
}

onMounted(async () => {
  await loadTree()
  syncByRoutePath(route.path)
})

watch(
  () => route.path,
  (p) => syncByRoutePath(p),
)
</script>

<template>
  <a-menu
    v-if="!loading"
    v-model:selectedKeys="selectedKeys"
  v-model:openKeys="openKeys"
    @menu-item-click="onMenuItemClick"
    :style="{ height: '100%', borderRight: 0 }"
    breakpoint="lg"
  >
    <MenuTree :nodes="menu" />
  </a-menu>
  <div v-else style="padding: 12px; color: var(--color-text-2);">加载菜单中...</div>
  
</template>

<style scoped>
</style>

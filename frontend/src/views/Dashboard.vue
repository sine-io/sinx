<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { getProfile, getUserMenus, getAllPerms } from '../utils/api'
import { clearToken, getToken } from '../utils/auth'
import MenuTree from '../components/MenuTree.vue'

const profile = ref<any>(null)
const menus = ref<any[]>([])
const perms = ref<Record<string, boolean>>({})
const selectedKey = ref<string>()
const loading = ref(false)
const router = useRouter()

function firstKey(nodes: any[]): string | undefined {
  for (const n of nodes || []) {
    const k = n.path || String(n.id)
    if (n.children && n.children.length) {
      return firstKey(n.children) || k
    }
    return k
  }
  return undefined
}

async function init() {
  try {
    loading.value = true
  const profileRes: any = await getProfile()
  profile.value = profileRes?.data
  const menusRes: any = await getUserMenus()
  menus.value = menusRes?.data || []
  const permsRes: any = await getAllPerms()
  perms.value = (permsRes?.data || {}) as Record<string, boolean>
  if (!selectedKey.value) {
    selectedKey.value = firstKey(menus.value)
  }
  } finally {
    loading.value = false
  }
}

function onLogout() {
  clearToken()
  router.replace('/login')
}

function onMenuClick(key: string) {
  selectedKey.value = key
}

onMounted(() => {
  if (!getToken()) {
    router.replace('/login')
    return
  }
  init()
})
</script>

<template>
  <div class="page">
    <a-layout class="layout">
      <a-layout-sider :width="220">
        <div style="height: 56px; display:flex; align-items:center; padding: 0 16px; font-weight: 600">SINX</div>
        <a-menu :selected-keys="selectedKey ? [selectedKey] : []" @menu-item-click="onMenuClick">
          <MenuTree :nodes="menus" />
        </a-menu>
      </a-layout-sider>
      <a-layout>
        <a-layout-header>
          <div style="display:flex; align-items:center; justify-content: space-between;">
            <div>仪表盘</div>
            <a-button type="outline" status="danger" @click="onLogout">退出登录</a-button>
          </div>
        </a-layout-header>
        <a-layout-content class="content">
          <a-card :loading="loading">
            <p><b>当前用户：</b> {{ profile?.username }} (ID: {{ profile?.id }})</p>
            <p><b>邮箱：</b> {{ profile?.email }}</p>
            <p><b>权限点数量：</b> {{ Object.keys(perms).length }}</p>
            <p><b>当前菜单：</b> {{ selectedKey || '-' }}</p>
          </a-card>
        </a-layout-content>
      </a-layout>
    </a-layout>
  </div>
  
</template>

<style scoped lang="less">
.page {
  height: 100%;
}
.layout {
  height: 100vh;
}
.content {
  padding: 16px;
}
</style>

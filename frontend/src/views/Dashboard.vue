<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { getProfile, getUserMenus, getAllPerms } from '../utils/api'
import { clearToken, getToken } from '../utils/auth'

const profile = ref<any>(null)
const menus = ref<any[]>([])
const perms = ref<Record<string, boolean>>({})
const loading = ref(false)
const router = useRouter()

async function init() {
  try {
    loading.value = true
  const profileRes: any = await getProfile()
  profile.value = profileRes?.data
  const menusRes: any = await getUserMenus()
  menus.value = menusRes?.data || []
  const permsRes: any = await getAllPerms()
  perms.value = (permsRes?.data || {}) as Record<string, boolean>
  } finally {
    loading.value = false
  }
}

function onLogout() {
  clearToken()
  router.replace('/login')
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
    <a-page-header title="仪表盘" :breadcrumb="{ routes: [{ path: '/', breadcrumbName: 'Home' }] }">
      <template #extra>
        <a-button type="outline" status="danger" @click="onLogout">退出登录</a-button>
      </template>
    </a-page-header>
    <a-card :loading="loading">
      <p><b>当前用户：</b> {{ profile?.username }} (ID: {{ profile?.id }})</p>
      <p><b>邮箱：</b> {{ profile?.email }}</p>
      <p><b>权限点数量：</b> {{ Object.keys(perms).length }}</p>
    </a-card>
  </div>
  
</template>

<style scoped lang="less">
.page {
  padding: 16px;
}
</style>

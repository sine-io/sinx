<script setup lang="ts">
import { computed, ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { getProfile } from './utils/api'
import { clearToken } from './utils/auth'
import SiderMenu from './components/SiderMenu.vue'
import Breadcrumbs from './components/Breadcrumbs.vue'

const route = useRoute()
const router = useRouter()
const isLogin = computed(() => route.name === 'login')

const username = ref<string>('')
async function loadProfile() {
  try {
    const res: any = await getProfile()
    username.value = res?.data?.username || ''
  } catch {}
}

function onLogout() {
  clearToken()
  router.replace('/login')
}

onMounted(() => {
  if (!isLogin.value) loadProfile()
})
</script>

<template>
  <template v-if="isLogin">
    <router-view />
  </template>
  <template v-else>
    <a-layout style="height: 100%">
      <a-layout-header style="display:flex;align-items:center;gap:12px;justify-content:space-between">
        <div style="font-weight:600">SinX Admin</div>
        <div style="display:flex;align-items:center;gap:12px">
          <span style="color:var(--color-text-2)">{{ username || '未登录' }}</span>
          <a-button size="small" status="danger" @click="onLogout">退出</a-button>
        </div>
      </a-layout-header>
      <a-layout>
        <a-layout-sider collapsible :width="220">
          <SiderMenu />
        </a-layout-sider>
        <a-layout-content style="padding:16px">
          <Breadcrumbs />
          <div style="height:12px" />
          <router-view />
        </a-layout-content>
      </a-layout>
    </a-layout>
  </template>
  
</template>

<style scoped>
</style>

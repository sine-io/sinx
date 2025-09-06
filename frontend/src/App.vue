<script setup lang="ts">
import { computed, ref, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { getProfile } from './utils/api'
import { clearToken, getToken } from './utils/auth'
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

// 监听路由变化：当进入受保护页面且本地有 token 时，刷新一次用户信息
watch(
  () => route.fullPath,
  () => {
    if (!isLogin.value && getToken()) {
      loadProfile()
    } else if (isLogin.value) {
      username.value = ''
    }
  }
)
</script>

<template>
  <template v-if="isLogin">
    <router-view />
  </template>
  <template v-else>
    <a-layout style="height: 100%">
      <a-layout-header class="header-bar">
        <div class="brand">SinX Admin</div>
        <div class="user-area">
          <span class="username">{{ username || '未登录' }}</span>
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
/* Header beautify */
.header-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  height: 56px;
  padding: 0 20px; /* prevent flush to edges */
  background: var(--color-bg-1, #fff);
  border-bottom: 1px solid var(--color-border-2, #e5e6eb);
  box-sizing: border-box;
}

.brand {
  font-weight: 600;
  font-size: 16px;
  letter-spacing: 0.3px;
  color: var(--color-text-1, #1d2129);
}

.user-area {
  display: flex;
  align-items: center;
  gap: 12px;
}

.username {
  color: var(--color-text-2, #4e5969);
}
</style>

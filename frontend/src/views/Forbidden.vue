<script setup lang="ts">
import { useRouter } from 'vue-router'
import { getToken } from '../utils/auth'
import { getFirstMenuPath } from '../router/dynamic'

const router = useRouter()

function goHome() {
  // 优先跳转首个可见菜单；若未登录则回登录；否则回根路径
  const first = getFirstMenuPath()
  const token = getToken()
  if (!token) {
    router.replace('/login')
    return
  }
  router.replace(first || '/')
}

function goBack() {
  router.back()
}
</script>

<template>
  <div style="display:flex;align-items:center;justify-content:center;min-height:60vh">
    <a-result status="403" title="无权访问" subtitle="抱歉，您没有访问该页面的权限。">
      <template #extra>
        <a-space>
          <a-button type="primary" @click="goHome">返回首页</a-button>
          <a-button @click="goBack">返回上一页</a-button>
        </a-space>
      </template>
    </a-result>
  </div>
</template>

<style scoped>
</style>

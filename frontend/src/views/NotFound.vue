<script setup lang="ts">
import { useRouter } from 'vue-router'
import { getFirstMenuPath } from '../router/dynamic'

const router = useRouter()

function goHome() {
  const first = getFirstMenuPath()
  router.replace(first || '/')
}

function goBack() {
  router.back()
}
</script>

<template>
  <div style="display:flex;align-items:center;justify-content:center;min-height:60vh">
    <a-result status="404" title="页面不存在" subtitle="抱歉，您访问的页面不存在或已被移除。">
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

<script setup lang="ts">
import { useRouter } from 'vue-router'

const router = useRouter()

function goHome() {
  router.replace('/')
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

<script setup lang="ts">
import { useRoute } from 'vue-router'
const route = useRoute()
</script>

<template>
  <a-result status="info" title="页面建设中">
    <template #subtitle>
      路由 {{ route.fullPath }} 暂无对应页面组件。
    </template>
  </a-result>
</template>

<style scoped>
</style>

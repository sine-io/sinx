<script setup lang="ts">
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'

const route = useRoute()
const router = useRouter()

// 基于当前路由的 matched 生成面包屑，过滤掉没有 meta.title 的项
const crumbs = computed(() => {
  return route.matched
    .filter((r) => r.meta && (r.meta as any).title)
    .map((r) => ({
      title: (r.meta as any).title as string,
      path: r.path,
    }))
})

function onCrumbClick(path: string, isLast: boolean) {
  if (!isLast) router.push(path)
}
</script>

<template>
  <a-breadcrumb>
    <a-breadcrumb-item
      v-for="(c, idx) in crumbs"
      :key="c.path || idx"
      :class="{ clickable: idx < crumbs.length - 1 }"
      @click="onCrumbClick(c.path, idx === crumbs.length - 1)"
    >
      {{ c.title }}
    </a-breadcrumb-item>
  </a-breadcrumb>
</template>

<style scoped>
.clickable { cursor: pointer; }
</style>

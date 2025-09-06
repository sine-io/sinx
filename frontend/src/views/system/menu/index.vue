<script setup lang="ts">
import { useRoute } from 'vue-router'
const route = useRoute()
</script>

<template>
  <a-space direction="vertical" size="large" fill>
    <a-card>
      <a-space style="width:100%" align="center" justify="space-between">
        <h3 style="margin:0">{{ route.meta.title || '菜单管理' }}</h3>
        <a-button type="primary">新增菜单</a-button>
      </a-space>
    </a-card>
    <a-card>
      <a-tree :data="[]" block-node :field-names="{ key: 'id', title: 'name', children: 'children' }" />
    </a-card>
  </a-space>
</template>

<style scoped>
</style>

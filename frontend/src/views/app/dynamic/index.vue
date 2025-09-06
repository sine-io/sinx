<script setup lang="ts">
import { useRoute } from 'vue-router'
const route = useRoute()
</script>

<template>
  <a-space direction="vertical" size="large" fill>
    <a-card>
      <a-space style="width:100%" align="center" justify="space-between">
        <h3 style="margin:0">{{ route.meta.title || '动态管理' }}</h3>
        <a-input-search placeholder="搜索动态" allow-clear style="width:220px" />
      </a-space>
    </a-card>
    <a-card>
      <a-table :columns="[{title:'动态ID',dataIndex:'id'},{title:'作者',dataIndex:'author'},{title:'发布时间',dataIndex:'createdAt'}]" :data="[]" :pagination="{ pageSize: 10 }" />
    </a-card>
  </a-space>
</template>

<style scoped>
</style>

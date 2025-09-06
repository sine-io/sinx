<script setup lang="ts">
import { useRoute } from 'vue-router'
const route = useRoute()
</script>

<template>
  <a-space direction="vertical" size="large" fill>
    <a-card>
      <a-space style="width:100%" align="center" justify="space-between">
        <h3 style="margin:0">{{ route.meta.title || '用户管理' }}</h3>
        <a-space>
          <a-input-search placeholder="搜索用户名/邮箱" allow-clear style="width:260px" />
          <a-button type="primary">新增用户</a-button>
        </a-space>
      </a-space>
    </a-card>
    <a-card>
      <a-table :columns="[{title:'用户ID',dataIndex:'id'},{title:'用户名',dataIndex:'username'},{title:'邮箱',dataIndex:'email'},{title:'状态',dataIndex:'status'}]" :data="[]" :pagination="{ pageSize: 10 }" />
    </a-card>
  </a-space>
</template>

<style scoped>
</style>

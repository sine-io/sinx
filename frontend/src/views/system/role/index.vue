<script setup lang="ts">
import { useRoute } from 'vue-router'
const route = useRoute()
</script>

<template>
  <a-space direction="vertical" size="large" fill>
    <a-card>
      <a-space style="width:100%" align="center" justify="space-between">
        <h3 style="margin:0">{{ route.meta.title || '角色管理' }}</h3>
        <a-space>
          <a-input-search placeholder="搜索角色" allow-clear style="width:220px" />
          <a-button type="primary">新增角色</a-button>
        </a-space>
      </a-space>
    </a-card>
    <a-card>
      <a-table :columns="[{title:'角色ID',dataIndex:'id'},{title:'角色名',dataIndex:'name'},{title:'描述',dataIndex:'remark'}]" :data="[]" :pagination="{ pageSize: 10 }" />
    </a-card>
  </a-space>
</template>

<style scoped>
</style>

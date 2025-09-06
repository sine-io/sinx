<script setup lang="ts">
import { useRoute } from 'vue-router'
const route = useRoute()
</script>

<template>
  <a-space direction="vertical" size="large" fill>
    <a-card>
      <a-space style="width:100%" align="center" justify="space-between">
        <h3 style="margin:0">{{ route.meta.title || '群组管理' }}</h3>
        <a-button type="primary">新建群组</a-button>
      </a-space>
    </a-card>
    <a-card>
      <a-table :columns="[{title:'群组ID',dataIndex:'id'},{title:'群组名',dataIndex:'name'},{title:'成员数',dataIndex:'memberCount'}]" :data="[]" :pagination="{ pageSize: 10 }" />
    </a-card>
  </a-space>
</template>

<style scoped>
</style>

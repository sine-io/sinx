<script setup lang="ts">
import { useRoute } from 'vue-router'
const route = useRoute()
</script>

<template>
  <a-space direction="vertical" size="large" fill>
    <a-card>
      <a-space style="width:100%" align="center" justify="space-between">
        <h3 style="margin:0">{{ route.meta.title || 'APP用户管理' }}</h3>
        <a-input-search placeholder="搜索用户" allow-clear style="width:220px" />
      </a-space>
    </a-card>
    <a-card>
      <a-table :columns="[{title:'UID',dataIndex:'id'},{title:'昵称',dataIndex:'nickname'},{title:'手机号',dataIndex:'phone'}]" :data="[]" :pagination="{ pageSize: 10 }" />
    </a-card>
  </a-space>
</template>

<style scoped>
</style>

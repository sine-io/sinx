<script setup lang="ts">
import { useRoute } from 'vue-router'
const route = useRoute()
</script>

<template>
  <a-space direction="vertical" size="large" fill>
    <a-card>
      <a-space style="width:100%" align="center" justify="space-between">
        <h3 style="margin:0">{{ route.meta.title || '文件管理' }}</h3>
        <a-space>
          <a-upload action="#" :show-file-list="false">
            <a-button type="primary">上传文件</a-button>
          </a-upload>
          <a-button status="danger">删除所选</a-button>
        </a-space>
      </a-space>
    </a-card>
    <a-card>
      <a-table :columns="[{title:'对象名',dataIndex:'name'},{title:'大小',dataIndex:'size'},{title:'修改时间',dataIndex:'updatedAt'}]" :data="[]" :row-selection="{}" :pagination="{ pageSize: 10 }" />
    </a-card>
  </a-space>
</template>

<style scoped>
</style>

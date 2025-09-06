<script setup lang="ts">
import { useRoute } from 'vue-router'
const route = useRoute()
</script>

<template>
  <a-card>
    <h3 style="margin:0 0 16px 0">{{ route.meta.title || '个人信息' }}</h3>
    <a-form :model="{}" layout="vertical" style="max-width:520px">
      <a-form-item label="昵称">
        <a-input placeholder="请输入昵称" />
      </a-form-item>
      <a-form-item label="邮箱">
        <a-input placeholder="请输入邮箱" />
      </a-form-item>
      <a-form-item>
        <a-button type="primary">保存</a-button>
      </a-form-item>
    </a-form>
  </a-card>
</template>

<style scoped>
</style>

<script setup lang="ts">
import { useRoute } from 'vue-router'
const route = useRoute()
</script>

<template>
  <a-card>
    <h3 style="margin:0 0 16px 0">{{ route.meta.title || '密码修改' }}</h3>
    <a-form :model="{}" layout="vertical" style="max-width:520px">
      <a-form-item label="原密码">
        <a-input-password placeholder="请输入原密码" />
      </a-form-item>
      <a-form-item label="新密码">
        <a-input-password placeholder="请输入新密码" />
      </a-form-item>
      <a-form-item label="确认密码">
        <a-input-password placeholder="请再次输入新密码" />
      </a-form-item>
      <a-form-item>
        <a-button type="primary">保存</a-button>
      </a-form-item>
    </a-form>
  </a-card>
</template>

<style scoped>
</style>

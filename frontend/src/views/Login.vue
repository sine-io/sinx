<script setup lang="ts">
import { reactive, ref } from 'vue'
import { Message } from '@arco-design/web-vue'
import { useRoute, useRouter } from 'vue-router'
import { login } from '../utils/api'
import { setToken } from '../utils/auth'

interface FormState { username: string; password: string }
const form = reactive<FormState>({ username: '', password: '' })
const loading = ref(false)
const router = useRouter()
const route = useRoute()

async function onSubmit() {
  if (!form.username || !form.password) return
  loading.value = true
  try {
    const res: any = await login(form)
    const token = res?.data?.token
    if (token) setToken(token)
    const redirect = (route.query.redirect as string) || '/'
    router.replace(redirect)
  } catch (e: any) {
    // Show a friendly error message
    const msg = e?.response?.data?.message || e?.message || '登录失败，请重试'
  Message.error(msg)
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="login-wrapper">
    <a-card class="login-card" title="登录">
      <a-form :model="form" layout="vertical" @submit="onSubmit">
        <a-form-item label="用户名">
          <a-input v-model="form.username" placeholder="请输入用户名" />
        </a-form-item>
        <a-form-item label="密码">
          <a-input-password v-model="form.password" placeholder="请输入密码" />
        </a-form-item>
        <a-form-item>
          <a-button type="primary" html-type="submit" :loading="loading" long>登录</a-button>
        </a-form-item>
      </a-form>
    </a-card>
  </div>
  
</template>

<style scoped lang="less">
.login-wrapper {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #f5f7fa;
}
.login-card {
  width: 360px;
}
</style>

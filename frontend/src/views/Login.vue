<script setup lang="ts">
import { reactive, ref, computed } from 'vue'
import { Message } from '@arco-design/web-vue'
import { useRoute, useRouter } from 'vue-router'
import { login, getAllPerms } from '../utils/api'
import { setToken } from '../utils/auth'
import { savePerms } from '../utils/perms'

interface FormState { username: string; password: string }
const form = reactive<FormState>({ username: '', password: '' })
const loading = ref(false)
const canSubmit = computed(() => !!form.username?.trim() && !!form.password?.trim())
// 使用 Arco 表单进行校验
const formRef = ref<any>(null)
const rules: any = {
  username: [{ required: true, message: '请输入用户名', trigger: ['change', 'blur'] }],
  password: [{ required: true, message: '请输入密码', trigger: ['change', 'blur'] }],
}
const router = useRouter()
const route = useRoute()

async function onSubmit() {
  // 额外防抖：空值直接拦截，避免触发网络请求
  if (!form.username?.trim() || !form.password?.trim()) {
    Message.error('请输入用户名和密码')
    return
  }
  // 统一使用表单校验，阻止空值提交
  try {
    await formRef.value?.validate()
  } catch {
    Message.error('请填写必填项')
    return
  }
  loading.value = true
  try {
    const res: any = await login(form)
    const token = res?.data?.token
    if (token) setToken(token)
    // 登录后先拉取权限并写入缓存，再进行路由跳转，确保守卫放行
    try {
      const permsRes: any = await getAllPerms()
      const list: string[] = permsRes?.data || []
      if (Array.isArray(list)) {
        savePerms(new Set(list))
      }
    } catch {
      // 拉取权限失败不阻塞登录跳转
    }
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
      <a-form ref="formRef" :model="form" :rules="rules" layout="vertical" @submit="onSubmit">
        <a-form-item field="username" label="用户名" :rules="rules.username">
          <a-input v-model="form.username" placeholder="请输入用户名" />
        </a-form-item>
        <a-form-item field="password" label="密码" :rules="rules.password">
          <a-input-password v-model="form.password" placeholder="请输入密码" />
        </a-form-item>
        <a-form-item>
          <a-button type="primary" html-type="submit" :loading="loading" :disabled="!canSubmit" long>登录</a-button>
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

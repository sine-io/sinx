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
  <div class="login-page">
    <div class="decor" aria-hidden="true">
      <div class="blob b1" />
      <div class="blob b2" />
      <div class="grid" />
    </div>
    <div class="login-container">
      <div class="left">
        <div class="brand">
          <img class="logo" src="/vite.svg" alt="logo" />
          <div class="brand-text">
            <div class="brand-title">SinX Admin</div>
            <div class="brand-sub">轻量 · 简洁 · 可扩展的权限后台</div>
          </div>
        </div>
        <ul class="highlights">
          <li>• 现代化 UI 与响应式布局</li>
          <li>• 动态路由与细粒度权限控制</li>
          <li>• 更顺滑的表单校验与交互</li>
        </ul>
      </div>

      <a-card class="login-card" :bordered="false">
        <template #title>
          <div class="card-title">欢迎登录</div>
        </template>
        <a-form ref="formRef" :model="form" :rules="rules" layout="vertical" @submit="onSubmit">
          <a-form-item field="username" label="用户名" :rules="rules.username">
            <a-input v-model="form.username" placeholder="请输入用户名" allow-clear />
          </a-form-item>
          <a-form-item field="password" label="密码" :rules="rules.password">
            <a-input-password v-model="form.password" placeholder="请输入密码" allow-clear />
          </a-form-item>
          <div class="form-actions">
            <a-link href="javascript:void(0)" @click="Message.info('请联系管理员重置密码')">忘记密码？</a-link>
          </div>
          <a-space direction="vertical" size="medium" style="width: 100%">
            <a-button type="primary" html-type="submit" :loading="loading" :disabled="!canSubmit" long>
              登录
            </a-button>
          </a-space>
        </a-form>
      </a-card>
    </div>
  </div>
  
</template>

<style scoped lang="less">
.login-page {
  position: relative;
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  overflow: hidden;
  background: linear-gradient(135deg, #f3f6ff 0%, #f9fbff 100%);
}

.decor {
  pointer-events: none;
  position: absolute;
  inset: 0;
}
.decor .blob {
  position: absolute;
  width: 520px;
  height: 520px;
  border-radius: 50%;
  filter: blur(60px);
  opacity: 0.6;
}
.decor .b1 {
  background: radial-gradient(closest-side, #8b5cf6, transparent 70%);
  top: -120px;
  left: -120px;
}
.decor .b2 {
  background: radial-gradient(closest-side, #38bdf8, transparent 70%);
  right: -160px;
  bottom: -160px;
}
.decor .grid {
  position: absolute;
  inset: 0;
  background-image: linear-gradient(rgba(0, 0, 0, 0.04) 1px, transparent 1px),
    linear-gradient(90deg, rgba(0, 0, 0, 0.04) 1px, transparent 1px);
  background-size: 24px 24px;
  mask-image: radial-gradient(circle at 30% 30%, #000 40%, transparent 70%);
}

.login-container {
  position: relative;
  display: grid;
  grid-template-columns: 1.2fr 1fr;
  gap: 48px;
  align-items: stretch;
  width: min(100%, 1024px);
  padding: 24px;
  box-sizing: border-box;
}

.left {
  display: flex;
  flex-direction: column;
  justify-content: center;
  padding: 32px 24px 32px 32px;
  border-radius: 16px;
  color: #1d2129;
  background: rgba(255, 255, 255, 0.7);
  backdrop-filter: saturate(1.2) blur(10px);
  border: 1px solid rgba(46, 51, 56, 0.08);
}

.brand {
  display: flex;
  align-items: center;
  gap: 14px;
}
.logo {
  width: 42px;
  height: 42px;
}
.brand-title {
  font-weight: 700;
  font-size: 22px;
  letter-spacing: 0.2px;
}
.brand-sub {
  margin-top: 2px;
  color: #4e5969;
  font-size: 13px;
}

.highlights {
  margin: 18px 0 0;
  padding: 0;
  list-style: none;
  color: #3d3d3d;
  line-height: 1.9;
}

.login-card {
  align-self: center;
  width: 420px;
  border-radius: 16px;
  box-shadow: 0 12px 30px rgba(37, 45, 54, 0.08);
  background: rgba(255, 255, 255, 0.85);
  backdrop-filter: blur(6px);
}
.card-title {
  font-weight: 600;
  color: #1d2129;
}

.form-actions {
  display: flex;
  justify-content: flex-end;
  margin: -4px 0 4px;
}

@media (max-width: 920px) {
  .login-container {
    grid-template-columns: 1fr;
    gap: 20px;
    padding: 16px;
  }
  .left {
    display: none;
  }
  .login-card {
    width: 92vw;
    max-width: 420px;
  }
}
</style>

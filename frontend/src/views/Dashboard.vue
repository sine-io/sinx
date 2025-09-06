<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { getProfile } from '../utils/api'
import { getToken } from '../utils/auth'

const profile = ref<any>(null)
const loading = ref(false)
const router = useRouter()

async function init() {
  try {
    loading.value = true
    const profileRes: any = await getProfile()
    profile.value = profileRes?.data
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  if (!getToken()) {
    router.replace('/login')
    return
  }
  init()
})
</script>

<template>
  <a-card :loading="loading">
    <p><b>当前用户：</b> {{ profile?.username }} (ID: {{ profile?.id }})</p>
    <p><b>邮箱：</b> {{ profile?.email }}</p>
  </a-card>
</template>

<style scoped>
</style>

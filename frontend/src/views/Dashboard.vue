<script setup lang="ts">
import { ref, onMounted, onBeforeUnmount, nextTick } from 'vue'
import { useRouter } from 'vue-router'
import { getProfile, getUserList, getRoleList, getMenuList } from '../utils/api'
import { getToken } from '../utils/auth'
import * as echarts from 'echarts'

type Stat = { label: string; value: number; link?: string }

const profile = ref<any>(null)
const loading = ref(false)
const statLoading = ref(false)
const router = useRouter()

const stats = ref<Stat[]>([
  { label: '用户数', value: 0, link: '/system/user' },
  { label: '角色数', value: 0, link: '/system/role' },
  { label: '菜单数', value: 0, link: '/system/menu' },
])

// ECharts
const chartRef = ref<HTMLDivElement | null>(null)
let chart: echarts.ECharts | null = null
function renderChart() {
  if (!chartRef.value) return
  if (!chart) {
    chart = echarts.init(chartRef.value)
  }
  const days = Array.from({ length: 14 }, (_, i) => `D${i + 1}`)
  const dataA = days.map((_d, i) => Math.round(30 + Math.sin(i / 2) * 10 + Math.random() * 8))
  const dataB = days.map((_d, i) => Math.round(15 + Math.cos(i / 2) * 6 + Math.random() * 6))
  chart.setOption({
    tooltip: { trigger: 'axis' },
    grid: { left: 40, right: 20, top: 30, bottom: 30 },
    xAxis: { type: 'category', data: days, boundaryGap: false },
    yAxis: { type: 'value' },
    legend: { data: ['活跃用户', '新增角色'] },
    series: [
      { name: '活跃用户', type: 'line', smooth: true, areaStyle: {}, data: dataA },
      { name: '新增角色', type: 'line', smooth: true, data: dataB },
    ],
  })
}
function handleResize() { chart?.resize() }

async function loadProfile() {
  const res: any = await getProfile()
  profile.value = res?.data
}

async function loadStats() {
  statLoading.value = true
  try {
    // 仅取 total，不需要真实数据内容
    const [u, r, m] = await Promise.all([
      getUserList(1, 1),
      getRoleList(1, 1),
      getMenuList(1, 1),
    ])
    const userTotal = (u?.data?.total ?? 0) as number
    const roleTotal = (r?.data?.total ?? 0) as number
    const menuTotal = (m?.data?.total ?? 0) as number
    stats.value = [
      { label: '用户数', value: userTotal, link: '/system/user' },
      { label: '角色数', value: roleTotal, link: '/system/role' },
      { label: '菜单数', value: menuTotal, link: '/system/menu' },
    ]
  } catch {
    // 忽略统计失败，保持默认 0
  } finally {
    statLoading.value = false
  }
}

function go(link?: string) {
  if (link) router.push(link)
}

async function init() {
  try {
    loading.value = true
    await Promise.all([loadProfile(), loadStats()])
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
  nextTick(() => {
    renderChart()
    window.addEventListener('resize', handleResize)
  })
})

onBeforeUnmount(() => {
  window.removeEventListener('resize', handleResize)
  chart?.dispose(); chart = null
})
</script>

<template>
  <div class="dashboard">
    <a-row :gutter="16" class="mb16">
      <a-col :span="24">
        <a-card :loading="loading" title="欢迎回来">
          <div class="profile">
            <div>
              <div class="hello">你好，{{ profile?.username || '——' }}</div>
              <div class="sub">ID: {{ profile?.id || '——' }} · 邮箱: {{ profile?.email || '——' }}</div>
            </div>
          </div>
        </a-card>
      </a-col>
    </a-row>

    <a-row :gutter="16" class="mb16">
      <a-col :span="8" v-for="s in stats" :key="s.label">
        <a-card :loading="statLoading" :hoverable="!!s.link" @click="go(s.link)">
          <div class="stat">
            <div class="label">{{ s.label }}</div>
            <div class="value">{{ s.value }}</div>
          </div>
        </a-card>
      </a-col>
    </a-row>

    <a-row :gutter="16" class="mb16">
      <a-col :span="16">
        <a-card title="活跃趋势" :bordered="true">
          <div ref="chartRef" class="chart"></div>
        </a-card>
      </a-col>
      <a-col :span="8">
        <a-card title="快捷入口">
          <a-space direction="vertical" fill>
            <a-button long type="primary" @click="go('/system/user')">用户管理</a-button>
            <a-button long type="secondary" @click="go('/system/role')">角色管理</a-button>
            <a-button long type="outline" @click="go('/system/menu')">菜单管理</a-button>
          </a-space>
        </a-card>
      </a-col>
    </a-row>

    <a-row :gutter="16">
      <a-col :span="24">
        <a-card title="最近活动（示例）" :bordered="true">
          <a-empty description="暂无数据" />
        </a-card>
      </a-col>
    </a-row>
  </div>
  
</template>

<style scoped>
.dashboard { padding: 4px; }
.mb16 { margin-bottom: 16px; }
.profile { display: flex; align-items: center; justify-content: space-between; }
.hello { font-size: 18px; font-weight: 600; }
.sub { color: var(--color-text-3); margin-top: 4px; }
.stat { display: flex; align-items: baseline; justify-content: space-between; }
.stat .label { color: var(--color-text-2); }
.stat .value { font-size: 28px; font-weight: 700; }
.chart-placeholder { height: 220px; display: flex; align-items: center; justify-content: center; color: var(--color-text-3); background: var(--color-fill-2); border-radius: 6px; }
</style>

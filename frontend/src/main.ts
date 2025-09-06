import { createApp } from 'vue'
import App from './App.vue'
import router from './router'
import { initDynamicRoutes } from './router/dynamic'
import ArcoVue from '@arco-design/web-vue'
import '@arco-design/web-vue/dist/arco.css'
import './styles/index.less'

const app = createApp(App)
app.use(router)
app.use(ArcoVue)

// 先初始化动态路由，再挂载应用
initDynamicRoutes(router).finally(() => {
	app.mount('#app')
})

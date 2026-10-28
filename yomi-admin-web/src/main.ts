import { createApp } from 'vue'
import { createPinia } from 'pinia'

import App from './App.vue'
import router from './router'
import { setupRouterGuards } from './router/guard'
import { permission } from './directives/permission'

import '@/styles/index.scss'
// Element Plus 暗黑主题变量（配合 VueUse useDark 切换 html.dark）
import 'element-plus/theme-chalk/dark/css-vars.css'

const app = createApp(App)

app.use(createPinia())
// 守卫要在路由安装（首次导航）前注册，且依赖 pinia 已可用
setupRouterGuards(router)
app.use(router)

app.directive('permission', permission)

app.mount('#app')

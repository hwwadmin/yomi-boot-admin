import { createApp } from 'vue'
import { createPinia } from 'pinia'

import App from './App.vue'
import { i18n } from './i18n'
import router from './router'
import { setupRouterGuards } from './router/guard'
import { permission } from './directives/permission'

// Element Plus 暗黑主题变量（配合 VueUse useDark 切换 html.dark）
import 'element-plus/theme-chalk/dark/css-vars.css'
// 项目 token 与 Element Plus 覆盖必须排在暗色变量之后，保证 html.dark 下优先生效
import '@/styles/index.scss'

const app = createApp(App)

app.use(i18n)
app.use(createPinia())
// 守卫要在路由安装（首次导航）前注册，且依赖 pinia 已可用
setupRouterGuards(router)
app.use(router)

app.directive('permission', permission)

app.mount('#app')

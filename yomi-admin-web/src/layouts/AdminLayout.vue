<script setup lang="ts">
import { computed } from 'vue'

import logoUrl from '@/assets/logo.svg'
import { useAppStore } from '@/stores/app'
import Sidebar from './components/Sidebar.vue'
import Navbar from './components/Navbar.vue'

defineOptions({ name: 'AdminLayout' })

const appStore = useAppStore()

const sidebarWidth = computed(() => (appStore.sidebarCollapsed ? '64px' : '224px'))

// vite define 注入的版本号，模板需经组件作用域中转
const appVersion = __APP_VERSION__
</script>

<template>
  <el-container class="admin-layout">
    <el-aside class="layout-aside" :width="sidebarWidth">
      <div class="layout-logo" :class="{ 'is-collapsed': appStore.sidebarCollapsed }">
        <img class="logo-badge" :src="logoUrl" alt="" aria-hidden="true" />
        <span v-show="!appStore.sidebarCollapsed" class="logo-text">Yomi Admin</span>
      </div>
      <Sidebar />
      <!-- 侧边栏底部装饰波浪，不参与滚动与交互 -->
      <div class="aside-decor" aria-hidden="true"></div>
      <div v-show="!appStore.sidebarCollapsed" class="layout-version">v{{ appVersion }}</div>
    </el-aside>
    <el-container class="layout-body">
      <!-- 内容区右下装饰波浪，位于页面内容之下 -->
      <div class="body-decor body-decor--mid" aria-hidden="true"></div>
      <div class="body-decor body-decor--deep" aria-hidden="true"></div>
      <el-header class="layout-header" height="60px">
        <Navbar />
      </el-header>
      <el-main class="layout-main">
        <router-view />
      </el-main>
    </el-container>
  </el-container>
</template>

<style scoped>
.admin-layout {
  height: 100vh;
}

.layout-aside {
  position: relative;
  display: flex;
  flex-direction: column;
  /* clip 禁止一切滚动（含 scrollIntoView），防止装饰层溢出把布局"平移"走 */
  overflow: hidden;
  overflow: clip;
  background: linear-gradient(180deg, var(--app-sider-bg) 0%, var(--app-sider-bg-end) 100%);
  box-shadow: inset -1px 0 0 rgb(255 255 255 / 5%);
  transition: width 0.2s ease;
}

/* 侧边栏底部斜向波浪 */
.aside-decor {
  position: absolute;
  bottom: -18%;
  left: -30%;
  width: 170%;
  height: 34%;
  background: rgb(64 158 255 / 9%);
  border-radius: 50%;
  transform: rotate(-14deg);
  pointer-events: none;
}

.aside-decor::after {
  content: '';
  position: absolute;
  bottom: 12%;
  left: 10%;
  width: 100%;
  height: 100%;
  background: rgb(64 158 255 / 7%);
  border-radius: 50%;
  transform: rotate(6deg);
}

.layout-version {
  position: relative;
  flex-shrink: 0;
  padding: 0 20px 14px;
  color: rgb(255 255 255 / 38%);
  font-size: 12px;
  font-variant-numeric: tabular-nums;
  letter-spacing: 0.4px;
}

.layout-logo {
  display: flex;
  flex-shrink: 0;
  align-items: center;
  gap: 10px;
  height: var(--app-header-height);
  padding: 0 18px;
  overflow: hidden;
  color: #fff;
  white-space: nowrap;
}

.layout-logo.is-collapsed {
  justify-content: center;
  padding: 0;
}

.logo-badge {
  flex-shrink: 0;
  width: 30px;
  height: 30px;
  /* 深青图标压深色侧边栏，靠亮青渐变区与白色字母辨识，微提亮避免沉底 */
  filter: brightness(1.12) drop-shadow(0 2px 5px rgb(0 0 0 / 40%));
}

.logo-text {
  font-size: 15px;
  font-weight: 600;
  letter-spacing: 0.3px;
}

.layout-body {
  position: relative;
  /* clip 同 aside：装饰层溢出不可被程序化滚动 */
  overflow: hidden;
  overflow: clip;
  background-color: var(--app-bg);
}

/* 内容区右下两层波浪 */
.body-decor {
  position: absolute;
  border-radius: 50%;
  pointer-events: none;
  z-index: 0;
}

.body-decor--mid {
  right: -16%;
  bottom: -34%;
  width: 62%;
  height: 74%;
  background: rgb(64 158 255 / 7%);
  transform: rotate(-8deg);
}

.body-decor--deep {
  right: 6%;
  bottom: -42%;
  width: 48%;
  height: 62%;
  background: rgb(23 162 184 / 5%);
  transform: rotate(6deg);
}

.layout-header {
  position: relative;
  z-index: 1;
  padding: 0 var(--app-main-padding-x);
  background-color: var(--app-surface);
  border-bottom: 1px solid var(--app-border);
}

.layout-main {
  position: relative;
  z-index: 1;
  padding: var(--app-main-padding-y) var(--app-main-padding-x);
  overflow: auto;
  background-color: transparent;
}
</style>

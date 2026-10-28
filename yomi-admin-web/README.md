# yomi-admin-web

Yomi 管理后台前端：基于 Vite + Vue 3 + TypeScript 的现代化后台脚手架，内置登录鉴权、动态菜单、暗黑模式与本地 mock，可直接作为业务模块的开发起点。

## 快速开始

要求 Node.js ≥ 22.18（或 ≥ 24.12），包管理器使用 pnpm 10。

```bash
pnpm install      # 安装依赖
pnpm dev          # 开发服务器 http://localhost:5173（内置 mock，账号 admin / admin123）
pnpm build        # 类型检查 + 生产构建（产物 dist/）
pnpm lint         # oxlint + eslint（自动修复）
pnpm format       # prettier 格式化
```

## 技术栈

| 依赖 | 版本 | 说明 |
|---|---|---|
| vue | 3.5.x | `<script setup>` + TS |
| vite | 8.x | 构建工具 |
| element-plus | 2.14.x | 组件库，按需自动引入 |
| vue-router | 5.x | hash 模式（API 与 v4 兼容） |
| pinia | 4.x | 状态管理 |
| axios | 1.x | HTTP 请求（统一封装） |
| @vueuse/core | 15.x | 组合式工具集 |
| typescript / sass | 6.x / 1.106.x | — |

工程化：ESLint 10 + oxlint + Prettier。

## 目录结构

```
├── mock/index.ts            # 本地 mock（vite 中间件，仅开发环境）
├── .env.development         # 接口地址 / 代理目标 / mock 开关
└── src/
    ├── api/                 # 接口层（按模块组织，类型集中 types.ts）
    ├── utils/request.ts     # axios 封装（token、统一响应、401 处理）
    ├── stores/              # user / app / permission（动态路由预留）
    ├── router/              # 静态路由 + 守卫（标题、登录校验）
    ├── layouts/             # AdminLayout + Sidebar / Navbar / Breadcrumb
    ├── directives/          # v-permission 按钮级权限
    ├── styles/              # 全局基础样式
    ├── types/               # 路由 meta 扩展、自动导入声明
    └── views/               # login / dashboard / system/user(CRUD 模板) / error
```

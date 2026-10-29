# yomi-admin-web

Yomi 管理后台前端：基于 Vite + Vue 3 + TypeScript 的现代化后台脚手架，内置登录鉴权、RBAC 菜单与按钮权限、暗黑模式与本地 mock，可直接作为业务模块的开发起点。

## 快速开始

要求 Node.js ≥ 22.18（或 ≥ 24.12），包管理器使用 pnpm 10。

```bash
pnpm install      # 安装依赖
pnpm dev          # 开发服务器 http://localhost:25101（mock 开关见 .env.development）
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

## 权限模型

后端在登录与 profile 接口返回：

```json
{
  "username": "admin",
  "nickname": "超级管理员",
  "permissions": ["*"],
  "menus": [{ "code": "system:user", "name": "系统用户", "uiPath": "/system/user" }]
}
```

- `menus`：可见菜单的 `uiPath`，前端用它过滤静态路由和侧栏；页面是否存在仍由 `router/routes.ts` 决定。
- `permissions`：权限编码集合，`*` 表示全部；用于 `v-permission` 按钮级控制。
- 后端授权链路是 `用户 → 角色 → 权限组 → 权限`：角色页只分配权限组，
  权限组页维护组内权限，角色拥有的权限由权限组自动去重聚合，前端只读展示。
- 路由守卫在用户信息加载后判断 `canAccess(path)`，无权限跳 `/403`。
- 新增页面要同时补 `router/routes.ts` 的静态路由和后端权限目录（种子 SQL 或权限管理页），
  并把权限加入对应权限组。

## 目录结构

```
├── mock/index.ts            # 本地 mock（vite 中间件，仅开发环境，含 RBAC 接口）
├── .env.development         # 接口地址 / 代理目标 / mock 开关
└── src/
    ├── api/                 # 接口层（user / auth / rbac / httpRecord，类型集中 types.ts）
    ├── utils/request.ts     # axios 封装（token、统一响应、401 处理）
    ├── stores/              # user / app / permission（菜单过滤与访问判定）
    ├── router/              # routes.ts 静态路由 + guard.ts 登录与权限守卫
    ├── layouts/             # AdminLayout + Sidebar / Navbar / Breadcrumb
    ├── directives/          # v-permission 按钮级权限
    ├── styles/              # 全局基础样式
    ├── types/               # 路由 meta 扩展、自动导入声明
    └── views/               # login / dashboard / system/{user,role,permission,permission-group,http-record} / error
```

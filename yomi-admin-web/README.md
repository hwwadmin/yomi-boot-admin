# yomi-admin-web

管理后台前端脚手架：Vite + Vue 3 + TypeScript + Element Plus + Vue Router(hash) + Pinia + Axios + VueUse。

## 快速开始

要求 Node.js ≥ 22.18（或 ≥ 24.12）。

```bash
npm install       # 安装依赖
npm run dev       # 启动开发服务器 http://localhost:5173
npm run build     # 类型检查 + 生产构建
npm run preview   # 预览生产构建
npm run lint      # oxlint + eslint（自动修复）
npm run format    # prettier 格式化
npm run type-check
```

开发模式下默认启用内置 mock，演示账号：**admin / admin123**。

## 目录结构

```
├── mock/index.ts            # dev mock（vite 中间件，内存数据，仅本地开发）
├── env.d.ts                 # ImportMetaEnv 类型
├── .env.development         # API 路径 / 代理目标 / mock 开关
├── .env.production
└── src/
    ├── api/                 # 接口层：按模块一个文件，类型集中在 types.ts
    ├── utils/request.ts     # axios 封装（token 注入、统一解包、错误提示、401 处理）
    ├── stores/              # user(登录态) / app(侧栏、暗黑) / permission(动态路由预留)
    ├── router/              # 静态路由 + 全局守卫（标题、登录校验）
    ├── layouts/             # AdminLayout + Sidebar / Navbar / Breadcrumb
    ├── directives/          # v-permission 按钮级权限指令
    ├── styles/              # 全局样式（业务样式写在组件内）
    ├── types/               # 路由 meta 扩展 + 自动生成的声明文件
    └── views/               # login / dashboard / system/user(CRUD 模板) / error
```

## 接口约定（后端对齐用）

所有接口返回统一响应体，`code === 200` 表示成功：

```jsonc
{ "code": 200, "message": "ok", "data": { } }
```

- 非成功 code：前端拦截器统一 `ElMessage` 提示并 reject，调用方无需重复提示；
- `code === 401`：拦截器清空登录态并跳转登录页（带 `redirect` 回跳参数）；
- 分页接口：请求 `?page=&pageSize=&其他筛选`，响应 `data: { list, total, page, pageSize }`；
- 认证：登录后所有请求自动携带 `Authorization: Bearer <token>`。

api 层写法（`src/api/`）：

```ts
export function getUserPage(params: UserQuery) {
  return request<PageResult<UserItem>>({ url: '/system/users', method: 'get', params })
}
```

## 新增一个页面/模块

1. `src/api/xxx.ts` 增加接口函数（类型进 `src/api/types.ts`）；
2. `src/views/xxx/index.vue` 新建页面，CRUD 可复制 `src/views/system/user/index.vue` 作模板；
3. `src/router/index.ts` 的 `root.children` 里注册路由（`meta.title` 必填，`meta.icon` 选填），
   菜单自动生成，无需改布局；
4. 需要按钮级权限时使用 `v-permission="'module:action'"`。

## mock 与真实后端切换

- 开发期 mock：`.env.development` 中 `VITE_USE_MOCK=true`（默认），`/api/*` 由 mock 插件在
  本地拦截，数据存内存，重启 dev server 即重置；
- 联调真实后端：`VITE_USE_MOCK=false`，`/api` 代理到 `VITE_PROXY_TARGET`（默认
  `http://localhost:8080`）；
- 生产构建不含 mock；`VITE_API_BASE_URL` 控制真实接口前缀（默认 `/api`，由网关/nginx 转发）。

## 其他说明

- 路由为 **hash 模式**（管理后台无 SEO 需求，部署无需服务端 fallback）；
- Element Plus 组件与 Vue/Router/Pinia/VueUse API 均自动按需引入，免手写 import；
- 暗黑模式：顶栏切换，Element Plus dark css-vars + VueUse `useDark`，持久化到 localStorage；
- 动态路由（后端下发菜单）预留：见 `src/stores/permission.ts` 与 `src/router/guard.ts` 注释，
  接入时在守卫中 `generateRoutes()` + `router.addRoute()` 即可；
- Vue Router 当前为 v5 大版本，与 v4 的 API（`createRouter` / `createWebHashHistory` / 守卫等）
  完全兼容；如需锁定 v4：`npm i vue-router@4`。

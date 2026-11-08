# yomi-admin-web

Yomi 管理后台前端：Vue 3 + TypeScript + Vite + Element Plus。后端在 `../yomi-admin-backend`。

## 环境要求

- Node.js ≥ 22.18（或 ≥ 24.12）
- pnpm 10

## 启动

```bash
pnpm install
pnpm dev
```

打开 <http://localhost:25101>，用 `admin / admin123` 登录。默认连本地后端，先把 [后端](../yomi-admin-backend/README.md) 跑起来。

接口配置在 `.env.development`：

- `VITE_PROXY_TARGET`：dev 代理 `/api` 的后端地址，默认 `http://localhost:8080`。
- `VITE_USE_MOCK=true`：不连后端，改走内置 mock（mock 的登录也带验证码，按页面显示输入）。

## 常用命令

```bash
pnpm build    # 类型检查 + 生产构建（产物 dist/）
pnpm preview  # 本地预览构建产物
pnpm lint     # 错误码/语言包对账 + oxlint + eslint（自动修复）
pnpm format   # prettier 格式化 src/
```

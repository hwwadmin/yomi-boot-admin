# yomi-admin-backend

Yomi 管理后台后端：Go 1.26 + [yomi-boot](https://github.com/hwwadmin/yomi-boot)。前端在 `../yomi-admin-web`。

## 环境要求

- Go 1.26
- MySQL 8
- Redis 7

## 启动

1. 初始化数据库（先自行建库，库名与 `dev.yml` 的 `database` 一致，建议 utf8mb4；脚本不指定库名，执行时选中它。`init.sql` 建表，`seed_*.sql` 按领域灌种子数据，按文件名顺序执行，可重复执行）：

```bash
mysql -uroot -p -e "CREATE DATABASE IF NOT EXISTS yomi_admin DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci"
for f in script/sql/*.sql; do mysql -uroot -p -D yomi_admin < "$f"; done
```

2. 按需修改 `apps/server/config/dev.yml` 里的 MySQL / Redis 连接，默认是 `127.0.0.1:3306` 和 `127.0.0.1:6379`。

3. 启动服务：

```bash
cd apps/server
YOMI_CONFIG=dev go run ./cmd
```

服务默认监听 `8080`。

## 使用

健康检查：

```bash
curl http://127.0.0.1:8080/api/v1/health
```

内置超管账号 `admin / admin123`，登录拿 token：

```bash
curl -s -X POST http://127.0.0.1:8080/api/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"username":"admin","password":"admin123"}'
```

后续请求带上 `Authorization: Bearer <token>`。

- 接口文档（Swagger UI）：<http://127.0.0.1:8080/api/system/api-doc>
- 接口的 HTTP 状态码恒为 200，用响应里的 `code` 判断成败（`OK` 表示成功）。

## 常用命令

在 `yomi-admin-backend` 根目录执行：

```bash
make check  # fmt + vet + lint
make test   # 单元测试
```

## 更多

新增业务模块的目录骨架与步骤见 [新增业务模块流程.md](./新增业务模块流程.md)。

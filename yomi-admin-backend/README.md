# yomi-admin-backend

Yomi 管理后台后端（P0）：基于 Go 1.26 + [yomi-boot](https://gitee.com/hq_project/yomi-boot) v1.2.0，
分层参考 `yomi-boot-template/module-mono` 的清洁架构 + DDD-light。

## 当前范围（P0）

- 登录 / 当前用户 / 登出
- 系统用户分页、新增、编辑、删除
- 统一使用 yomi-boot 默认响应协议

暂不包含：RBAC、角色菜单、动态路由、登录/操作日志、部门、字典、密码重置。

## 目录

```text
apps/server/        编排层：cmd、config、bootstrap、handler、usecase/{auth,user}、test
domains/user/       用户域：entity/input/output、contract、provider、model/repo/service
script/sql/         建库与建表脚本
script/dev/         格式、vet、lint、测试脚本
```

usecase 按 handler 分组（`auth` / `user`），文件名统一为 `<action>_usecase.go`。

依赖方向：`handler → usecase → service_contract`；`provider → internal/service → repo_contract → repo + model`。
`model` 不出域，写路径由 usecase 通过 `tx.Executor` 开启事务。

## 依赖

```go
require gitee.com/hq_project/yomi-boot v1.2.0
```

`go.work` 固定 `go 1.26.0`（yomi-boot 的 go.mod 要求 `>= 1.26.0`）。

## 本地启动

1. 执行 SQL：

```bash
mysql -uroot -p < script/sql/001_init.sql
mysql -uroot -p < script/sql/010_user.sql
```

2. `dev.yml` 直接连接本地 Docker 提供的 MySQL / Redis（`127.0.0.1:3306` / `127.0.0.1:6379`），账号密码已写在配置里，dev 不需要设置任何环境变量。应用保留 `YOMI_ADMIN` 环境变量前缀；`prod.yml` 使用 `${YOMI_ADMIN_DB_PASSWORD}` / `${YOMI_ADMIN_REDIS_PASSWORD}` 占位。

后端统一使用 UTC：`app.timezone: UTC`，MySQL 连接串固定 `loc=UTC&time_zone='+00:00'`，种子脚本也会先 `SET time_zone = '+00:00'`。时间日期字段返回给前端时使用 **Unix 毫秒时间戳**，由前端按客户端所在时区转换展示。

配置只保留 `dev.yml` / `prod.yml`；`YOMI_CONFIG` 不填时默认 `dev`。

3. 启动：

```bash
cd apps/server
YOMI_CONFIG=dev go run ./cmd
```

健康检查：

```bash
curl http://127.0.0.1:8080/api/v1/health
```

种子账号：`admin / admin123`。

## 常用命令

```bash
make tidy     # go work sync + 各 module go mod tidy
make check    # fmt + vet + lint
make test     # 单元测试（排除 apps/server/test 集成目录）
make cross-build
```

集成测试需要先启动服务：

```bash
go test ./apps/server/test -v
```

## 响应协议

直接使用 yomi-boot `DefaultEncoder`，HTTP 恒为 200：

```json
{ "code": "OK", "msg": "success", "timestamp": 1761619200000, "data": {} }
```

分页：`data` 是列表，`total` 在顶层。失败：`{ "code": "ADMIN_USER_NOT_FOUND", "msg": "用户不存在", "timestamp": ... }`。

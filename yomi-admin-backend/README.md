# yomi-admin-backend

Yomi 管理后台后端：基于 Go 1.26 + [yomi-boot](https://gitee.com/hq_project/yomi-boot) v1.2.0.1，
分层参考 `yomi-boot-template/module-mono` 的清洁架构 + DDD-light。

## 当前范围

- 登录 / 当前用户 / 登出
- 系统用户分页、新增、编辑、删除
- RBAC：角色管理、权限目录、角色绑权限、用户绑角色
- 菜单与按钮权限下发，接口级鉴权中间件
- HTTP 访问记录：yomi-boot `http_record` 采集、进程内消费者落库、分页与详情查询
- 统一使用 yomi-boot 默认响应协议

暂不包含：动态路由、登录/操作日志、部门、字典、密码重置。

## 目录

```text
apps/server/        编排层：cmd、config、bootstrap、handler、middleware、usecase、support、test
domains/user/       用户域：entity/input/output、contract、provider、repo/model/mapper/service
domains/rbac/       角色权限域：enums/entity/input/output、contract、provider、repo/model/mapper/service
domains/httprecord/ 访问记录域：entity/enums/input/output、contract、provider、repo/model/mapper/service
script/sql/         建库、用户表、RBAC 表、访问记录表与种子
script/dev/         格式、vet、lint、测试脚本
```

usecase 按模块分组（`auth` / `user` / `rbac` / `httprecord`），文件名统一为 `<action>_usecase.go`，一个文件一个用例。
依赖方向：`handler → usecase → service_contract`；`usecase → support（浅层错误、跨域策略）`；
`bootstrap → provider + support`；`provider → internal/service → repo_contract → repo + model/mapper`。
`support` 是 usecase 层的复用逻辑：浅层错误，以及需要组合多个领域契约的策略
（超管授予/回收、最后一名启用超管保护）。`repo` 包内的持久化模型为非导出类型，只出现在仓储层；
多语句写路径由 usecase 通过 `tx.Executor` 开启事务。
`support` 属于 `apps/server` 入口私有；新增入口默认只复用领域，若涉及同样的超管不变量，
需要复用或实现同等策略，不能静默绕过。

## 数据约定

- 表名统一 `t_` 前缀，主键为应用层生成的雪花 ID，不使用自增。
- 状态、类型等闭集字段使用 `varchar` + 大写枚举值（`ENABLED` / `DISABLED`、`MENU` / `API`）。
- 所有表包含统一审计字段；角色和权限不做软删除，用状态停用；关系表物理删除。
- Entity 时间字段为 Unix 毫秒时间戳，HTTP 出入参 ID 使用 `dto.JSONInt64`（序列化为字符串）。
- 访问记录表 `t_http_record` 以 `trace_id + request_at` 去重，保证消费重投幂等；记录会持续增长，
  生产环境请按需定期归档或清理历史数据。

## 权限模型

- 角色：`t_rbac_role`，内置 `super_admin` 不可禁用、不可绑定权限，判定时短路放行。
- 权限目录：`t_rbac_permission`，`MENU` 行对应前端路由路径，`API` 行对应 gin 路由模板（如 `/api/system/users/:id`）。
- 关系：`t_rbac_user_role_rel`、`t_rbac_role_permission_rel`。
- 鉴权：请求进入后按 `HTTP method + c.FullPath()` 精确匹配 API 权限；超管直接放行。
- 菜单：登录 / profile 返回用户可见菜单，前端用它过滤静态路由与侧栏。
- 权限目录由 SQL 种子维护，管理端只读；新增路由必须同时补 `020_rbac.sql` 与权限编码，否则
  `rbac_route_catalog_test.go` 会失败。
- 保护规则：内置角色不可变更；只有超管能授予或回收超管角色；不能解绑自己的超管角色；
  不能禁用/删除最后一名启用状态的超管。这些跨域授权策略位于
  `apps/server/internal/support/superadmin`，角色/权限自身的格式、状态与关联完整性校验位于 rbac 领域服务。

## 依赖

```go
require gitee.com/hq_project/yomi-boot v1.2.1-0.20261028113112-1a7295dd465e
```

Gitee tag `v1.2.0.1` 不是三段式 semver，Go 会把它解析为上面的伪版本（commit `1a7295dd465e`）。

`go.work` 固定 `go 1.26.0`（yomi-boot 的 go.mod 要求 `>= 1.26.0`），工作区包含
`apps/server`、`domains/user`、`domains/rbac`、`domains/httprecord`。

## 访问记录

- 采集：yomi-boot `http_record` 中间件，配置在 `config/*.yml` 的 `http_record` 段；`provider: memory`
  表示记录先入进程内队列，由 httprecord 插件在 `OnReady` 阶段启动内存消费者。
- 消费：业务在 `Wire` 中通过 `httprecordplugin.From(slot).MustRegisterHandler(...)` 注入落库处理函数
  （见 `internal/bootstrap/http_record_handler.go`）；`memory` 恒消费，`consumer.enabled` 只对
  `kafka` / `redis_stream` 生效。Kafka 可用 `consumer.max_retry` 控制同一记录的重试次数（默认 3，超限丢弃）；
  `consumer.group_id` 留空时按 `app.name + "-httprecord"` 派生。
- 落库：`t_http_record`；`record_request_body` / `record_response_body` 控制是否采集正文，
  正文与请求头统一经 `masking` 段脱敏（密码、令牌等不会明文落库）。
- 查询：`GET /api/system/http-records`（`keyword` / `method` / `statusClass` / `userId` /
  `startTime` / `endTime` 过滤）与 `GET /api/system/http-records/:id`；菜单权限 `system:http-record`。
- 如需跨进程可靠投递，可把 `http_record.provider` 切到 `kafka` 或 `redis_stream`，同时启用对应插件；
  消费者由 httprecord 插件统一托管，无需改业务代码。

## 本地启动

1. 执行 SQL：

```bash
mysql -uroot -p < script/sql/001_init.sql
mysql -uroot -p < script/sql/010_user.sql
mysql -uroot -p < script/sql/020_rbac.sql
mysql -uroot -p < script/sql/030_http_record.sql
```

`020_rbac.sql` 会创建 RBAC 四张表、写入权限目录种子，并把 `admin` 绑定为内置超管。
`030_http_record.sql` 创建访问记录表；升级已有环境时重跑 `020` + `030` 即可（种子使用 `INSERT IGNORE`）。
`010_user.sql` 会在建好 `t_user` 后删除旧版 `sys_user` 表。
**必须先执行 SQL 再启动服务**：鉴权中间件依赖角色关系，缺表会导致受保护接口报错。

> 脚本内已加 `SET NAMES utf8mb4`；如果之前用非 utf8mb4 客户端导入过，表/字段注释会乱码，
> 需要先 `DROP` 对应表再重新执行脚本（数据行可以通过接口或种子重建）。

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

种子账号：`admin / admin123`（内置超管，拥有全部权限）。

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
无权限：`{ "code": "ADMIN_PERMISSION_DENIED", "msg": "没有访问权限", ... }`。

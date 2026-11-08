# yomi-admin-backend

Yomi 管理后台后端，Go 1.26 + [yomi-boot](https://github.com/hwwadmin/yomi-boot)。
前端在 `../yomi-admin-web`。

分层约定见同目录的 [清洁架构与DDD-light方法论.md](./清洁架构与DDD-light方法论.md)，这里只写怎么用、怎么跑。


## 本地启动

1. 建库、建表、灌种子：

```bash
mysql -uroot -p < script/sql/001_init.sql
mysql -uroot -p < script/sql/010_user.sql
mysql -uroot -p < script/sql/020_rbac.sql
mysql -uroot -p < script/sql/030_http_record.sql
mysql -uroot -p < script/sql/040_sys_param.sql
mysql -uroot -p < script/sql/050_audit_log.sql
mysql -uroot -p < script/sql/060_sys_biz.sql
```

`020_rbac.sql` 会建 RBAC 表、写权限目录与种子权限组，并把 `admin` 设成内置超管。
`040_sys_param.sql` 建系统参数表（`t_sys_param`），参数管理页与业务读参数都依赖它。
`050_audit_log.sql` 建操作审计表（`t_audit_log`），登录、改密、重置密码的审计写入与审计日志页都依赖它。
`060_sys_biz.sql` 建业务幂等表（`sys_biz`）。
`sys_biz` 是创建用户/角色/权限/权限组/参数与改密的幂等记录表，缺表时携带 `X-Idempotency-Key` 的请求会报错。
**先执行 SQL 再启动服务**，否则受保护接口会因为缺表报错。

已有环境升级时：补执行 `050_audit_log.sql` 与 `060_sys_biz.sql`，并重跑一次 `020_rbac.sql`（新增了审计日志的菜单与接口权限）；
审计表缺失时写入是"尽力而为"降级——业务照常跑，只在服务日志里打"审计写入失败"告警，审计日志页会报错。

2. 配置：

- `dev.yml` 已写好本地 MySQL / Redis（`127.0.0.1:3306`、`127.0.0.1:6379`），dev 不用设环境变量。
- `prod.yml` 用 `${YOMI_ADMIN_DB_PASSWORD}`、`${YOMI_ADMIN_REDIS_PASSWORD}` 占位。
- `YOMI_CONFIG` 不填时默认 `dev`。
- 后端统一 UTC；接口里的时间用 Unix 毫秒，前端按本地时区展示。

3. 启动：

```bash
cd apps/server
YOMI_CONFIG=dev go run ./cmd
```

健康检查：`curl http://127.0.0.1:8080/api/v1/health`

种子账号：`admin / admin123`（内置超管，拥有全部权限）。

## 目录

```text
apps/server/        接口与用例：cmd、config、internal/{handler,usecase,bootstrap,support,config}
domains/user/       用户
domains/rbac/       角色、权限组与权限
domains/httprecord/ 访问记录
domains/sysparam/   系统参数
domains/auditlog/   操作审计日志
persist/            各域共用的表字段与仓储辅助
script/sql/         建库与种子 SQL
script/dev/         开发脚本
```

简单说：`handler` 只接参数；`usecase` 管流程、事务和鉴权；`domain` 管业务规则和存储；
`support` 放多个模块共用的小规则，比如超管保护。五个 domain 都是独立的 Go module，可以单独复用。
表的主键、审计字段和软删除列在 `persist/orm`，各域仓储嵌入它。

访问记录字段 `traceParent` 存本跳 W3C `traceparent`。链路头是 `traceparent` / `tracestate`。

装配按上下文拆分：`internal/bootstrap/wire.go` 只列模块与跨模块依赖，各上下文的 Handler 构造
放在 `wire_<context>.go`；新增接口只改对应上下文文件，不要往中心函数里堆构造代码。

## HTTP 接口约定

- 对外接口只用 `GET` 和 `POST`：查询用 `GET`（参数走 Query），变更用 `POST`（参数走 JSON Body）。
- 不使用路径参数，ID 放在 Query 或 Body 里；路径资源段用单数名词。
- 变更接口统一为 `POST /api/{resource}/{action}`，ID 用 Body 里的 `id` / `{resource}Id`。
- 查询接口：分页 `GET /api/{resource}`，详情 `GET /api/{resource}/detail?id=`。

## API 文档（OpenAPI 3.1 + Swagger UI）

- 文档源：`apps/server/api/openapi.yaml`（OpenAPI 3.1，手写维护，随二进制内嵌）。
- 页面入口：`GET /api/system/api-doc`（Swagger UI）；schema：`GET /api/system/api-doc/schema`。
- 开关：`biz.docs.enabled`，dev 默认开、prod 默认关（可用 `YOMI_ADMIN_BIZ_DOCS_ENABLED` 覆盖）；
  四条文档路由已加进 `token.skip_validate_paths` 免鉴权（与 RBAC 放行清单共用）。
- 新增/下线路由必须同步修改 `openapi.yaml`：`internal/handler/api_doc_coverage_test.go`
  对文档与已注册路由双向对账，`api/embed_test.go` 校验文档是合法的 3.1 文档。
- 更换 Swagger UI 资源版本：`bash script/dev/fetch_swagger_ui.sh [版本号]`（当前内嵌 5.34.0）。


## 数据约定

- 表名统一 `t_` 前缀；主键是应用生成的雪花 ID。
- 状态/类型用 `varchar` + 大写枚举：`ENABLED`/`DISABLED`、`MENU`/`API`。
- 角色、权限组、权限物理删除；用户仍是软删除，删除用户时同步清理其角色绑定。
- 关系表物理删除，不留孤儿关系。
- Entity 时间是 Unix 毫秒；接口里的 ID 用 `dto.JSONInt64`（字符串）。

## 鉴权缓存（RBAC 投影）

鉴权热路径不查库：服务把六张 RBAC 表压缩成一份全量快照写进 Redis
（`yomi:admin:rbac:projection`），进程内再保留一份最近读到的快照（默认每 5 秒
从 Redis 同步一次、原子切换指针，`biz.rbac.cache.memory_sync_seconds`，设为 0 关闭），
命中内存的请求不再访问 Redis。内存未就绪时降级为一次性读 Redis，
Redis 不可用 / 内容损坏 / 结构版本不符时再降级为 SQL 直查，
因此缓存故障只影响性能，不会误判权限。

- **写操作不会自动重建投影**：改动最迟一个刷新周期（默认 60 秒）生效；要立即生效就在角色页点
  「刷新权限缓存」（`POST /api/system/rbac-cache/refresh`），执行刷新的实例立即生效，
  其它实例最迟一个内存同步周期（默认 5 秒）内生效。
- 定时刷新默认每 60 秒跑一次（`biz.rbac.cache.refresh_interval_seconds`，设为 0 关闭），
  同时兜住"直接改库"这类绕过接口的变更；多实例靠分布式锁串行，只有一台真正重建。
- 启动时先探测 Redis 里的投影：存在且可解析就复用，缺失 / 损坏 / 结构版本不符才从数据库重建
  （`biz.rbac.cache.bootstrap_warmup`），滚动重启不会覆盖已有投影。
- 投影结构变更时递增 `domains/rbac/internal/snapshot` 的 `Schema`，存量投影会被自动判定为不可用并重建。

配置在 `biz.rbac.cache`：`enabled`、`key`、`lock_key`、`bootstrap_warmup`、
`refresh_interval_seconds`、`projection_max_bytes`、`lock_wait_seconds`。

## 常用命令

```bash
make tidy         # 同步 go.work，整理各 module 依赖
make check        # fmt + vet + lint
make test         # 单元测试
make cross-build  # 交叉编译
```

集成测试要先启动服务：`go test ./apps/server/test -v`。服务启用了登录验证码时，集成测试
会在登录处跳过；需要完整跑通时以 `YOMI_ADMIN_BIZ_CAPTCHA_ENABLED=false` 启动服务。

## 响应格式

统一用 yomi-boot 默认格式，HTTP 状态码恒为 200：

```json
{ "code": "OK", "msg": "success", "timestamp": 1761619200000, "data": {} }
```

- 分页：`data` 是列表，`total` 在顶层。
- 失败：`{ "code": "ADMIN_USER_NOT_FOUND", "msg": "用户不存在", ... }`。
- 无权限：`{ "code": "ADMIN_PERMISSION_DENIED", "msg": "没有访问权限", ... }`。

## 依赖

- Go 1.26。
- yomi-boot v1.0.2-rc.1（github.com/hwwadmin/yomi-boot）。
- `go.work` 包含 `apps/server`、五个 domain module，以及 `persist`（各域共用的表字段与仓储辅助；boot 不再预置表结构）。

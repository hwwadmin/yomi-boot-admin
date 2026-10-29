# yomi-admin-backend

Yomi 管理后台后端，Go 1.26 + [yomi-boot](https://gitee.com/hq_project/yomi-boot)。
前端在 `../yomi-admin-web`。

分层约定见同目录的 [清洁架构与DDD-light方法论.md](./清洁架构与DDD-light方法论.md)，这里只写怎么用、怎么跑。

## 能做什么

- 登录（含图像验证码）、当前用户、登出
- 用户管理：分页、新增、编辑、删除
- RBAC：角色、权限目录、权限组，角色绑权限组、权限组绑权限、用户绑角色
- 菜单与按钮权限下发，接口级鉴权
- HTTP 访问记录：采集、落库、分页与详情
- 统一响应格式

暂不包含：动态路由、登录/操作日志、部门、字典、密码重置。

## 本地启动

1. 建库、建表、灌种子：

```bash
mysql -uroot -p < script/sql/001_init.sql
mysql -uroot -p < script/sql/010_user.sql
mysql -uroot -p < script/sql/020_rbac.sql
mysql -uroot -p < script/sql/030_http_record.sql
```

`020_rbac.sql` 会建 RBAC 表、写权限目录与种子权限组，并把 `admin` 设成内置超管。
**先执行 SQL 再启动服务**，否则受保护接口会因为缺表报错。

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
script/sql/         建库与种子 SQL
script/dev/         开发脚本
```

简单说：`handler` 只接参数；`usecase` 管流程、事务和鉴权；`domain` 管业务规则和存储；
`support` 放多个模块共用的小规则，比如超管保护。三个 domain 都是独立的 Go module，可以单独复用。

## HTTP 接口约定

- 对外接口只用 `GET` 和 `POST`：查询用 `GET`（参数走 Query），变更用 `POST`（参数走 JSON Body）。
- 不使用路径参数，ID 放在 Query 或 Body 里；路径资源段用单数名词。
- 变更接口统一为 `POST /api/{resource}/{action}`，ID 用 Body 里的 `id` / `{resource}Id`。
- 查询接口：分页 `GET /api/{resource}`，详情 `GET /api/{resource}/detail?id=`。


## 登录验证码

- 登录前需要图形验证码（yomi-boot `component/common/captcha`）：`GET /api/auth/captcha`
  返回 `{ id, image, expiresIn }`，登录入参带 `captchaId`、`captchaCode`；答案只存 Redis
  摘要，一次性消费。
- 同一 IP 刷新验证码有最小间隔，配置在 `biz.captcha.refresh_interval_seconds`（默认 10 秒，
  0 表示不限流），超限返回 `ADMIN_CAPTCHA_TOO_FREQUENT`；限流由 `ratelimit` 插件提供，
  需要 `ratelimit.enabled=true`。部署在反向代理后时要配置 `http.trusted_proxies`，
  否则所有请求都会按代理 IP 限流。
- 本地跑集成测试时用 `YOMI_ADMIN_BIZ_CAPTCHA_ENABLED=false` 关闭校验。

## 超管与 Token

- 超管变更（绑角色、禁用、删除）会先抢一把 Redis 分布式锁，key 是
  `yomi-admin:superadmin:guard`。抢不到返回 `ADMIN_SUPER_ADMIN_BUSY`；Redis 异常返回
  `ADMIN_SUPER_ADMIN_LOCK_UNAVAILABLE`，稍后重试。
- 以后新增入口时，只要它也会改超管状态，就必须用同一个锁 key；否则并发时超管保护会失效。
- 禁用/删除用户后会撤销该用户全部 Token。撤销失败最多尝试 3 次；仍然失败返回
  `ADMIN_TOKEN_REVOKE_FAILED`。这时用户状态已经改好，只是 Token 没撤掉，需要重试。
- 超管保护：只有超管能授予/回收超管角色；不能解绑自己的超管；不能禁用/删除最后一名启用超管。
- 内置 `super_admin` 不可禁用、不可改权限。

## 权限模型

- 授权链路固定为 `用户 → 角色 → 权限组 → 权限`，角色不直连权限：
  角色表 `t_rbac_role`、权限组表 `t_rbac_permission_group`、权限表 `t_rbac_permission`，
  关系表 `t_rbac_user_role_rel`、`t_rbac_role_group_rel`、`t_rbac_group_permission_rel`。
- 角色拥有的权限 = 所绑权限组内权限的去重并集，仓储层用 `DISTINCT` 聚合，
  同一个权限无论被多少个角色、多少个权限组覆盖都只算一次。
- 鉴权同时要求角色、权限组、权限三者都启用；停用权限组会立刻对使用它的所有角色失效
  （权限不写入 Token，下一个请求即生效）。
- 角色详情里的 `permissionIds` 是结构聚合（只看绑定关系，不看启停），
  用于管理端展示；运行时鉴权会再叠加状态过滤。
- `MENU` 权限对应前端路由；`API` 权限对应 `HTTP method + 路由模板`，
  例如 `GET /api/system/user`。鉴权按请求方法 + `c.FullPath()` 匹配 API 权限；
  内置超管 `super_admin` 直接放行，不依赖权限组。
- 登录和 profile 返回用户可见菜单，前端用它过滤路由和侧栏。
- 权限目录由 SQL 种子维护，管理端只读。新增路由要同时补 `020_rbac.sql` 的权限与
  权限组关系，否则 `rbac_route_catalog_test.go` 和 `TestRBACSeedPermissionGroupCoverage` 会失败。
- 种子内置 `system_user`、`system_role`、`system_permission`、`system_http_record`
  四个权限组，可直接组合成业务角色。

## 数据约定

- 表名统一 `t_` 前缀；主键是应用生成的雪花 ID。
- 状态/类型用 `varchar` + 大写枚举：`ENABLED`/`DISABLED`、`MENU`/`API`。
- 角色、权限组、权限不软删除，用状态停用；关系表物理删除。
- Entity 时间是 Unix 毫秒；接口里的 ID 用 `dto.JSONInt64`（字符串）。

## 访问记录

- yomi-boot 的 `http_record` 负责采集，默认先进进程内队列，再落库到 `t_http_record`。
- 默认 `provider: memory`：进程内队列 + 后台消费者；`consumer.enabled` 只对
  `kafka` / `redis_stream` 生效。
- 需要跨进程可靠投递时，把 `http_record.provider` 改成 `kafka` 或 `redis_stream`，
  并启用对应插件。
- 请求体、响应体、请求头按 `masking` 配置脱敏，密码和令牌不会明文落库。
- 落库按 `trace_id + request_at` 去重，消费重投不会重复写入。
- 查询接口：`GET /api/system/http-record`、`GET /api/system/http-record/detail?id=`，
  菜单权限 `system:http-record`。记录会一直增长，生产环境记得定期归档或清理。

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
- yomi-boot v1.2.1。
- `go.work` 包含 `apps/server` 和三个 domain module。

# yomi-admin-backend

Yomi 管理后台后端，Go 1.26 + [yomi-boot](https://gitee.com/hq_project/yomi-boot)。
前端在 `../yomi-admin-web`。

分层约定见同目录的 [清洁架构与DDD-light方法论.md](./清洁架构与DDD-light方法论.md)，这里只写怎么用、怎么跑。


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


## 数据约定

- 表名统一 `t_` 前缀；主键是应用生成的雪花 ID。
- 状态/类型用 `varchar` + 大写枚举：`ENABLED`/`DISABLED`、`MENU`/`API`。
- 角色、权限组、权限物理删除；用户仍是软删除，删除用户时同步清理其角色绑定。
- 关系表物理删除，不留孤儿关系。
- Entity 时间是 Unix 毫秒；接口里的 ID 用 `dto.JSONInt64`（字符串）。

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

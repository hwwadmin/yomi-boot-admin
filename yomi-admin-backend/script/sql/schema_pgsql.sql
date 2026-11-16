-- yomi-admin-backend 表结构初始化（PostgreSQL 版）：只含 DDL，可重复执行
-- MySQL 版见 schema_mysql.sql，两份字段口径一致。
-- 种子数据按领域拆在 seed_*.sql，按文件名顺序在本文件之后执行（种子目前为 MySQL 方言）：
--   seed_10_user.sql  用户域（内置管理员）
--   seed_20_rbac.sql  RBAC 域（角色 / 权限目录 / 权限组及关系）
-- 不在此指定库名：先自行建库（ENCODING UTF8，库名与 dev.yml 的 database 一致），
-- 执行时选中它，例如：psql -U postgres -d yomi_admin -f script/sql/schema_pgsql.sql
-- update_time 的 MySQL ON UPDATE 语义由应用层（GORM autoUpdateTime）维护，PG 侧不设触发器。
-- PostgreSQL 的索引/约束名在同一 schema 内必须唯一，除 t_sys_biz 外统一加表名前缀；
-- t_sys_biz 沿用 yomi-boot 的 sql/pgsql/t_sys_biz.sql 命名，避免与框架脚本重复建索引。

-- ============================================================
-- 用户
-- ============================================================

CREATE TABLE IF NOT EXISTS t_user (
    id            BIGINT       NOT NULL,
    username      VARCHAR(64)  NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    nickname      VARCHAR(64)  NOT NULL DEFAULT '',
    email         VARCHAR(128) NOT NULL DEFAULT '',
    phone         VARCHAR(32)  NOT NULL DEFAULT '',
    status        VARCHAR(20)  NOT NULL DEFAULT 'ENABLED',
    create_time   TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    create_by     BIGINT       DEFAULT NULL,
    update_time   TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    update_by     BIGINT       DEFAULT NULL,
    delete_time   TIMESTAMP    DEFAULT NULL,
    delete_by     BIGINT       DEFAULT NULL,
    PRIMARY KEY (id),
    CONSTRAINT uk_t_user_username UNIQUE (username)
);

CREATE INDEX IF NOT EXISTS idx_t_user_delete_time ON t_user (delete_time);

COMMENT ON TABLE t_user IS '用户';
COMMENT ON COLUMN t_user.id IS '主键（雪花ID）';
COMMENT ON COLUMN t_user.username IS '登录名';
COMMENT ON COLUMN t_user.password_hash IS 'bcrypt 哈希';
COMMENT ON COLUMN t_user.nickname IS '昵称';
COMMENT ON COLUMN t_user.email IS '邮箱';
COMMENT ON COLUMN t_user.phone IS '手机号';
COMMENT ON COLUMN t_user.status IS '状态：ENABLED-启用，DISABLED-禁用';
COMMENT ON COLUMN t_user.create_time IS '创建时间';
COMMENT ON COLUMN t_user.create_by IS '创建人ID';
COMMENT ON COLUMN t_user.update_time IS '更新时间';
COMMENT ON COLUMN t_user.update_by IS '更新人ID';
COMMENT ON COLUMN t_user.delete_time IS '删除时间';
COMMENT ON COLUMN t_user.delete_by IS '删除人ID';

-- ============================================================
-- RBAC
-- ============================================================

CREATE TABLE IF NOT EXISTS t_rbac_role (
    id          BIGINT       NOT NULL,
    code        VARCHAR(64)  NOT NULL,
    name        VARCHAR(64)  NOT NULL,
    builtin     BOOLEAN      NOT NULL DEFAULT FALSE,
    status      VARCHAR(20)  NOT NULL DEFAULT 'ENABLED',
    remark      VARCHAR(255) NOT NULL DEFAULT '',
    create_time TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    create_by   BIGINT       DEFAULT NULL,
    update_time TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    update_by   BIGINT       DEFAULT NULL,
    delete_time TIMESTAMP    DEFAULT NULL,
    delete_by   BIGINT       DEFAULT NULL,
    PRIMARY KEY (id),
    CONSTRAINT uk_t_rbac_role_code UNIQUE (code)
);

CREATE INDEX IF NOT EXISTS idx_t_rbac_role_status ON t_rbac_role (status);

COMMENT ON TABLE t_rbac_role IS '角色';
COMMENT ON COLUMN t_rbac_role.id IS '主键（雪花ID）';
COMMENT ON COLUMN t_rbac_role.code IS '角色编码，全局唯一；super_admin 为内置超管';
COMMENT ON COLUMN t_rbac_role.name IS '角色名称';
COMMENT ON COLUMN t_rbac_role.builtin IS '是否内置：false-否，true-是';
COMMENT ON COLUMN t_rbac_role.status IS '状态：ENABLED-启用，DISABLED-禁用';
COMMENT ON COLUMN t_rbac_role.remark IS '备注';
COMMENT ON COLUMN t_rbac_role.create_time IS '创建时间';
COMMENT ON COLUMN t_rbac_role.create_by IS '创建人ID';
COMMENT ON COLUMN t_rbac_role.update_time IS '更新时间';
COMMENT ON COLUMN t_rbac_role.update_by IS '更新人ID';
COMMENT ON COLUMN t_rbac_role.delete_time IS '删除时间';
COMMENT ON COLUMN t_rbac_role.delete_by IS '删除人ID';

CREATE TABLE IF NOT EXISTS t_rbac_permission (
    id              BIGINT       NOT NULL,
    code            VARCHAR(96)  NOT NULL,
    name            VARCHAR(64)  NOT NULL,
    permission_type VARCHAR(20)  NOT NULL,
    ui_path         VARCHAR(191) DEFAULT NULL,
    http_method     VARCHAR(16)  DEFAULT NULL,
    http_path       VARCHAR(191) DEFAULT NULL,
    status          VARCHAR(20)  NOT NULL DEFAULT 'ENABLED',
    remark          VARCHAR(255) NOT NULL DEFAULT '',
    create_time     TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    create_by       BIGINT       DEFAULT NULL,
    update_time     TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    update_by       BIGINT       DEFAULT NULL,
    delete_time     TIMESTAMP    DEFAULT NULL,
    delete_by       BIGINT       DEFAULT NULL,
    PRIMARY KEY (id),
    CONSTRAINT uk_t_rbac_permission_code UNIQUE (code),
    CONSTRAINT uk_t_rbac_permission_http_route UNIQUE (http_method, http_path),
    CONSTRAINT uk_t_rbac_permission_ui_path UNIQUE (ui_path)
);

CREATE INDEX IF NOT EXISTS idx_t_rbac_permission_type_status ON t_rbac_permission (permission_type, status);

COMMENT ON TABLE t_rbac_permission IS '权限目录';
COMMENT ON COLUMN t_rbac_permission.id IS '主键（雪花ID）';
COMMENT ON COLUMN t_rbac_permission.code IS '权限编码，全局唯一';
COMMENT ON COLUMN t_rbac_permission.name IS '权限名称';
COMMENT ON COLUMN t_rbac_permission.permission_type IS '权限类型：MENU-菜单，API-接口';
COMMENT ON COLUMN t_rbac_permission.ui_path IS '菜单路径，permission_type=MENU 时必填';
COMMENT ON COLUMN t_rbac_permission.http_method IS 'HTTP 方法，permission_type=API 时必填';
COMMENT ON COLUMN t_rbac_permission.http_path IS 'gin 路由模板（以 / 开头，通常含 /api），permission_type=API 时必填';
COMMENT ON COLUMN t_rbac_permission.status IS '状态：ENABLED-启用，DISABLED-禁用';
COMMENT ON COLUMN t_rbac_permission.remark IS '备注';
COMMENT ON COLUMN t_rbac_permission.create_time IS '创建时间';
COMMENT ON COLUMN t_rbac_permission.create_by IS '创建人ID';
COMMENT ON COLUMN t_rbac_permission.update_time IS '更新时间';
COMMENT ON COLUMN t_rbac_permission.update_by IS '更新人ID';
COMMENT ON COLUMN t_rbac_permission.delete_time IS '删除时间';
COMMENT ON COLUMN t_rbac_permission.delete_by IS '删除人ID';

CREATE TABLE IF NOT EXISTS t_rbac_permission_group (
    id          BIGINT       NOT NULL,
    code        VARCHAR(64)  NOT NULL,
    name        VARCHAR(64)  NOT NULL,
    status      VARCHAR(20)  NOT NULL DEFAULT 'ENABLED',
    remark      VARCHAR(255) NOT NULL DEFAULT '',
    create_time TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    create_by   BIGINT       DEFAULT NULL,
    update_time TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    update_by   BIGINT       DEFAULT NULL,
    delete_time TIMESTAMP    DEFAULT NULL,
    delete_by   BIGINT       DEFAULT NULL,
    PRIMARY KEY (id),
    CONSTRAINT uk_t_rbac_permission_group_code UNIQUE (code)
);

CREATE INDEX IF NOT EXISTS idx_t_rbac_permission_group_status ON t_rbac_permission_group (status);

COMMENT ON TABLE t_rbac_permission_group IS '权限组';
COMMENT ON COLUMN t_rbac_permission_group.id IS '主键（雪花ID）';
COMMENT ON COLUMN t_rbac_permission_group.code IS '权限组编码，全局唯一';
COMMENT ON COLUMN t_rbac_permission_group.name IS '权限组名称';
COMMENT ON COLUMN t_rbac_permission_group.status IS '状态：ENABLED-启用，DISABLED-禁用';
COMMENT ON COLUMN t_rbac_permission_group.remark IS '备注';
COMMENT ON COLUMN t_rbac_permission_group.create_time IS '创建时间';
COMMENT ON COLUMN t_rbac_permission_group.create_by IS '创建人ID';
COMMENT ON COLUMN t_rbac_permission_group.update_time IS '更新时间';
COMMENT ON COLUMN t_rbac_permission_group.update_by IS '更新人ID';
COMMENT ON COLUMN t_rbac_permission_group.delete_time IS '删除时间';
COMMENT ON COLUMN t_rbac_permission_group.delete_by IS '删除人ID';

CREATE TABLE IF NOT EXISTS t_rbac_user_role_rel (
    id          BIGINT    NOT NULL,
    user_id     BIGINT    NOT NULL,
    role_id     BIGINT    NOT NULL,
    create_time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    create_by   BIGINT    DEFAULT NULL,
    update_time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    update_by   BIGINT    DEFAULT NULL,
    delete_time TIMESTAMP DEFAULT NULL,
    delete_by   BIGINT    DEFAULT NULL,
    PRIMARY KEY (id),
    CONSTRAINT uk_t_rbac_user_role_rel_user_role UNIQUE (user_id, role_id)
);

CREATE INDEX IF NOT EXISTS idx_t_rbac_user_role_rel_role_id ON t_rbac_user_role_rel (role_id);

COMMENT ON TABLE t_rbac_user_role_rel IS '用户-角色关系';
COMMENT ON COLUMN t_rbac_user_role_rel.id IS '主键（雪花ID）';
COMMENT ON COLUMN t_rbac_user_role_rel.user_id IS '用户ID（t_user.id）';
COMMENT ON COLUMN t_rbac_user_role_rel.role_id IS '角色ID';
COMMENT ON COLUMN t_rbac_user_role_rel.create_time IS '创建时间';
COMMENT ON COLUMN t_rbac_user_role_rel.create_by IS '创建人ID';
COMMENT ON COLUMN t_rbac_user_role_rel.update_time IS '更新时间';
COMMENT ON COLUMN t_rbac_user_role_rel.update_by IS '更新人ID';
COMMENT ON COLUMN t_rbac_user_role_rel.delete_time IS '删除时间';
COMMENT ON COLUMN t_rbac_user_role_rel.delete_by IS '删除人ID';

CREATE TABLE IF NOT EXISTS t_rbac_role_group_rel (
    id          BIGINT    NOT NULL,
    role_id     BIGINT    NOT NULL,
    group_id    BIGINT    NOT NULL,
    create_time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    create_by   BIGINT    DEFAULT NULL,
    update_time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    update_by   BIGINT    DEFAULT NULL,
    delete_time TIMESTAMP DEFAULT NULL,
    delete_by   BIGINT    DEFAULT NULL,
    PRIMARY KEY (id),
    CONSTRAINT uk_t_rbac_role_group_rel_role_group UNIQUE (role_id, group_id)
);

CREATE INDEX IF NOT EXISTS idx_t_rbac_role_group_rel_group_id ON t_rbac_role_group_rel (group_id);

COMMENT ON TABLE t_rbac_role_group_rel IS '角色-权限组关系';
COMMENT ON COLUMN t_rbac_role_group_rel.id IS '主键（雪花ID）';
COMMENT ON COLUMN t_rbac_role_group_rel.role_id IS '角色ID';
COMMENT ON COLUMN t_rbac_role_group_rel.group_id IS '权限组ID';
COMMENT ON COLUMN t_rbac_role_group_rel.create_time IS '创建时间';
COMMENT ON COLUMN t_rbac_role_group_rel.create_by IS '创建人ID';
COMMENT ON COLUMN t_rbac_role_group_rel.update_time IS '更新时间';
COMMENT ON COLUMN t_rbac_role_group_rel.update_by IS '更新人ID';
COMMENT ON COLUMN t_rbac_role_group_rel.delete_time IS '删除时间';
COMMENT ON COLUMN t_rbac_role_group_rel.delete_by IS '删除人ID';

CREATE TABLE IF NOT EXISTS t_rbac_group_permission_rel (
    id            BIGINT    NOT NULL,
    group_id      BIGINT    NOT NULL,
    permission_id BIGINT    NOT NULL,
    create_time   TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    create_by     BIGINT    DEFAULT NULL,
    update_time   TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    update_by     BIGINT    DEFAULT NULL,
    delete_time   TIMESTAMP DEFAULT NULL,
    delete_by     BIGINT    DEFAULT NULL,
    PRIMARY KEY (id),
    CONSTRAINT uk_t_rbac_group_permission_rel_group_permission UNIQUE (group_id, permission_id)
);

CREATE INDEX IF NOT EXISTS idx_t_rbac_group_permission_rel_permission_id ON t_rbac_group_permission_rel (permission_id);

COMMENT ON TABLE t_rbac_group_permission_rel IS '权限组-权限关系';
COMMENT ON COLUMN t_rbac_group_permission_rel.id IS '主键（雪花ID）';
COMMENT ON COLUMN t_rbac_group_permission_rel.group_id IS '权限组ID';
COMMENT ON COLUMN t_rbac_group_permission_rel.permission_id IS '权限ID';
COMMENT ON COLUMN t_rbac_group_permission_rel.create_time IS '创建时间';
COMMENT ON COLUMN t_rbac_group_permission_rel.create_by IS '创建人ID';
COMMENT ON COLUMN t_rbac_group_permission_rel.update_time IS '更新时间';
COMMENT ON COLUMN t_rbac_group_permission_rel.update_by IS '更新人ID';
COMMENT ON COLUMN t_rbac_group_permission_rel.delete_time IS '删除时间';
COMMENT ON COLUMN t_rbac_group_permission_rel.delete_by IS '删除人ID';

-- ============================================================
-- HTTP 访问记录（httprecord 采集后由业务消费者落库）
-- ============================================================

CREATE TABLE IF NOT EXISTS t_http_record (
    id                      BIGINT        NOT NULL,
    service_type            VARCHAR(64)   NOT NULL DEFAULT '',
    service_version         VARCHAR(32)   NOT NULL DEFAULT '',
    trace_parent            VARCHAR(55)   NOT NULL DEFAULT '',
    request_at              TIMESTAMP(3)  NOT NULL,
    duration_ms             BIGINT        NOT NULL DEFAULT 0,
    method                  VARCHAR(16)   NOT NULL,
    path                    VARCHAR(512)  NOT NULL,
    route                   VARCHAR(512)  NOT NULL DEFAULT '',
    query                   TEXT          NOT NULL,
    client_ip               VARCHAR(64)   NOT NULL DEFAULT '',
    user_agent              VARCHAR(1024) NOT NULL DEFAULT '',
    request_headers         TEXT          NOT NULL,
    request_body            TEXT          NOT NULL,
    request_body_truncated  BOOLEAN       NOT NULL DEFAULT FALSE,
    user_id                 BIGINT        NOT NULL DEFAULT 0,
    client_type             VARCHAR(32)   NOT NULL DEFAULT '',
    status_code             INT           NOT NULL DEFAULT 0,
    response_body           TEXT          NOT NULL,
    response_body_truncated BOOLEAN       NOT NULL DEFAULT FALSE,
    biz_code                VARCHAR(64)   NOT NULL DEFAULT '',
    biz_msg                 VARCHAR(1024) NOT NULL DEFAULT '',
    create_time             TIMESTAMP     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    create_by               BIGINT        DEFAULT NULL,
    update_time             TIMESTAMP     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    update_by               BIGINT        DEFAULT NULL,
    delete_time             TIMESTAMP     DEFAULT NULL,
    delete_by               BIGINT        DEFAULT NULL,
    PRIMARY KEY (id),
    CONSTRAINT uk_t_http_record_traceparent_request UNIQUE (trace_parent, request_at)
);

CREATE INDEX IF NOT EXISTS idx_t_http_record_request_at ON t_http_record (request_at);
CREATE INDEX IF NOT EXISTS idx_t_http_record_user_id ON t_http_record (user_id);
CREATE INDEX IF NOT EXISTS idx_t_http_record_delete_time ON t_http_record (delete_time);

COMMENT ON TABLE t_http_record IS 'HTTP 访问记录';
COMMENT ON COLUMN t_http_record.id IS '主键（雪花ID）';
COMMENT ON COLUMN t_http_record.service_type IS '来源服务名';
COMMENT ON COLUMN t_http_record.service_version IS '来源服务版本';
COMMENT ON COLUMN t_http_record.trace_parent IS '本跳 W3C traceparent';
COMMENT ON COLUMN t_http_record.request_at IS '请求到达时间';
COMMENT ON COLUMN t_http_record.duration_ms IS '处理耗时（毫秒）';
COMMENT ON COLUMN t_http_record.method IS 'HTTP 方法';
COMMENT ON COLUMN t_http_record.path IS '请求路径';
COMMENT ON COLUMN t_http_record.route IS 'gin 路由模板';
COMMENT ON COLUMN t_http_record.query IS '脱敏后的 query string';
COMMENT ON COLUMN t_http_record.client_ip IS '客户端IP';
COMMENT ON COLUMN t_http_record.user_agent IS '客户端 User-Agent';
COMMENT ON COLUMN t_http_record.request_headers IS '脱敏后的请求头 JSON';
COMMENT ON COLUMN t_http_record.request_body IS '脱敏后的请求体';
COMMENT ON COLUMN t_http_record.request_body_truncated IS '请求体是否被截断：false-否，true-是';
COMMENT ON COLUMN t_http_record.user_id IS '当前登录用户ID，未登录为0';
COMMENT ON COLUMN t_http_record.client_type IS '端类型，如 admin';
COMMENT ON COLUMN t_http_record.status_code IS 'HTTP 状态码';
COMMENT ON COLUMN t_http_record.response_body IS '脱敏后的响应体';
COMMENT ON COLUMN t_http_record.response_body_truncated IS '响应体是否被截断：false-否，true-是';
COMMENT ON COLUMN t_http_record.biz_code IS '业务响应码';
COMMENT ON COLUMN t_http_record.biz_msg IS '业务响应消息';
COMMENT ON COLUMN t_http_record.create_time IS '创建时间';
COMMENT ON COLUMN t_http_record.create_by IS '创建人ID';
COMMENT ON COLUMN t_http_record.update_time IS '更新时间';
COMMENT ON COLUMN t_http_record.update_by IS '更新人ID';
COMMENT ON COLUMN t_http_record.delete_time IS '删除时间';
COMMENT ON COLUMN t_http_record.delete_by IS '删除人ID';

-- ============================================================
-- 系统参数（动态配置数据源）
-- ============================================================

CREATE TABLE IF NOT EXISTS t_sys_param (
    id          BIGINT        NOT NULL,
    param_key   VARCHAR(128)  NOT NULL,
    param_value VARCHAR(2048) NOT NULL DEFAULT '',
    value_type  VARCHAR(20)   NOT NULL,
    name        VARCHAR(64)   NOT NULL,
    remark      VARCHAR(255)  NOT NULL DEFAULT '',
    status      VARCHAR(20)   NOT NULL DEFAULT 'ENABLED',
    create_time TIMESTAMP     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    create_by   BIGINT        DEFAULT NULL,
    update_time TIMESTAMP     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    update_by   BIGINT        DEFAULT NULL,
    delete_time TIMESTAMP     DEFAULT NULL,
    delete_by   BIGINT        DEFAULT NULL,
    PRIMARY KEY (id),
    CONSTRAINT uk_t_sys_param_param_key UNIQUE (param_key)
);

CREATE INDEX IF NOT EXISTS idx_t_sys_param_status ON t_sys_param (status);

COMMENT ON TABLE t_sys_param IS '系统参数';
COMMENT ON COLUMN t_sys_param.id IS '主键（雪花ID）';
COMMENT ON COLUMN t_sys_param.param_key IS '参数键，全局唯一，小写点分命名（如 order.auto_cancel_days）';
COMMENT ON COLUMN t_sys_param.param_value IS '参数值，统一字符串形态存储，按 value_type 解释';
COMMENT ON COLUMN t_sys_param.value_type IS '值类型：BOOL-布尔，INT-整数，STRING-字符串，JSON-JSON';
COMMENT ON COLUMN t_sys_param.name IS '参数名称（展示用）';
COMMENT ON COLUMN t_sys_param.remark IS '备注';
COMMENT ON COLUMN t_sys_param.status IS '状态：ENABLED-启用，DISABLED-禁用（禁用=读取方回退默认值）';
COMMENT ON COLUMN t_sys_param.create_time IS '创建时间';
COMMENT ON COLUMN t_sys_param.create_by IS '创建人ID';
COMMENT ON COLUMN t_sys_param.update_time IS '更新时间';
COMMENT ON COLUMN t_sys_param.update_by IS '更新人ID';
COMMENT ON COLUMN t_sys_param.delete_time IS '删除时间';
COMMENT ON COLUMN t_sys_param.delete_by IS '删除人ID';

-- ============================================================
-- 操作审计日志：按需记录需要关注的业务操作
-- ============================================================

CREATE TABLE IF NOT EXISTS t_audit_log (
    id            BIGINT       NOT NULL,
    audit_id      BIGINT       NOT NULL,
    phase         VARCHAR(16)  NOT NULL,
    event_code    VARCHAR(64)  NOT NULL,
    operator_id   BIGINT       NOT NULL DEFAULT 0,
    operator_name VARCHAR(64)  NOT NULL DEFAULT '',
    client_ip     VARCHAR(64)  NOT NULL DEFAULT '',
    target_type   VARCHAR(32)  NOT NULL DEFAULT '',
    target_id     BIGINT       NOT NULL DEFAULT 0,
    result        VARCHAR(16)  NOT NULL DEFAULT 'PENDING',
    biz_code      VARCHAR(64)  NOT NULL DEFAULT '',
    create_time   TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    create_by     BIGINT       DEFAULT NULL,
    update_time   TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    update_by     BIGINT       DEFAULT NULL,
    delete_time   TIMESTAMP    DEFAULT NULL,
    delete_by     BIGINT       DEFAULT NULL,
    PRIMARY KEY (id)
);

CREATE INDEX IF NOT EXISTS idx_t_audit_log_audit_id ON t_audit_log (audit_id);
CREATE INDEX IF NOT EXISTS idx_t_audit_log_create_time ON t_audit_log (create_time);
CREATE INDEX IF NOT EXISTS idx_t_audit_log_phase_create_time ON t_audit_log (phase, create_time);
CREATE INDEX IF NOT EXISTS idx_t_audit_log_operator_id_create_time ON t_audit_log (operator_id, create_time);
CREATE INDEX IF NOT EXISTS idx_t_audit_log_event_code_create_time ON t_audit_log (event_code, create_time);
CREATE INDEX IF NOT EXISTS idx_t_audit_log_result_create_time ON t_audit_log (result, create_time);

COMMENT ON TABLE t_audit_log IS '操作审计日志';
COMMENT ON COLUMN t_audit_log.id IS '主键（雪花ID）';
COMMENT ON COLUMN t_audit_log.audit_id IS '审计事件实例ID，开始/结束共用';
COMMENT ON COLUMN t_audit_log.phase IS '阶段：STARTED-开始，FINISHED-结束';
COMMENT ON COLUMN t_audit_log.event_code IS '事件编码，按业务事件扩展，如 auth.login';
COMMENT ON COLUMN t_audit_log.operator_id IS '业务操作者用户ID；登录失败或用户不存在为0';
COMMENT ON COLUMN t_audit_log.operator_name IS '业务操作者用户名快照；登录事件为尝试的用户名';
COMMENT ON COLUMN t_audit_log.client_ip IS '来源IP';
COMMENT ON COLUMN t_audit_log.target_type IS '目标类型，如 user；无目标为空';
COMMENT ON COLUMN t_audit_log.target_id IS '目标ID；无目标为0';
COMMENT ON COLUMN t_audit_log.result IS '结果：PENDING-开始未出结果，SUCCESS-成功，FAILURE-失败';
COMMENT ON COLUMN t_audit_log.biz_code IS '业务码；成功为 OK';
COMMENT ON COLUMN t_audit_log.create_time IS '事件发生时间（UTC）';
COMMENT ON COLUMN t_audit_log.create_by IS '审计写入人ID；系统写入，可为空';
COMMENT ON COLUMN t_audit_log.update_time IS '更新时间；审计只追加，保留统一字段';
COMMENT ON COLUMN t_audit_log.update_by IS '更新人ID';
COMMENT ON COLUMN t_audit_log.delete_time IS '删除时间；审计不软删，保留统一字段';
COMMENT ON COLUMN t_audit_log.delete_by IS '删除人ID';

-- ============================================================
-- 业务幂等表（表名与 yomi-boot 的 idempotency 组件保持一致）
-- 权威 DDL 由本仓维护，框架升级改表结构时同步修改这里。
-- 索引与约束命名沿用 yomi-boot 的 sql/pgsql/t_sys_biz.sql。
-- ============================================================

CREATE TABLE IF NOT EXISTS t_sys_biz (
    id             BIGINT       NOT NULL,
    biz_id         VARCHAR(100) NOT NULL,
    biz_sub_id     VARCHAR(100) NOT NULL,
    operation_type VARCHAR(32)  NOT NULL,
    status         SMALLINT     NOT NULL DEFAULT 1,
    response_data  JSONB        DEFAULT NULL,
    create_time    TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    create_by      BIGINT       DEFAULT NULL,
    update_time    TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    update_by      BIGINT       DEFAULT NULL,
    delete_time    TIMESTAMP    DEFAULT NULL,
    delete_by      BIGINT       DEFAULT NULL,
    PRIMARY KEY (id),
    CONSTRAINT uk_biz UNIQUE (biz_id, biz_sub_id)
);

CREATE INDEX IF NOT EXISTS idx_operation ON t_sys_biz (operation_type);
CREATE INDEX IF NOT EXISTS idx_create_time ON t_sys_biz (create_time);
CREATE INDEX IF NOT EXISTS idx_status_update_time ON t_sys_biz (status, update_time);

COMMENT ON TABLE t_sys_biz IS '业务幂等记录表';
COMMENT ON COLUMN t_sys_biz.id IS '主键（雪花ID）';
COMMENT ON COLUMN t_sys_biz.biz_id IS '业务全局ID（一次 Saga/业务意图）';
COMMENT ON COLUMN t_sys_biz.biz_sub_id IS '业务子ID（单次领域写调用）';
COMMENT ON COLUMN t_sys_biz.operation_type IS '操作类型（如 UPDATE_USER_STATUS）';
COMMENT ON COLUMN t_sys_biz.status IS '1-处理中 2-已完成';
COMMENT ON COLUMN t_sys_biz.response_data IS '成功时的返回结果缓存（用于重试直接返回）';
COMMENT ON COLUMN t_sys_biz.create_time IS '创建时间（UTC）';
COMMENT ON COLUMN t_sys_biz.create_by IS '创建人ID';
COMMENT ON COLUMN t_sys_biz.update_time IS '更新时间（UTC）';
COMMENT ON COLUMN t_sys_biz.update_by IS '更新人ID';
COMMENT ON COLUMN t_sys_biz.delete_time IS '删除时间';
COMMENT ON COLUMN t_sys_biz.delete_by IS '删除人ID';

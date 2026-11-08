-- yomi-admin-backend 表结构初始化：只含 DDL，可重复执行
-- 种子数据按领域拆在 seed_*.sql，按文件名顺序在 init.sql 之后执行：
--   seed_10_user.sql  用户域（内置管理员）
--   seed_20_rbac.sql  RBAC 域（角色 / 权限目录 / 权限组及关系）
-- 不在此指定库名：先自行建库（utf8mb4 / utf8mb4_unicode_ci，库名与 dev.yml 的 database 一致），
-- 执行时选中它，例如：mysql -uroot -p -D yomi_admin < script/sql/init.sql
SET NAMES utf8mb4;

-- ============================================================
-- 用户
-- ============================================================

CREATE TABLE IF NOT EXISTS `t_user` (
    `id`            BIGINT       NOT NULL COMMENT '主键（雪花ID）',
    `username`      VARCHAR(64)  NOT NULL COMMENT '登录名',
    `password_hash` VARCHAR(255) NOT NULL COMMENT 'bcrypt 哈希',
    `nickname`      VARCHAR(64)  NOT NULL DEFAULT '' COMMENT '昵称',
    `email`         VARCHAR(128) NOT NULL DEFAULT '' COMMENT '邮箱',
    `phone`         VARCHAR(32)  NOT NULL DEFAULT '' COMMENT '手机号',
    `status`        VARCHAR(20)  NOT NULL DEFAULT 'ENABLED' COMMENT '状态：ENABLED-启用，DISABLED-禁用',
    `create_time`   DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `create_by`     BIGINT       DEFAULT NULL COMMENT '创建人ID',
    `update_time`   DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    `update_by`     BIGINT       DEFAULT NULL COMMENT '更新人ID',
    `delete_time`   DATETIME     DEFAULT NULL COMMENT '删除时间',
    `delete_by`     BIGINT       DEFAULT NULL COMMENT '删除人ID',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_username` (`username`),
    KEY `idx_delete_time` (`delete_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户';

-- ============================================================
-- RBAC
-- ============================================================

CREATE TABLE IF NOT EXISTS `t_rbac_role` (
    `id`          BIGINT       NOT NULL COMMENT '主键（雪花ID）',
    `code`        VARCHAR(64)  NOT NULL COMMENT '角色编码，全局唯一；super_admin 为内置超管',
    `name`        VARCHAR(64)  NOT NULL COMMENT '角色名称',
    `builtin`     TINYINT      NOT NULL DEFAULT 0 COMMENT '是否内置：0-否，1-是',
    `status`      VARCHAR(20)  NOT NULL DEFAULT 'ENABLED' COMMENT '状态：ENABLED-启用，DISABLED-禁用',
    `remark`      VARCHAR(255) NOT NULL DEFAULT '' COMMENT '备注',
    `create_time` DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `create_by`   BIGINT       DEFAULT NULL COMMENT '创建人ID',
    `update_time` DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    `update_by`   BIGINT       DEFAULT NULL COMMENT '更新人ID',
    `delete_time` DATETIME     DEFAULT NULL COMMENT '删除时间',
    `delete_by`   BIGINT       DEFAULT NULL COMMENT '删除人ID',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_code` (`code`),
    KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='角色';

CREATE TABLE IF NOT EXISTS `t_rbac_permission` (
    `id`              BIGINT       NOT NULL COMMENT '主键（雪花ID）',
    `code`            VARCHAR(96)  NOT NULL COMMENT '权限编码，全局唯一',
    `name`            VARCHAR(64)  NOT NULL COMMENT '权限名称',
    `permission_type` VARCHAR(20)  NOT NULL COMMENT '权限类型：MENU-菜单，API-接口',
    `ui_path`         VARCHAR(191) DEFAULT NULL COMMENT '菜单路径，permission_type=MENU 时必填',
    `http_method`     VARCHAR(16)  DEFAULT NULL COMMENT 'HTTP 方法，permission_type=API 时必填',
    `http_path`       VARCHAR(191) DEFAULT NULL COMMENT 'gin 路由模板（以 / 开头，通常含 /api），permission_type=API 时必填',
    `status`          VARCHAR(20)  NOT NULL DEFAULT 'ENABLED' COMMENT '状态：ENABLED-启用，DISABLED-禁用',
    `remark`          VARCHAR(255) NOT NULL DEFAULT '' COMMENT '备注',
    `create_time`     DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `create_by`       BIGINT       DEFAULT NULL COMMENT '创建人ID',
    `update_time`     DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    `update_by`       BIGINT       DEFAULT NULL COMMENT '更新人ID',
    `delete_time`     DATETIME     DEFAULT NULL COMMENT '删除时间',
    `delete_by`       BIGINT       DEFAULT NULL COMMENT '删除人ID',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_code` (`code`),
    UNIQUE KEY `uk_http_route` (`http_method`, `http_path`),
    UNIQUE KEY `uk_ui_path` (`ui_path`),
    KEY `idx_permission_type_status` (`permission_type`, `status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='权限目录';

CREATE TABLE IF NOT EXISTS `t_rbac_permission_group` (
    `id`          BIGINT       NOT NULL COMMENT '主键（雪花ID）',
    `code`        VARCHAR(64)  NOT NULL COMMENT '权限组编码，全局唯一',
    `name`        VARCHAR(64)  NOT NULL COMMENT '权限组名称',
    `status`      VARCHAR(20)  NOT NULL DEFAULT 'ENABLED' COMMENT '状态：ENABLED-启用，DISABLED-禁用',
    `remark`      VARCHAR(255) NOT NULL DEFAULT '' COMMENT '备注',
    `create_time` DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `create_by`   BIGINT       DEFAULT NULL COMMENT '创建人ID',
    `update_time` DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    `update_by`   BIGINT       DEFAULT NULL COMMENT '更新人ID',
    `delete_time` DATETIME     DEFAULT NULL COMMENT '删除时间',
    `delete_by`   BIGINT       DEFAULT NULL COMMENT '删除人ID',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_code` (`code`),
    KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='权限组';

CREATE TABLE IF NOT EXISTS `t_rbac_user_role_rel` (
    `id`          BIGINT   NOT NULL COMMENT '主键（雪花ID）',
    `user_id`     BIGINT   NOT NULL COMMENT '用户ID（t_user.id）',
    `role_id`     BIGINT   NOT NULL COMMENT '角色ID',
    `create_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `create_by`   BIGINT   DEFAULT NULL COMMENT '创建人ID',
    `update_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    `update_by`   BIGINT   DEFAULT NULL COMMENT '更新人ID',
    `delete_time` DATETIME DEFAULT NULL COMMENT '删除时间',
    `delete_by`   BIGINT   DEFAULT NULL COMMENT '删除人ID',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_user_role` (`user_id`, `role_id`),
    KEY `idx_role_id` (`role_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户-角色关系';

CREATE TABLE IF NOT EXISTS `t_rbac_role_group_rel` (
    `id`          BIGINT   NOT NULL COMMENT '主键（雪花ID）',
    `role_id`     BIGINT   NOT NULL COMMENT '角色ID',
    `group_id`    BIGINT   NOT NULL COMMENT '权限组ID',
    `create_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `create_by`   BIGINT   DEFAULT NULL COMMENT '创建人ID',
    `update_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    `update_by`   BIGINT   DEFAULT NULL COMMENT '更新人ID',
    `delete_time` DATETIME DEFAULT NULL COMMENT '删除时间',
    `delete_by`   BIGINT   DEFAULT NULL COMMENT '删除人ID',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_role_group` (`role_id`, `group_id`),
    KEY `idx_group_id` (`group_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='角色-权限组关系';

CREATE TABLE IF NOT EXISTS `t_rbac_group_permission_rel` (
    `id`            BIGINT   NOT NULL COMMENT '主键（雪花ID）',
    `group_id`      BIGINT   NOT NULL COMMENT '权限组ID',
    `permission_id` BIGINT   NOT NULL COMMENT '权限ID',
    `create_time`   DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `create_by`     BIGINT   DEFAULT NULL COMMENT '创建人ID',
    `update_time`   DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    `update_by`     BIGINT   DEFAULT NULL COMMENT '更新人ID',
    `delete_time`   DATETIME DEFAULT NULL COMMENT '删除时间',
    `delete_by`     BIGINT   DEFAULT NULL COMMENT '删除人ID',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_group_permission` (`group_id`, `permission_id`),
    KEY `idx_permission_id` (`permission_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='权限组-权限关系';

-- ============================================================
-- HTTP 访问记录（httprecord 采集后由业务消费者落库）
-- ============================================================

CREATE TABLE IF NOT EXISTS `t_http_record` (
    `id`                      BIGINT       NOT NULL COMMENT '主键（雪花ID）',
    `service_type`            VARCHAR(64)  NOT NULL DEFAULT '' COMMENT '来源服务名',
    `service_version`         VARCHAR(32)  NOT NULL DEFAULT '' COMMENT '来源服务版本',
    `trace_parent`            VARCHAR(55)  NOT NULL DEFAULT '' COMMENT '本跳 W3C traceparent',
    `request_at`              DATETIME(3)  NOT NULL COMMENT '请求到达时间',
    `duration_ms`             BIGINT       NOT NULL DEFAULT 0 COMMENT '处理耗时（毫秒）',
    `method`                  VARCHAR(16)  NOT NULL COMMENT 'HTTP 方法',
    `path`                    VARCHAR(512) NOT NULL COMMENT '请求路径',
    `route`                   VARCHAR(512) NOT NULL DEFAULT '' COMMENT 'gin 路由模板',
    `query`                   TEXT         NOT NULL COMMENT '脱敏后的 query string',
    `client_ip`               VARCHAR(64)  NOT NULL DEFAULT '' COMMENT '客户端IP',
    `user_agent`              VARCHAR(1024) NOT NULL DEFAULT '' COMMENT '客户端 User-Agent',
    `request_headers`         TEXT         NOT NULL COMMENT '脱敏后的请求头 JSON',
    `request_body`            MEDIUMTEXT   NOT NULL COMMENT '脱敏后的请求体',
    `request_body_truncated`  TINYINT      NOT NULL DEFAULT 0 COMMENT '请求体是否被截断：0-否，1-是',
    `user_id`                 BIGINT       NOT NULL DEFAULT 0 COMMENT '当前登录用户ID，未登录为0',
    `client_type`             VARCHAR(32)  NOT NULL DEFAULT '' COMMENT '端类型，如 admin',
    `status_code`             INT          NOT NULL DEFAULT 0 COMMENT 'HTTP 状态码',
    `response_body`           MEDIUMTEXT   NOT NULL COMMENT '脱敏后的响应体',
    `response_body_truncated` TINYINT      NOT NULL DEFAULT 0 COMMENT '响应体是否被截断：0-否，1-是',
    `biz_code`                VARCHAR(64)  NOT NULL DEFAULT '' COMMENT '业务响应码',
    `biz_msg`                 VARCHAR(1024) NOT NULL DEFAULT '' COMMENT '业务响应消息',
    `create_time`             DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `create_by`               BIGINT       DEFAULT NULL COMMENT '创建人ID',
    `update_time`             DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    `update_by`               BIGINT       DEFAULT NULL COMMENT '更新人ID',
    `delete_time`             DATETIME     DEFAULT NULL COMMENT '删除时间',
    `delete_by`               BIGINT       DEFAULT NULL COMMENT '删除人ID',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_traceparent_request` (`trace_parent`, `request_at`),
    KEY `idx_request_at` (`request_at`),
    KEY `idx_user_id` (`user_id`),
    KEY `idx_delete_time` (`delete_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='HTTP 访问记录';

-- ============================================================
-- 系统参数（动态配置数据源）
-- ============================================================

CREATE TABLE IF NOT EXISTS `t_sys_param` (
    `id`          BIGINT        NOT NULL COMMENT '主键（雪花ID）',
    `param_key`   VARCHAR(128)  NOT NULL COMMENT '参数键，全局唯一，小写点分命名（如 order.auto_cancel_days）',
    `param_value` VARCHAR(2048) NOT NULL DEFAULT '' COMMENT '参数值，统一字符串形态存储，按 value_type 解释',
    `value_type`  VARCHAR(20)   NOT NULL COMMENT '值类型：BOOL-布尔，INT-整数，STRING-字符串，JSON-JSON',
    `name`        VARCHAR(64)   NOT NULL COMMENT '参数名称（展示用）',
    `remark`      VARCHAR(255)  NOT NULL DEFAULT '' COMMENT '备注',
    `status`      VARCHAR(20)   NOT NULL DEFAULT 'ENABLED' COMMENT '状态：ENABLED-启用，DISABLED-禁用（禁用=读取方回退默认值）',
    `create_time` DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `create_by`   BIGINT        DEFAULT NULL COMMENT '创建人ID',
    `update_time` DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    `update_by`   BIGINT        DEFAULT NULL COMMENT '更新人ID',
    `delete_time` DATETIME      DEFAULT NULL COMMENT '删除时间',
    `delete_by`   BIGINT        DEFAULT NULL COMMENT '删除人ID',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_param_key` (`param_key`),
    KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='系统参数';

-- ============================================================
-- 操作审计日志：按需记录需要关注的业务操作
-- ============================================================

CREATE TABLE IF NOT EXISTS `t_audit_log` (
    `id`            BIGINT       NOT NULL COMMENT '主键（雪花ID）',
    `audit_id`      BIGINT       NOT NULL COMMENT '审计事件实例ID，开始/结束共用',
    `phase`         VARCHAR(16)  NOT NULL COMMENT '阶段：STARTED-开始，FINISHED-结束',
    `event_code`    VARCHAR(64)  NOT NULL COMMENT '事件编码，按业务事件扩展，如 auth.login',
    `operator_id`   BIGINT       NOT NULL DEFAULT 0 COMMENT '业务操作者用户ID；登录失败或用户不存在为0',
    `operator_name` VARCHAR(64)  NOT NULL DEFAULT '' COMMENT '业务操作者用户名快照；登录事件为尝试的用户名',
    `client_ip`     VARCHAR(64)  NOT NULL DEFAULT '' COMMENT '来源IP',
    `target_type`   VARCHAR(32)  NOT NULL DEFAULT '' COMMENT '目标类型，如 user；无目标为空',
    `target_id`     BIGINT       NOT NULL DEFAULT 0 COMMENT '目标ID；无目标为0',
    `result`        VARCHAR(16)  NOT NULL DEFAULT 'PENDING' COMMENT '结果：PENDING-开始未出结果，SUCCESS-成功，FAILURE-失败',
    `biz_code`      VARCHAR(64)  NOT NULL DEFAULT '' COMMENT '业务码；成功为 OK',
    `create_time`   DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '事件发生时间（UTC）',
    `create_by`     BIGINT       DEFAULT NULL COMMENT '审计写入人ID；系统写入，可为空',
    `update_time`   DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间；审计只追加，保留统一字段',
    `update_by`     BIGINT       DEFAULT NULL COMMENT '更新人ID',
    `delete_time`   DATETIME     DEFAULT NULL COMMENT '删除时间；审计不软删，保留统一字段',
    `delete_by`     BIGINT       DEFAULT NULL COMMENT '删除人ID',
    PRIMARY KEY (`id`),
    KEY `idx_audit_id` (`audit_id`),
    KEY `idx_create_time` (`create_time`),
    KEY `idx_phase_create_time` (`phase`, `create_time`),
    KEY `idx_operator_id_create_time` (`operator_id`, `create_time`),
    KEY `idx_event_code_create_time` (`event_code`, `create_time`),
    KEY `idx_result_create_time` (`result`, `create_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='操作审计日志';

-- ============================================================
-- 业务幂等表（表名与 yomi-boot 的 idempotency 组件保持一致）
-- 权威 DDL 由本仓维护，框架升级改表结构时同步修改这里。
-- ============================================================

CREATE TABLE IF NOT EXISTS `t_sys_biz` (
    `id`                  BIGINT       NOT NULL COMMENT '主键（雪花ID）',
    `biz_id`              VARCHAR(100) NOT NULL COMMENT '业务全局ID（一次 Saga/业务意图）',
    `biz_sub_id`          VARCHAR(100) NOT NULL COMMENT '业务子ID（单次领域写调用）',
    `operation_type`      VARCHAR(32)  NOT NULL COMMENT '操作类型（如 UPDATE_USER_STATUS）',
    `status`              TINYINT      NOT NULL DEFAULT 1 COMMENT '1-处理中 2-已完成',
    `response_data`       JSON         DEFAULT NULL COMMENT '成功时的返回结果缓存（用于重试直接返回）',
    `create_time`         DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间（UTC）',
    `create_by`           BIGINT       DEFAULT NULL COMMENT '创建人ID',
    `update_time`         DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间（UTC）',
    `update_by`           BIGINT       DEFAULT NULL COMMENT '更新人ID',
    `delete_time`         DATETIME     DEFAULT NULL COMMENT '删除时间',
    `delete_by`           BIGINT       DEFAULT NULL COMMENT '删除人ID',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_biz` (`biz_id`, `biz_sub_id`),
    KEY `idx_operation` (`operation_type`),
    KEY `idx_create_time` (`create_time`),
    KEY `idx_status_update_time` (`status`, `update_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='业务幂等记录表';

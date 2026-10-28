-- RBAC 角色与权限表，以及内置超管与权限目录种子
SET NAMES utf8mb4;
USE yomi_admin;

-- 后端统一使用 UTC，种子时间也按 UTC 写入
SET time_zone = '+00:00';

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
    `http_path`       VARCHAR(191) DEFAULT NULL COMMENT 'gin 路由模板（含 /api），permission_type=API 时必填',
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

CREATE TABLE IF NOT EXISTS `t_rbac_role_permission_rel` (
    `id`            BIGINT   NOT NULL COMMENT '主键（雪花ID）',
    `role_id`       BIGINT   NOT NULL COMMENT '角色ID',
    `permission_id` BIGINT   NOT NULL COMMENT '权限ID',
    `create_time`   DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `create_by`     BIGINT   DEFAULT NULL COMMENT '创建人ID',
    `update_time`   DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    `update_by`     BIGINT   DEFAULT NULL COMMENT '更新人ID',
    `delete_time`   DATETIME DEFAULT NULL COMMENT '删除时间',
    `delete_by`     BIGINT   DEFAULT NULL COMMENT '删除人ID',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_role_permission` (`role_id`, `permission_id`),
    KEY `idx_permission_id` (`permission_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='角色-权限关系';

-- 内置超管角色
INSERT IGNORE INTO `t_rbac_role` (`id`, `code`, `name`, `builtin`, `status`, `remark`)
VALUES (100000000000000101, 'super_admin', '超级管理员', 1, 'ENABLED', '内置角色，拥有全部权限');

-- 菜单权限
INSERT IGNORE INTO `t_rbac_permission` (`id`, `code`, `name`, `permission_type`, `ui_path`, `http_method`, `http_path`, `status`)
VALUES (100000000000001001, 'system:user', '系统用户', 'MENU', '/system/user', NULL, NULL, 'ENABLED');
INSERT IGNORE INTO `t_rbac_permission` (`id`, `code`, `name`, `permission_type`, `ui_path`, `http_method`, `http_path`, `status`)
VALUES (100000000000001002, 'system:role', '角色管理', 'MENU', '/system/role', NULL, NULL, 'ENABLED');
INSERT IGNORE INTO `t_rbac_permission` (`id`, `code`, `name`, `permission_type`, `ui_path`, `http_method`, `http_path`, `status`)
VALUES (100000000000001003, 'system:permission', '权限管理', 'MENU', '/system/permission', NULL, NULL, 'ENABLED');

-- 用户接口权限
INSERT IGNORE INTO `t_rbac_permission` (`id`, `code`, `name`, `permission_type`, `ui_path`, `http_method`, `http_path`, `status`)
VALUES (100000000000001011, 'system:user:list', '用户列表', 'API', NULL, 'GET', '/api/system/users', 'ENABLED');
INSERT IGNORE INTO `t_rbac_permission` (`id`, `code`, `name`, `permission_type`, `ui_path`, `http_method`, `http_path`, `status`)
VALUES (100000000000001012, 'system:user:create', '创建用户', 'API', NULL, 'POST', '/api/system/users', 'ENABLED');
INSERT IGNORE INTO `t_rbac_permission` (`id`, `code`, `name`, `permission_type`, `ui_path`, `http_method`, `http_path`, `status`)
VALUES (100000000000001013, 'system:user:update', '更新用户', 'API', NULL, 'PUT', '/api/system/users/:id', 'ENABLED');
INSERT IGNORE INTO `t_rbac_permission` (`id`, `code`, `name`, `permission_type`, `ui_path`, `http_method`, `http_path`, `status`)
VALUES (100000000000001014, 'system:user:delete', '删除用户', 'API', NULL, 'DELETE', '/api/system/users/:id', 'ENABLED');

-- 角色接口权限
INSERT IGNORE INTO `t_rbac_permission` (`id`, `code`, `name`, `permission_type`, `ui_path`, `http_method`, `http_path`, `status`)
VALUES (100000000000001021, 'system:role:list', '角色列表', 'API', NULL, 'GET', '/api/system/role', 'ENABLED');
INSERT IGNORE INTO `t_rbac_permission` (`id`, `code`, `name`, `permission_type`, `ui_path`, `http_method`, `http_path`, `status`)
VALUES (100000000000001022, 'system:role:detail', '角色详情', 'API', NULL, 'GET', '/api/system/role/:id', 'ENABLED');
INSERT IGNORE INTO `t_rbac_permission` (`id`, `code`, `name`, `permission_type`, `ui_path`, `http_method`, `http_path`, `status`)
VALUES (100000000000001023, 'system:role:create', '创建角色', 'API', NULL, 'POST', '/api/system/role', 'ENABLED');
INSERT IGNORE INTO `t_rbac_permission` (`id`, `code`, `name`, `permission_type`, `ui_path`, `http_method`, `http_path`, `status`)
VALUES (100000000000001024, 'system:role:update', '更新角色', 'API', NULL, 'PUT', '/api/system/role/:id', 'ENABLED');
INSERT IGNORE INTO `t_rbac_permission` (`id`, `code`, `name`, `permission_type`, `ui_path`, `http_method`, `http_path`, `status`)
VALUES (100000000000001025, 'system:role:bind-permission', '角色绑定权限', 'API', NULL, 'PUT', '/api/system/role/:id/permission', 'ENABLED');

-- 权限目录接口权限
INSERT IGNORE INTO `t_rbac_permission` (`id`, `code`, `name`, `permission_type`, `ui_path`, `http_method`, `http_path`, `status`)
VALUES (100000000000001031, 'system:permission:list', '权限列表', 'API', NULL, 'GET', '/api/system/permission', 'ENABLED');
INSERT IGNORE INTO `t_rbac_permission` (`id`, `code`, `name`, `permission_type`, `ui_path`, `http_method`, `http_path`, `status`)
VALUES (100000000000001032, 'system:permission:status', '权限启停', 'API', NULL, 'PUT', '/api/system/permission/:id/status', 'ENABLED');

-- 用户角色接口权限
INSERT IGNORE INTO `t_rbac_permission` (`id`, `code`, `name`, `permission_type`, `ui_path`, `http_method`, `http_path`, `status`)
VALUES (100000000000001041, 'system:user:role:list', '用户角色查询', 'API', NULL, 'GET', '/api/system/user/:id/role', 'ENABLED');
INSERT IGNORE INTO `t_rbac_permission` (`id`, `code`, `name`, `permission_type`, `ui_path`, `http_method`, `http_path`, `status`)
VALUES (100000000000001042, 'system:user:role:bind', '用户角色绑定', 'API', NULL, 'PUT', '/api/system/user/:id/role', 'ENABLED');

-- 内置管理员绑定超管角色
INSERT IGNORE INTO `t_rbac_user_role_rel` (`id`, `user_id`, `role_id`, `create_by`, `update_by`)
SELECT 100000000000002001, u.id, r.id, u.id, u.id
FROM `t_user` u
JOIN `t_rbac_role` r ON r.code = 'super_admin'
WHERE u.username = 'admin' AND u.delete_time IS NULL;

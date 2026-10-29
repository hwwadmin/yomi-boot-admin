-- RBAC 角色、权限组、权限目录，以及内置超管与种子数据
SET NAMES utf8mb4;
USE yomi_admin;

-- 后端统一使用 UTC，种子时间也按 UTC 写入
SET time_zone = '+00:00';

-- 角色直连权限的旧关系表已废弃，权限一律经权限组授予
DROP TABLE IF EXISTS `t_rbac_role_permission_rel`;

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

-- 内置超管角色
INSERT IGNORE INTO `t_rbac_role` (`id`, `code`, `name`, `builtin`, `status`, `remark`)
VALUES (100000000000000101, 'super_admin', '超级管理员', 1, 'ENABLED', '内置角色，拥有全部权限');

-- 接口权限目录与权限组：种子只负责初始化，管理端可继续维护
-- 菜单权限，名称与前端导航（权限管理下的 角色 / 权限组 / 权限）保持一致
INSERT IGNORE INTO `t_rbac_permission` (`id`, `code`, `name`, `permission_type`, `ui_path`, `http_method`, `http_path`, `status`)
VALUES (100000000000001001, 'system:user', '系统用户', 'MENU', '/system/user', NULL, NULL, 'ENABLED');
INSERT IGNORE INTO `t_rbac_permission` (`id`, `code`, `name`, `permission_type`, `ui_path`, `http_method`, `http_path`, `status`)
VALUES (100000000000001002, 'system:role', '角色', 'MENU', '/system/role', NULL, NULL, 'ENABLED');
INSERT IGNORE INTO `t_rbac_permission` (`id`, `code`, `name`, `permission_type`, `ui_path`, `http_method`, `http_path`, `status`)
VALUES (100000000000001003, 'system:permission', '权限', 'MENU', '/system/permission', NULL, NULL, 'ENABLED');
INSERT IGNORE INTO `t_rbac_permission` (`id`, `code`, `name`, `permission_type`, `ui_path`, `http_method`, `http_path`, `status`)
VALUES (100000000000001004, 'system:http-record', '访问记录', 'MENU', '/system/http-record', NULL, NULL, 'ENABLED');
INSERT IGNORE INTO `t_rbac_permission` (`id`, `code`, `name`, `permission_type`, `ui_path`, `http_method`, `http_path`, `status`)
VALUES (100000000000001005, 'system:permission-group', '权限组', 'MENU', '/system/permission-group', NULL, NULL, 'ENABLED');

-- INSERT IGNORE 不会更新已存在的行，这里显式修正既有库中的旧菜单名称
UPDATE `t_rbac_permission` SET `name` = '角色' WHERE `code` = 'system:role' AND `permission_type` = 'MENU';
UPDATE `t_rbac_permission` SET `name` = '权限' WHERE `code` = 'system:permission' AND `permission_type` = 'MENU';

-- 用户接口权限
INSERT IGNORE INTO `t_rbac_permission` (`id`, `code`, `name`, `permission_type`, `ui_path`, `http_method`, `http_path`, `status`)
VALUES (100000000000001011, 'system:user:list', '用户列表', 'API', NULL, 'GET', '/api/system/user', 'ENABLED');
INSERT IGNORE INTO `t_rbac_permission` (`id`, `code`, `name`, `permission_type`, `ui_path`, `http_method`, `http_path`, `status`)
VALUES (100000000000001012, 'system:user:create', '创建用户', 'API', NULL, 'POST', '/api/system/user/create', 'ENABLED');
INSERT IGNORE INTO `t_rbac_permission` (`id`, `code`, `name`, `permission_type`, `ui_path`, `http_method`, `http_path`, `status`)
VALUES (100000000000001013, 'system:user:update', '更新用户', 'API', NULL, 'POST', '/api/system/user/update', 'ENABLED');
INSERT IGNORE INTO `t_rbac_permission` (`id`, `code`, `name`, `permission_type`, `ui_path`, `http_method`, `http_path`, `status`)
VALUES (100000000000001014, 'system:user:delete', '删除用户', 'API', NULL, 'POST', '/api/system/user/delete', 'ENABLED');

-- 角色接口权限
INSERT IGNORE INTO `t_rbac_permission` (`id`, `code`, `name`, `permission_type`, `ui_path`, `http_method`, `http_path`, `status`)
VALUES (100000000000001021, 'system:role:list', '角色列表', 'API', NULL, 'GET', '/api/system/role', 'ENABLED');
INSERT IGNORE INTO `t_rbac_permission` (`id`, `code`, `name`, `permission_type`, `ui_path`, `http_method`, `http_path`, `status`)
VALUES (100000000000001022, 'system:role:detail', '角色详情', 'API', NULL, 'GET', '/api/system/role/detail', 'ENABLED');
INSERT IGNORE INTO `t_rbac_permission` (`id`, `code`, `name`, `permission_type`, `ui_path`, `http_method`, `http_path`, `status`)
VALUES (100000000000001023, 'system:role:create', '创建角色', 'API', NULL, 'POST', '/api/system/role/create', 'ENABLED');
INSERT IGNORE INTO `t_rbac_permission` (`id`, `code`, `name`, `permission_type`, `ui_path`, `http_method`, `http_path`, `status`)
VALUES (100000000000001024, 'system:role:update', '更新角色', 'API', NULL, 'POST', '/api/system/role/update', 'ENABLED');
INSERT IGNORE INTO `t_rbac_permission` (`id`, `code`, `name`, `permission_type`, `ui_path`, `http_method`, `http_path`, `status`)
VALUES (100000000000001025, 'system:role:bind-group', '角色绑定权限组', 'API', NULL, 'POST', '/api/system/role/bind-group', 'ENABLED');
INSERT IGNORE INTO `t_rbac_permission` (`id`, `code`, `name`, `permission_type`, `ui_path`, `http_method`, `http_path`, `status`)
VALUES (100000000000001026, 'system:role:delete', '删除角色', 'API', NULL, 'POST', '/api/system/role/delete', 'ENABLED');
INSERT IGNORE INTO `t_rbac_permission` (`id`, `code`, `name`, `permission_type`, `ui_path`, `http_method`, `http_path`, `status`)
VALUES (100000000000001027, 'system:role:status', '角色启停', 'API', NULL, 'POST', '/api/system/role/set-status', 'ENABLED');

-- 权限目录接口权限
INSERT IGNORE INTO `t_rbac_permission` (`id`, `code`, `name`, `permission_type`, `ui_path`, `http_method`, `http_path`, `status`)
VALUES (100000000000001031, 'system:permission:list', '权限列表', 'API', NULL, 'GET', '/api/system/permission', 'ENABLED');
INSERT IGNORE INTO `t_rbac_permission` (`id`, `code`, `name`, `permission_type`, `ui_path`, `http_method`, `http_path`, `status`)
VALUES (100000000000001032, 'system:permission:status', '权限启停', 'API', NULL, 'POST', '/api/system/permission/set-status', 'ENABLED');
INSERT IGNORE INTO `t_rbac_permission` (`id`, `code`, `name`, `permission_type`, `ui_path`, `http_method`, `http_path`, `status`)
VALUES (100000000000001033, 'system:permission:create', '创建权限', 'API', NULL, 'POST', '/api/system/permission/create', 'ENABLED');
INSERT IGNORE INTO `t_rbac_permission` (`id`, `code`, `name`, `permission_type`, `ui_path`, `http_method`, `http_path`, `status`)
VALUES (100000000000001034, 'system:permission:update', '更新权限', 'API', NULL, 'POST', '/api/system/permission/update', 'ENABLED');
INSERT IGNORE INTO `t_rbac_permission` (`id`, `code`, `name`, `permission_type`, `ui_path`, `http_method`, `http_path`, `status`)
VALUES (100000000000001035, 'system:permission:delete', '删除权限', 'API', NULL, 'POST', '/api/system/permission/delete', 'ENABLED');

-- 用户角色接口权限
INSERT IGNORE INTO `t_rbac_permission` (`id`, `code`, `name`, `permission_type`, `ui_path`, `http_method`, `http_path`, `status`)
VALUES (100000000000001041, 'system:user:role:list', '用户角色查询', 'API', NULL, 'GET', '/api/system/user/role', 'ENABLED');
INSERT IGNORE INTO `t_rbac_permission` (`id`, `code`, `name`, `permission_type`, `ui_path`, `http_method`, `http_path`, `status`)
VALUES (100000000000001042, 'system:user:role:bind', '用户角色绑定', 'API', NULL, 'POST', '/api/system/user/bind-role', 'ENABLED');

-- 访问记录接口权限
INSERT IGNORE INTO `t_rbac_permission` (`id`, `code`, `name`, `permission_type`, `ui_path`, `http_method`, `http_path`, `status`)
VALUES (100000000000001051, 'system:http-record:list', '访问记录列表', 'API', NULL, 'GET', '/api/system/http-record', 'ENABLED');
INSERT IGNORE INTO `t_rbac_permission` (`id`, `code`, `name`, `permission_type`, `ui_path`, `http_method`, `http_path`, `status`)
VALUES (100000000000001052, 'system:http-record:detail', '访问记录详情', 'API', NULL, 'GET', '/api/system/http-record/detail', 'ENABLED');

-- 权限组接口权限
INSERT IGNORE INTO `t_rbac_permission` (`id`, `code`, `name`, `permission_type`, `ui_path`, `http_method`, `http_path`, `status`)
VALUES (100000000000001061, 'system:permission-group:list', '权限组列表', 'API', NULL, 'GET', '/api/system/permission-group', 'ENABLED');
INSERT IGNORE INTO `t_rbac_permission` (`id`, `code`, `name`, `permission_type`, `ui_path`, `http_method`, `http_path`, `status`)
VALUES (100000000000001062, 'system:permission-group:detail', '权限组详情', 'API', NULL, 'GET', '/api/system/permission-group/detail', 'ENABLED');
INSERT IGNORE INTO `t_rbac_permission` (`id`, `code`, `name`, `permission_type`, `ui_path`, `http_method`, `http_path`, `status`)
VALUES (100000000000001063, 'system:permission-group:create', '创建权限组', 'API', NULL, 'POST', '/api/system/permission-group/create', 'ENABLED');
INSERT IGNORE INTO `t_rbac_permission` (`id`, `code`, `name`, `permission_type`, `ui_path`, `http_method`, `http_path`, `status`)
VALUES (100000000000001064, 'system:permission-group:update', '更新权限组', 'API', NULL, 'POST', '/api/system/permission-group/update', 'ENABLED');
INSERT IGNORE INTO `t_rbac_permission` (`id`, `code`, `name`, `permission_type`, `ui_path`, `http_method`, `http_path`, `status`)
VALUES (100000000000001065, 'system:permission-group:bind-permission', '权限组绑定权限', 'API', NULL, 'POST', '/api/system/permission-group/bind-permission', 'ENABLED');
INSERT IGNORE INTO `t_rbac_permission` (`id`, `code`, `name`, `permission_type`, `ui_path`, `http_method`, `http_path`, `status`)
VALUES (100000000000001066, 'system:permission-group:delete', '删除权限组', 'API', NULL, 'POST', '/api/system/permission-group/delete', 'ENABLED');
INSERT IGNORE INTO `t_rbac_permission` (`id`, `code`, `name`, `permission_type`, `ui_path`, `http_method`, `http_path`, `status`)
VALUES (100000000000001067, 'system:permission-group:status', '权限组启停', 'API', NULL, 'POST', '/api/system/permission-group/set-status', 'ENABLED');

-- 种子权限组：按现有菜单域给出可直接使用的权限包
INSERT IGNORE INTO `t_rbac_permission_group` (`id`, `code`, `name`, `status`, `remark`)
VALUES (100000000000003001, 'system_user', '用户管理', 'ENABLED', '用户菜单、用户 CRUD 与用户角色绑定');
INSERT IGNORE INTO `t_rbac_permission_group` (`id`, `code`, `name`, `status`, `remark`)
VALUES (100000000000003002, 'system_role', '角色管理', 'ENABLED', '角色菜单、角色 CRUD 与角色绑定权限组');
INSERT IGNORE INTO `t_rbac_permission_group` (`id`, `code`, `name`, `status`, `remark`)
VALUES (100000000000003003, 'system_permission', '权限管理', 'ENABLED', '权限目录与权限组的查看、维护');
INSERT IGNORE INTO `t_rbac_permission_group` (`id`, `code`, `name`, `status`, `remark`)
VALUES (100000000000003004, 'system_http_record', '访问记录', 'ENABLED', '访问记录菜单、列表与详情');

-- 用户管理组
INSERT IGNORE INTO `t_rbac_group_permission_rel` (`id`, `group_id`, `permission_id`)
VALUES (100000000000005001, 100000000000003001, 100000000000001001);
INSERT IGNORE INTO `t_rbac_group_permission_rel` (`id`, `group_id`, `permission_id`)
VALUES (100000000000005002, 100000000000003001, 100000000000001011);
INSERT IGNORE INTO `t_rbac_group_permission_rel` (`id`, `group_id`, `permission_id`)
VALUES (100000000000005003, 100000000000003001, 100000000000001012);
INSERT IGNORE INTO `t_rbac_group_permission_rel` (`id`, `group_id`, `permission_id`)
VALUES (100000000000005004, 100000000000003001, 100000000000001013);
INSERT IGNORE INTO `t_rbac_group_permission_rel` (`id`, `group_id`, `permission_id`)
VALUES (100000000000005005, 100000000000003001, 100000000000001014);
INSERT IGNORE INTO `t_rbac_group_permission_rel` (`id`, `group_id`, `permission_id`)
VALUES (100000000000005006, 100000000000003001, 100000000000001041);
INSERT IGNORE INTO `t_rbac_group_permission_rel` (`id`, `group_id`, `permission_id`)
VALUES (100000000000005007, 100000000000003001, 100000000000001042);

-- 角色管理组
INSERT IGNORE INTO `t_rbac_group_permission_rel` (`id`, `group_id`, `permission_id`)
VALUES (100000000000005011, 100000000000003002, 100000000000001002);
INSERT IGNORE INTO `t_rbac_group_permission_rel` (`id`, `group_id`, `permission_id`)
VALUES (100000000000005012, 100000000000003002, 100000000000001021);
INSERT IGNORE INTO `t_rbac_group_permission_rel` (`id`, `group_id`, `permission_id`)
VALUES (100000000000005013, 100000000000003002, 100000000000001022);
INSERT IGNORE INTO `t_rbac_group_permission_rel` (`id`, `group_id`, `permission_id`)
VALUES (100000000000005014, 100000000000003002, 100000000000001023);
INSERT IGNORE INTO `t_rbac_group_permission_rel` (`id`, `group_id`, `permission_id`)
VALUES (100000000000005015, 100000000000003002, 100000000000001024);
INSERT IGNORE INTO `t_rbac_group_permission_rel` (`id`, `group_id`, `permission_id`)
VALUES (100000000000005016, 100000000000003002, 100000000000001025);
INSERT IGNORE INTO `t_rbac_group_permission_rel` (`id`, `group_id`, `permission_id`)
VALUES (100000000000005017, 100000000000003002, 100000000000001026);
INSERT IGNORE INTO `t_rbac_group_permission_rel` (`id`, `group_id`, `permission_id`)
VALUES (100000000000005018, 100000000000003002, 100000000000001027);

-- 权限管理组
INSERT IGNORE INTO `t_rbac_group_permission_rel` (`id`, `group_id`, `permission_id`)
VALUES (100000000000005021, 100000000000003003, 100000000000001003);
INSERT IGNORE INTO `t_rbac_group_permission_rel` (`id`, `group_id`, `permission_id`)
VALUES (100000000000005022, 100000000000003003, 100000000000001005);
INSERT IGNORE INTO `t_rbac_group_permission_rel` (`id`, `group_id`, `permission_id`)
VALUES (100000000000005023, 100000000000003003, 100000000000001031);
INSERT IGNORE INTO `t_rbac_group_permission_rel` (`id`, `group_id`, `permission_id`)
VALUES (100000000000005024, 100000000000003003, 100000000000001032);
INSERT IGNORE INTO `t_rbac_group_permission_rel` (`id`, `group_id`, `permission_id`)
VALUES (100000000000005025, 100000000000003003, 100000000000001061);
INSERT IGNORE INTO `t_rbac_group_permission_rel` (`id`, `group_id`, `permission_id`)
VALUES (100000000000005026, 100000000000003003, 100000000000001062);
INSERT IGNORE INTO `t_rbac_group_permission_rel` (`id`, `group_id`, `permission_id`)
VALUES (100000000000005027, 100000000000003003, 100000000000001063);
INSERT IGNORE INTO `t_rbac_group_permission_rel` (`id`, `group_id`, `permission_id`)
VALUES (100000000000005028, 100000000000003003, 100000000000001064);
INSERT IGNORE INTO `t_rbac_group_permission_rel` (`id`, `group_id`, `permission_id`)
VALUES (100000000000005029, 100000000000003003, 100000000000001065);
INSERT IGNORE INTO `t_rbac_group_permission_rel` (`id`, `group_id`, `permission_id`)
VALUES (100000000000005041, 100000000000003003, 100000000000001033);
INSERT IGNORE INTO `t_rbac_group_permission_rel` (`id`, `group_id`, `permission_id`)
VALUES (100000000000005042, 100000000000003003, 100000000000001034);
INSERT IGNORE INTO `t_rbac_group_permission_rel` (`id`, `group_id`, `permission_id`)
VALUES (100000000000005043, 100000000000003003, 100000000000001035);
INSERT IGNORE INTO `t_rbac_group_permission_rel` (`id`, `group_id`, `permission_id`)
VALUES (100000000000005044, 100000000000003003, 100000000000001066);
INSERT IGNORE INTO `t_rbac_group_permission_rel` (`id`, `group_id`, `permission_id`)
VALUES (100000000000005045, 100000000000003003, 100000000000001067);

-- 访问记录组
INSERT IGNORE INTO `t_rbac_group_permission_rel` (`id`, `group_id`, `permission_id`)
VALUES (100000000000005031, 100000000000003004, 100000000000001004);
INSERT IGNORE INTO `t_rbac_group_permission_rel` (`id`, `group_id`, `permission_id`)
VALUES (100000000000005032, 100000000000003004, 100000000000001051);
INSERT IGNORE INTO `t_rbac_group_permission_rel` (`id`, `group_id`, `permission_id`)
VALUES (100000000000005033, 100000000000003004, 100000000000001052);

-- 内置管理员绑定超管角色
INSERT IGNORE INTO `t_rbac_user_role_rel` (`id`, `user_id`, `role_id`, `create_by`, `update_by`)
SELECT 100000000000002001, u.id, r.id, u.id, u.id
FROM `t_user` u
JOIN `t_rbac_role` r ON r.code = 'super_admin'
WHERE u.username = 'admin' AND u.delete_time IS NULL;

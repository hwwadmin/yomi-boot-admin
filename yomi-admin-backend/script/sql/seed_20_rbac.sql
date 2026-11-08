-- RBAC 域种子数据（先执行 init.sql 建表、seed_10_user.sql 灌内置管理员）
-- 角色、权限目录、权限组及关系，以及内置超管绑定
SET NAMES utf8mb4;

-- 后端统一使用 UTC，种子时间也按 UTC 写入
SET time_zone = '+00:00';

-- 角色直连权限的旧关系表已废弃，权限一律经权限组授予
DROP TABLE IF EXISTS `t_rbac_role_permission_rel`;

-- 内置超管角色
INSERT IGNORE INTO `t_rbac_role` (`id`, `code`, `name`, `builtin`, `status`, `remark`)
VALUES (100000000000000101, 'super_admin', '超级管理员', 1, 'ENABLED', '内置角色，拥有全部权限');

-- 接口权限目录与权限组：种子只负责初始化，管理端可继续维护
-- 目录权限：只作分组节点，路由守卫与侧栏展开不依赖它，授予后可在权限目录中显式管理目录层级
INSERT IGNORE INTO `t_rbac_permission` (`id`, `code`, `name`, `permission_type`, `ui_path`, `http_method`, `http_path`, `status`)
VALUES (100000000000000981, 'system', '系统管理', 'MENU', '/system', NULL, NULL, 'ENABLED');
INSERT IGNORE INTO `t_rbac_permission` (`id`, `code`, `name`, `permission_type`, `ui_path`, `http_method`, `http_path`, `status`)
VALUES (100000000000000982, 'system:rbac', '权限管理', 'MENU', '/system/rbac', NULL, NULL, 'ENABLED');

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
UPDATE `t_rbac_permission` SET `name` = '系统参数' WHERE `code` = 'system:param' AND `permission_type` = 'MENU';

-- 用户接口权限
INSERT IGNORE INTO `t_rbac_permission` (`id`, `code`, `name`, `permission_type`, `ui_path`, `http_method`, `http_path`, `status`)
VALUES (100000000000001011, 'system:user:list', '用户列表', 'API', NULL, 'GET', '/api/system/user', 'ENABLED');
INSERT IGNORE INTO `t_rbac_permission` (`id`, `code`, `name`, `permission_type`, `ui_path`, `http_method`, `http_path`, `status`)
VALUES (100000000000001012, 'system:user:create', '创建用户', 'API', NULL, 'POST', '/api/system/user/create', 'ENABLED');
INSERT IGNORE INTO `t_rbac_permission` (`id`, `code`, `name`, `permission_type`, `ui_path`, `http_method`, `http_path`, `status`)
VALUES (100000000000001013, 'system:user:update', '更新用户', 'API', NULL, 'POST', '/api/system/user/update', 'ENABLED');
INSERT IGNORE INTO `t_rbac_permission` (`id`, `code`, `name`, `permission_type`, `ui_path`, `http_method`, `http_path`, `status`)
VALUES (100000000000001014, 'system:user:delete', '删除用户', 'API', NULL, 'POST', '/api/system/user/delete', 'ENABLED');
-- 重置密码属于敏感操作，后端只允许启用态超管执行；此处仅提供目录项与按钮显隐
INSERT IGNORE INTO `t_rbac_permission` (`id`, `code`, `name`, `permission_type`, `ui_path`, `http_method`, `http_path`, `status`)
VALUES (100000000000001015, 'system:user:reset-password', '重置用户密码', 'API', NULL, 'POST', '/api/system/user/reset-password', 'ENABLED');

-- 用户会话接口权限：挂在用户管理页的抽屉里，随用户菜单一起授予
INSERT IGNORE INTO `t_rbac_permission` (`id`, `code`, `name`, `permission_type`, `ui_path`, `http_method`, `http_path`, `status`)
VALUES (100000000000001016, 'system:user:session:list', '用户会话列表', 'API', NULL, 'GET', '/api/system/user/session', 'ENABLED');
INSERT IGNORE INTO `t_rbac_permission` (`id`, `code`, `name`, `permission_type`, `ui_path`, `http_method`, `http_path`, `status`)
VALUES (100000000000001017, 'system:user:session:revoke', '强制下线会话', 'API', NULL, 'POST', '/api/system/user/session/revoke', 'ENABLED');
INSERT IGNORE INTO `t_rbac_permission` (`id`, `code`, `name`, `permission_type`, `ui_path`, `http_method`, `http_path`, `status`)
VALUES (100000000000001018, 'system:user:session:revoke-all', '强制下线用户全部会话', 'API', NULL, 'POST', '/api/system/user/session/revoke-all', 'ENABLED');

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

-- 鉴权缓存接口权限
INSERT IGNORE INTO `t_rbac_permission` (`id`, `code`, `name`, `permission_type`, `ui_path`, `http_method`, `http_path`, `status`)
VALUES (100000000000001071, 'system:rbac_cache:refresh', '刷新权限缓存', 'API', NULL, 'POST', '/api/system/rbac-cache/refresh', 'ENABLED');

-- 系统参数菜单与接口权限
INSERT IGNORE INTO `t_rbac_permission` (`id`, `code`, `name`, `permission_type`, `ui_path`, `http_method`, `http_path`, `status`)
VALUES (100000000000001006, 'system:param', '系统参数', 'MENU', '/system/param', NULL, NULL, 'ENABLED');
INSERT IGNORE INTO `t_rbac_permission` (`id`, `code`, `name`, `permission_type`, `ui_path`, `http_method`, `http_path`, `status`)
VALUES (100000000000001081, 'system:param:list', '参数列表', 'API', NULL, 'GET', '/api/system/param', 'ENABLED');
INSERT IGNORE INTO `t_rbac_permission` (`id`, `code`, `name`, `permission_type`, `ui_path`, `http_method`, `http_path`, `status`)
VALUES (100000000000001082, 'system:param:detail', '参数详情', 'API', NULL, 'GET', '/api/system/param/detail', 'ENABLED');
INSERT IGNORE INTO `t_rbac_permission` (`id`, `code`, `name`, `permission_type`, `ui_path`, `http_method`, `http_path`, `status`)
VALUES (100000000000001083, 'system:param:create', '创建参数', 'API', NULL, 'POST', '/api/system/param/create', 'ENABLED');
INSERT IGNORE INTO `t_rbac_permission` (`id`, `code`, `name`, `permission_type`, `ui_path`, `http_method`, `http_path`, `status`)
VALUES (100000000000001084, 'system:param:update', '更新参数', 'API', NULL, 'POST', '/api/system/param/update', 'ENABLED');
INSERT IGNORE INTO `t_rbac_permission` (`id`, `code`, `name`, `permission_type`, `ui_path`, `http_method`, `http_path`, `status`)
VALUES (100000000000001085, 'system:param:delete', '删除参数', 'API', NULL, 'POST', '/api/system/param/delete', 'ENABLED');
INSERT IGNORE INTO `t_rbac_permission` (`id`, `code`, `name`, `permission_type`, `ui_path`, `http_method`, `http_path`, `status`)
VALUES (100000000000001086, 'system:param:status', '参数启停', 'API', NULL, 'POST', '/api/system/param/set-status', 'ENABLED');

-- 审计日志菜单与接口权限
INSERT IGNORE INTO `t_rbac_permission` (`id`, `code`, `name`, `permission_type`, `ui_path`, `http_method`, `http_path`, `status`)
VALUES (100000000000001007, 'system:audit-log', '审计日志', 'MENU', '/system/audit-log', NULL, NULL, 'ENABLED');
INSERT IGNORE INTO `t_rbac_permission` (`id`, `code`, `name`, `permission_type`, `ui_path`, `http_method`, `http_path`, `status`)
VALUES (100000000000001091, 'system:audit-log:list', '审计日志列表', 'API', NULL, 'GET', '/api/system/audit-log', 'ENABLED');

-- 种子权限组：按现有菜单域给出可直接使用的权限包
INSERT IGNORE INTO `t_rbac_permission_group` (`id`, `code`, `name`, `status`, `remark`)
VALUES (100000000000003001, 'system_user', '用户管理', 'ENABLED', '用户菜单、用户 CRUD 与用户角色绑定');
INSERT IGNORE INTO `t_rbac_permission_group` (`id`, `code`, `name`, `status`, `remark`)
VALUES (100000000000003002, 'system_role', '角色管理', 'ENABLED', '角色菜单、角色 CRUD 与角色绑定权限组');
INSERT IGNORE INTO `t_rbac_permission_group` (`id`, `code`, `name`, `status`, `remark`)
VALUES (100000000000003003, 'system_permission', '权限管理', 'ENABLED', '权限目录与权限组的查看、维护');
INSERT IGNORE INTO `t_rbac_permission_group` (`id`, `code`, `name`, `status`, `remark`)
VALUES (100000000000003004, 'system_http_record', '访问记录', 'ENABLED', '访问记录菜单、列表与详情');
INSERT IGNORE INTO `t_rbac_permission_group` (`id`, `code`, `name`, `status`, `remark`)
VALUES (100000000000003005, 'system_param', '参数管理', 'ENABLED', '系统参数菜单、CRUD 与启停');
INSERT IGNORE INTO `t_rbac_permission_group` (`id`, `code`, `name`, `status`, `remark`)
VALUES (100000000000003006, 'system_audit_log', '审计日志', 'ENABLED', '审计日志菜单与列表');

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
-- 重置密码按钮归入用户管理组；实际执行仍被后端限制为启用态超管
INSERT IGNORE INTO `t_rbac_group_permission_rel` (`id`, `group_id`, `permission_id`)
VALUES (100000000000005008, 100000000000003001, 100000000000001015);
-- 会话查看与下线归入用户管理组：能停用用户的人本来就能撤销其全部令牌，不构成额外提权
INSERT IGNORE INTO `t_rbac_group_permission_rel` (`id`, `group_id`, `permission_id`)
VALUES (100000000000005091, 100000000000003001, 100000000000001016);
INSERT IGNORE INTO `t_rbac_group_permission_rel` (`id`, `group_id`, `permission_id`)
VALUES (100000000000005092, 100000000000003001, 100000000000001017);
INSERT IGNORE INTO `t_rbac_group_permission_rel` (`id`, `group_id`, `permission_id`)
VALUES (100000000000005093, 100000000000003001, 100000000000001018);

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
-- 刷新缓存按钮随角色页展示，归入角色管理组
INSERT IGNORE INTO `t_rbac_group_permission_rel` (`id`, `group_id`, `permission_id`)
VALUES (100000000000005051, 100000000000003002, 100000000000001071);

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

-- 参数管理组
INSERT IGNORE INTO `t_rbac_group_permission_rel` (`id`, `group_id`, `permission_id`)
VALUES (100000000000005061, 100000000000003005, 100000000000001006);
INSERT IGNORE INTO `t_rbac_group_permission_rel` (`id`, `group_id`, `permission_id`)
VALUES (100000000000005062, 100000000000003005, 100000000000001081);
INSERT IGNORE INTO `t_rbac_group_permission_rel` (`id`, `group_id`, `permission_id`)
VALUES (100000000000005063, 100000000000003005, 100000000000001082);
INSERT IGNORE INTO `t_rbac_group_permission_rel` (`id`, `group_id`, `permission_id`)
VALUES (100000000000005064, 100000000000003005, 100000000000001083);
INSERT IGNORE INTO `t_rbac_group_permission_rel` (`id`, `group_id`, `permission_id`)
VALUES (100000000000005065, 100000000000003005, 100000000000001084);
INSERT IGNORE INTO `t_rbac_group_permission_rel` (`id`, `group_id`, `permission_id`)
VALUES (100000000000005066, 100000000000003005, 100000000000001085);
INSERT IGNORE INTO `t_rbac_group_permission_rel` (`id`, `group_id`, `permission_id`)
VALUES (100000000000005067, 100000000000003005, 100000000000001086);

-- 审计日志组
INSERT IGNORE INTO `t_rbac_group_permission_rel` (`id`, `group_id`, `permission_id`)
VALUES (100000000000005071, 100000000000003006, 100000000000001007);
INSERT IGNORE INTO `t_rbac_group_permission_rel` (`id`, `group_id`, `permission_id`)
VALUES (100000000000005072, 100000000000003006, 100000000000001091);

-- 种子组补挂目录权限：组内叶子菜单挂在哪个目录，就一并授予该目录链，权限组作为完整菜单包使用
INSERT IGNORE INTO `t_rbac_group_permission_rel` (`id`, `group_id`, `permission_id`)
VALUES (100000000000005111, 100000000000003001, 100000000000000981);
INSERT IGNORE INTO `t_rbac_group_permission_rel` (`id`, `group_id`, `permission_id`)
VALUES (100000000000005112, 100000000000003002, 100000000000000981);
INSERT IGNORE INTO `t_rbac_group_permission_rel` (`id`, `group_id`, `permission_id`)
VALUES (100000000000005113, 100000000000003002, 100000000000000982);
INSERT IGNORE INTO `t_rbac_group_permission_rel` (`id`, `group_id`, `permission_id`)
VALUES (100000000000005114, 100000000000003003, 100000000000000981);
INSERT IGNORE INTO `t_rbac_group_permission_rel` (`id`, `group_id`, `permission_id`)
VALUES (100000000000005115, 100000000000003003, 100000000000000982);
INSERT IGNORE INTO `t_rbac_group_permission_rel` (`id`, `group_id`, `permission_id`)
VALUES (100000000000005116, 100000000000003004, 100000000000000981);
INSERT IGNORE INTO `t_rbac_group_permission_rel` (`id`, `group_id`, `permission_id`)
VALUES (100000000000005117, 100000000000003005, 100000000000000981);
INSERT IGNORE INTO `t_rbac_group_permission_rel` (`id`, `group_id`, `permission_id`)
VALUES (100000000000005118, 100000000000003006, 100000000000000981);

-- 内置管理员绑定超管角色（依赖 seed_10_user.sql 已灌入 admin）
INSERT IGNORE INTO `t_rbac_user_role_rel` (`id`, `user_id`, `role_id`, `create_by`, `update_by`)
SELECT 100000000000002001, u.id, r.id, u.id, u.id
FROM `t_user` u
JOIN `t_rbac_role` r ON r.code = 'super_admin'
WHERE u.username = 'admin' AND u.delete_time IS NULL;

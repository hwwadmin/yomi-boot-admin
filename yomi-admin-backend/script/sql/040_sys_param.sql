-- 系统参数表（动态配置数据源；权限目录种子见 020_rbac.sql）
SET NAMES utf8mb4;
USE yomi_admin;

-- 后端统一使用 UTC，种子时间也按 UTC 写入
SET time_zone = '+00:00';

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

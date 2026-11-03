-- 业务幂等表（sys_biz）
-- 表名与 yomi-boot 的 idempotency 组件保持一致；权威 DDL 由本仓维护，框架升级改表结构时同步修改这里。
SET NAMES utf8mb4;
USE yomi_admin;

-- 后端统一使用 UTC，默认时间也按 UTC 写入
SET time_zone = '+00:00';

CREATE TABLE IF NOT EXISTS `sys_biz` (
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

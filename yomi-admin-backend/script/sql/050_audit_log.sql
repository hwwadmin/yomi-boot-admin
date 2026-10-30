-- 操作审计日志表：按需记录需要关注的业务操作
SET NAMES utf8mb4;
USE yomi_admin;

-- 后端统一使用 UTC
SET time_zone = '+00:00';

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

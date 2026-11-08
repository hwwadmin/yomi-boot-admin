-- HTTP 访问记录表（httprecord 采集后由业务消费者落库）
SET NAMES utf8mb4;
USE yomi_admin;

-- 后端统一使用 UTC，记录时间按 UTC 写入
SET time_zone = '+00:00';

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

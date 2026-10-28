-- 系统用户表与内置管理员
USE yomi_admin;

-- 后端统一使用 UTC，种子时间也按 UTC 写入
SET time_zone = '+00:00';

CREATE TABLE IF NOT EXISTS sys_user (
    id            BIGINT       NOT NULL COMMENT '主键（雪花ID）',
    username      VARCHAR(64)  NOT NULL COMMENT '登录名',
    password_hash VARCHAR(255) NOT NULL COMMENT 'bcrypt 哈希',
    nickname      VARCHAR(64)  NOT NULL DEFAULT '' COMMENT '昵称',
    email         VARCHAR(128) NOT NULL DEFAULT '' COMMENT '邮箱',
    phone         VARCHAR(32)  NOT NULL DEFAULT '' COMMENT '手机号',
    status        TINYINT      NOT NULL DEFAULT 1 COMMENT '0 禁用 / 1 正常',
    create_time   DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    create_by     BIGINT       DEFAULT NULL COMMENT '创建人',
    update_time   DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    update_by     BIGINT       DEFAULT NULL COMMENT '更新人',
    delete_time   DATETIME     DEFAULT NULL COMMENT '删除时间',
    delete_by     BIGINT       DEFAULT NULL COMMENT '删除人',
    PRIMARY KEY (id),
    UNIQUE KEY uk_sys_user_username (username),
    KEY idx_sys_user_delete_time (delete_time)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='系统用户';

-- 内置管理员 admin / admin123（bcrypt cost 10）
INSERT INTO sys_user (id, username, password_hash, nickname, status, create_time, update_time)
VALUES (100000000000000001, 'admin', '$2a$10$3cgxANk0gaQbx.Sz7QlF6uNxREYPJIXV1zz8GahaLOIRfUkhqwiUy', '超级管理员', 1, NOW(), NOW()) AS seed
ON DUPLICATE KEY UPDATE nickname = seed.nickname;

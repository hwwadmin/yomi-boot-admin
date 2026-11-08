-- 用户域种子数据（先执行 init.sql 建表）
SET NAMES utf8mb4;

-- 后端统一使用 UTC，种子时间也按 UTC 写入
SET time_zone = '+00:00';

-- 内置管理员 admin / admin123（bcrypt cost 10）
INSERT INTO `t_user` (`id`, `username`, `password_hash`, `nickname`, `status`, `create_time`, `update_time`)
VALUES (100000000000000001, 'admin', '$2a$10$3cgxANk0gaQbx.Sz7QlF6uNxREYPJIXV1zz8GahaLOIRfUkhqwiUy', '超级管理员', 'ENABLED', NOW(), NOW()) AS seed
ON DUPLICATE KEY UPDATE `nickname` = seed.`nickname`;

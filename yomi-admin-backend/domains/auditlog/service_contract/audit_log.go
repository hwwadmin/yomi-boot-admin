// Package service_contract 操作审计日志域服务契约（共享内核：零域间依赖）
package service_contract

import (
	"context"
	"yomi-admin-backend/domains/auditlog/pkg/shared/input"
	"yomi-admin-backend/domains/auditlog/pkg/shared/output"
)

// AuditLogWriter 审计日志写入能力
type AuditLogWriter interface {
	// Record 写入一条审计记录
	Record(ctx context.Context, in *input.RecordInput) error
}

// AuditLogReader 审计日志查询能力
type AuditLogReader interface {
	// Page 分页查询审计日志
	Page(ctx context.Context, in *input.PageInput) (*output.PageLogsOutput, error)
}

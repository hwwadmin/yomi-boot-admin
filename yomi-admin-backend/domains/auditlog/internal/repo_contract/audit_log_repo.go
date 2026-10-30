// Package repo_contract 操作审计日志仓储契约
package repo_contract

import (
	"context"
	"yomi-admin-backend/domains/auditlog/pkg/shared/entity"
	"yomi-admin-backend/domains/auditlog/pkg/shared/enums"
)

// Filter 审计日志分页过滤条件
type Filter struct {
	Keyword   string
	EventCode string
	Phase     enums.Phase
	Result    enums.Result
	AuditID   int64
	StartTime int64
	EndTime   int64
	Page      int
	PageSize  int
}

// AuditLogRepo 操作审计日志仓储
type AuditLogRepo interface {
	// Create 写入一条审计记录
	Create(ctx context.Context, log *entity.Log) error
	// Page 分页查询审计日志，按事件时间与主键倒序
	Page(ctx context.Context, filter *Filter) ([]*entity.Log, int64, error)
}

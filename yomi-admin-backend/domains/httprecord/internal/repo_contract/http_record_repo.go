// Package repo_contract HTTP 访问记录仓储契约
package repo_contract

import (
	"context"
	"yomi-admin-backend/domains/httprecord/pkg/shared/entity"
)

// RecordFilter 访问记录分页过滤条件
type RecordFilter struct {
	Keyword     string
	Method      string
	StatusClass string
	UserID      int64
	StartTime   int64 // Unix 毫秒时间戳
	EndTime     int64 // Unix 毫秒时间戳
	Page        int
	PageSize    int
}

// RecordRepo 访问记录仓储
type RecordRepo interface {
	// Create 写入访问记录；trace_id + request_at 重复时忽略
	Create(ctx context.Context, record *entity.Record) error
	// FindByID 按主键查询访问记录
	FindByID(ctx context.Context, id int64) (*entity.Record, bool, error)
	// Page 分页查询访问记录，按请求时间倒序
	Page(ctx context.Context, filter *RecordFilter) ([]*entity.Record, int64, error)
}

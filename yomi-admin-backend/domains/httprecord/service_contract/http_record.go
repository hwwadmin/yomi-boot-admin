// Package service_contract HTTP 访问记录域服务契约
package service_contract

import (
	"context"
	"yomi-admin-backend/domains/httprecord/pkg/shared/entity"
	"yomi-admin-backend/domains/httprecord/pkg/shared/input"
	"yomi-admin-backend/domains/httprecord/pkg/shared/output"
)

// RecordService 访问记录领域服务
type RecordService interface {
	// Save 持久化一条访问记录
	Save(ctx context.Context, in *input.SaveRecordInput) error
	// Page 分页查询访问记录
	Page(ctx context.Context, in *input.PageRecordsInput) (*output.PageRecordsOutput, error)
	// Get 按主键查询访问记录详情
	Get(ctx context.Context, id int64) (*entity.Record, error)
}

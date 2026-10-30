// Package output 操作审计日志域出参
package output

import "yomi-admin-backend/domains/auditlog/pkg/shared/entity"

// PageLogsOutput 审计日志分页出参
type PageLogsOutput struct {
	Items []*entity.Log
	Total int64
}

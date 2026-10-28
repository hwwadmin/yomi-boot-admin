// Package output HTTP 访问记录域出参
package output

import "yomi-admin-backend/domains/httprecord/pkg/shared/entity"

// PageRecordsOutput 访问记录分页出参
type PageRecordsOutput struct {
	Items []*entity.Record
	Total int64
}

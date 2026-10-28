// Package httprecord HTTP 访问记录编排
package httprecord

import (
	"yomi-admin-backend/domains/httprecord/pkg/shared/entity"

	"gitee.com/hq_project/yomi-boot/component/core/shared/dto"
)

// RecordDetail 访问记录详情
type RecordDetail struct {
	RecordItem
	Query                 string `json:"query"`
	UserAgent             string `json:"userAgent"`
	RequestHeaders        string `json:"requestHeaders"`
	RequestBody           string `json:"requestBody"`
	RequestBodyTruncated  bool   `json:"requestBodyTruncated"`
	ResponseBody          string `json:"responseBody"`
	ResponseBodyTruncated bool   `json:"responseBodyTruncated"`
	CreateTime            int64  `json:"createTime"`
}

// toRecordItem 将领域实体转换为列表项
func toRecordItem(record *entity.Record) RecordItem {
	return RecordItem{
		ID:             dto.JSONInt64(record.ID),
		ServiceType:    record.ServiceType,
		ServiceVersion: record.ServiceVersion,
		TraceID:        record.TraceID,
		RequestAt:      record.RequestAt,
		DurationMS:     record.DurationMS,
		Method:         record.Method,
		Path:           record.Path,
		Route:          record.Route,
		ClientIP:       record.ClientIP,
		UserID:         dto.JSONInt64(record.UserID),
		ClientType:     record.ClientType,
		StatusCode:     record.StatusCode,
		BizCode:        record.BizCode,
		BizMsg:         record.BizMsg,
	}
}

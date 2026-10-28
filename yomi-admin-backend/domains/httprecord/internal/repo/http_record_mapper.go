// Package repo HTTP 访问记录仓储实现
package repo

import (
	"time"
	"yomi-admin-backend/domains/httprecord/pkg/shared/entity"
)

// toRecordEntity 将持久化模型转换为领域实体
func toRecordEntity(row *httpRecordModel) *entity.Record {
	if row == nil {
		return nil
	}
	return &entity.Record{
		ID:                    row.ID,
		ServiceType:           row.ServiceType,
		ServiceVersion:        row.ServiceVersion,
		TraceID:               row.TraceID,
		RequestAt:             row.RequestAt.UnixMilli(),
		DurationMS:            row.DurationMS,
		Method:                row.Method,
		Path:                  row.Path,
		Route:                 row.Route,
		Query:                 row.Query,
		ClientIP:              row.ClientIP,
		UserAgent:             row.UserAgent,
		RequestHeaders:        row.RequestHeaders,
		RequestBody:           row.RequestBody,
		RequestBodyTruncated:  row.RequestBodyTruncated,
		UserID:                row.UserID,
		ClientType:            row.ClientType,
		StatusCode:            row.StatusCode,
		ResponseBody:          row.ResponseBody,
		ResponseBodyTruncated: row.ResponseBodyTruncated,
		BizCode:               row.BizCode,
		BizMsg:                row.BizMsg,
		CreateTime:            row.CreateTime.UnixMilli(),
	}
}

// toRecordModel 将领域实体转换为持久化模型
func toRecordModel(record *entity.Record) *httpRecordModel {
	if record == nil {
		return nil
	}
	row := &httpRecordModel{
		ServiceType:           record.ServiceType,
		ServiceVersion:        record.ServiceVersion,
		TraceID:               record.TraceID,
		RequestAt:             time.UnixMilli(record.RequestAt),
		DurationMS:            record.DurationMS,
		Method:                record.Method,
		Path:                  record.Path,
		Route:                 record.Route,
		Query:                 record.Query,
		ClientIP:              record.ClientIP,
		UserAgent:             record.UserAgent,
		RequestHeaders:        record.RequestHeaders,
		RequestBody:           record.RequestBody,
		RequestBodyTruncated:  record.RequestBodyTruncated,
		UserID:                record.UserID,
		ClientType:            record.ClientType,
		StatusCode:            record.StatusCode,
		ResponseBody:          record.ResponseBody,
		ResponseBodyTruncated: record.ResponseBodyTruncated,
		BizCode:               record.BizCode,
		BizMsg:                record.BizMsg,
	}
	row.ID = record.ID
	return row
}

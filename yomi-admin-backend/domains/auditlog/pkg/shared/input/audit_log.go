// Package input 操作审计日志域入参
package input

import "yomi-admin-backend/domains/auditlog/pkg/shared/enums"

// RecordInput 写入一条审计记录
type RecordInput struct {
	AuditID      int64
	Phase        enums.Phase
	EventCode    string
	OperatorID   int64
	OperatorName string
	ClientIP     string
	TargetType   string
	TargetID     int64
	Result       enums.Result
	BizCode      string
	CreateTime   int64 // Unix 毫秒时间戳；0 表示使用当前时间
}

// PageInput 审计日志分页入参
type PageInput struct {
	Keyword   string
	EventCode string
	Phase     enums.Phase
	Result    enums.Result
	AuditID   int64
	StartTime int64 // Unix 毫秒时间戳，含
	EndTime   int64 // Unix 毫秒时间戳，含
	Page      int
	PageSize  int
}

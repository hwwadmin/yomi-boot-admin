// Package entity 操作审计日志实体
package entity

import "yomi-admin-backend/domains/auditlog/pkg/shared/enums"

// Log 表示一条操作审计记录（非持久化模型）
type Log struct {
	ID           int64
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
	CreateTime   int64 // Unix 毫秒时间戳
}

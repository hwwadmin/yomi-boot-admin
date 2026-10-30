// Package repo 操作审计日志仓储实现
package repo

import (
	"time"
	"yomi-admin-backend/domains/auditlog/pkg/shared/entity"
)

// toLogEntity 将持久化模型转换为领域实体
func toLogEntity(row *auditLogModel) *entity.Log {
	if row == nil {
		return nil
	}
	return &entity.Log{
		ID:           row.ID,
		AuditID:      row.AuditID,
		Phase:        row.Phase,
		EventCode:    row.EventCode,
		OperatorID:   row.OperatorID,
		OperatorName: row.OperatorName,
		ClientIP:     row.ClientIP,
		TargetType:   row.TargetType,
		TargetID:     row.TargetID,
		Result:       row.Result,
		BizCode:      row.BizCode,
		CreateTime:   row.CreateTime.UnixMilli(),
	}
}

// toLogModel 将领域实体转换为持久化模型
func toLogModel(log *entity.Log) *auditLogModel {
	if log == nil {
		return nil
	}
	row := &auditLogModel{
		AuditID:      log.AuditID,
		Phase:        log.Phase,
		EventCode:    log.EventCode,
		OperatorID:   log.OperatorID,
		OperatorName: log.OperatorName,
		ClientIP:     log.ClientIP,
		TargetType:   log.TargetType,
		TargetID:     log.TargetID,
		Result:       log.Result,
		BizCode:      log.BizCode,
	}
	row.ID = log.ID
	row.CreateTime = time.UnixMilli(log.CreateTime).UTC()
	return row
}

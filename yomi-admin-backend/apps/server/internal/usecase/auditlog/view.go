package auditlog

import (
	"yomi-admin-backend/domains/auditlog/pkg/shared/entity"

	"gitee.com/hq_project/yomi-boot/component/core/shared/dto"
)

// toLogItem 将审计日志实体转换为列表项；目标用户名不落库，由 fillTargetNames 回填
func toLogItem(log *entity.Log) LogItem {
	if log == nil {
		return LogItem{}
	}
	return LogItem{
		ID:           dto.JSONInt64(log.ID),
		AuditID:      dto.JSONInt64(log.AuditID),
		Phase:        log.Phase,
		EventCode:    log.EventCode,
		OperatorID:   dto.JSONInt64(log.OperatorID),
		OperatorName: log.OperatorName,
		ClientIP:     log.ClientIP,
		TargetType:   log.TargetType,
		TargetID:     dto.JSONInt64(log.TargetID),
		Result:       log.Result,
		BizCode:      log.BizCode,
		CreateTime:   log.CreateTime,
	}
}

// Package repo 操作审计日志仓储实现
package repo

import (
	"yomi-admin-backend/domains/auditlog/pkg/shared/enums"

	"gitee.com/hq_project/yomi-boot/component/common/db/orm"
)

// tableName 是操作审计日志表名
const tableName = "t_audit_log"

// auditLogModel 是 t_audit_log 持久化模型；审计只追加，统一字段复用 orm.BaseModel。
type auditLogModel struct {
	AuditID      int64        `gorm:"column:audit_id"`
	Phase        enums.Phase  `gorm:"column:phase;size:16"`
	EventCode    string       `gorm:"column:event_code;size:64"`
	OperatorID   int64        `gorm:"column:operator_id"`
	OperatorName string       `gorm:"column:operator_name;size:64"`
	ClientIP     string       `gorm:"column:client_ip;size:64"`
	TargetType   string       `gorm:"column:target_type;size:32"`
	TargetID     int64        `gorm:"column:target_id"`
	Result       enums.Result `gorm:"column:result;size:16"`
	BizCode      string       `gorm:"column:biz_code;size:64"`
	orm.BaseModel
}

// TableName 返回表名
func (auditLogModel) TableName() string {
	return tableName
}

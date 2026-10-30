// Package repo 系统参数仓储实现
package repo

import (
	"gitee.com/hq_project/yomi-boot/component/common/db/orm"
	"yomi-admin-backend/domains/sysparam/pkg/shared/enums"
)

// tableName 是系统参数表名
const tableName = "t_sys_param"

// sysParamModel 是 t_sys_param 的持久化模型
type sysParamModel struct {
	Key       string               `gorm:"column:param_key;size:128;uniqueIndex:uk_param_key"`
	Value     string               `gorm:"column:param_value;size:2048"`
	ValueType enums.ParamValueType `gorm:"column:value_type;size:20"`
	Name      string               `gorm:"column:name;size:64"`
	Remark    string               `gorm:"column:remark;size:255"`
	Status    enums.Status         `gorm:"column:status;size:20;index:idx_status"`
	orm.BaseModel
}

// TableName 返回表名
func (sysParamModel) TableName() string {
	return tableName
}

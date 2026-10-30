// Package input 系统参数域服务入参
package input

import "yomi-admin-backend/domains/sysparam/pkg/shared/enums"

// PageParamsInput 参数分页查询入参
type PageParamsInput struct {
	Keyword   string
	Status    enums.Status
	ValueType enums.ParamValueType
	Page      int
	PageSize  int
}

// CreateParamInput 创建参数入参；键创建后不可变
type CreateParamInput struct {
	Key        string
	Value      string
	ValueType  enums.ParamValueType
	Name       string
	Remark     string
	OperatorID int64
}

// UpdateParamInput 更新参数入参；键与值类型不可变，启停走独立入参
type UpdateParamInput struct {
	ID         int64
	Value      string
	Name       string
	Remark     string
	OperatorID int64
}

// SetParamStatusInput 设置参数启停入参
type SetParamStatusInput struct {
	ID         int64
	Status     enums.Status
	OperatorID int64
}

// DeleteParamInput 删除参数入参
type DeleteParamInput struct {
	ID         int64
	OperatorID int64
}

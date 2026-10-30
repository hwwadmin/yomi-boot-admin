// Package repo 系统参数仓储实现
package repo

import (
	"yomi-admin-backend/domains/sysparam/pkg/shared/entity"
)

// toParamEntity 将持久化模型转换为领域实体
func toParamEntity(row *sysParamModel) *entity.Param {
	if row == nil {
		return nil
	}
	return &entity.Param{
		ID:         row.ID,
		Key:        row.Key,
		Value:      row.Value,
		ValueType:  row.ValueType,
		Name:       row.Name,
		Remark:     row.Remark,
		Status:     row.Status,
		CreateTime: row.CreateTime.UnixMilli(),
		UpdateTime: row.UpdateTime.UnixMilli(),
	}
}

// toParamModel 将领域实体转换为持久化模型；operatorID 写入 create_by
func toParamModel(param *entity.Param, operatorID int64) *sysParamModel {
	if param == nil {
		return nil
	}
	row := &sysParamModel{
		Key:       param.Key,
		Value:     param.Value,
		ValueType: param.ValueType,
		Name:      param.Name,
		Remark:    param.Remark,
		Status:    param.Status,
	}
	row.ID = param.ID
	if operatorID > 0 {
		row.CreateBy = &operatorID
	}
	return row
}

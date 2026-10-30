// Package service_contract 系统参数域服务契约（共享内核：零域间依赖，业务域可直接依赖）
package service_contract

import (
	"context"
	"yomi-admin-backend/domains/sysparam/pkg/shared/entity"
	"yomi-admin-backend/domains/sysparam/pkg/shared/input"
	"yomi-admin-backend/domains/sysparam/pkg/shared/output"
)

// SysParamAdminService 参数管理端服务
type SysParamAdminService interface {
	// PageParams 分页查询参数
	PageParams(ctx context.Context, in *input.PageParamsInput) (*output.PageParamsOutput, error)
	// GetParam 按主键查询参数详情
	GetParam(ctx context.Context, id int64) (*entity.Param, error)
	// CreateParam 创建参数；键全局唯一且创建后不可变
	CreateParam(ctx context.Context, in *input.CreateParamInput) (*entity.Param, error)
	// UpdateParam 更新参数值、名称与备注（键与值类型不可变）
	UpdateParam(ctx context.Context, in *input.UpdateParamInput) error
	// SetParamStatus 设置参数启停；禁用即覆盖值失效，读取方回退默认值
	SetParamStatus(ctx context.Context, in *input.SetParamStatusInput) error
	// DeleteParam 物理删除参数
	DeleteParam(ctx context.Context, in *input.DeleteParamInput) error
}

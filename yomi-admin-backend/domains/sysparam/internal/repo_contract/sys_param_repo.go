// Package repo_contract 系统参数仓储契约
package repo_contract

import (
	"context"
	"yomi-admin-backend/domains/sysparam/pkg/shared/entity"
	"yomi-admin-backend/domains/sysparam/pkg/shared/enums"
)

// ParamFilter 参数分页过滤条件
type ParamFilter struct {
	Keyword   string // 模糊匹配参数键、名称与备注
	Status    enums.Status
	ValueType enums.ParamValueType
	Page      int
	PageSize  int
}

// SysParamRepo 系统参数仓储
type SysParamRepo interface {
	// Page 分页查询参数，按键倒序
	Page(ctx context.Context, filter *ParamFilter) ([]*entity.Param, int64, error)
	// FindByID 按主键查询参数
	FindByID(ctx context.Context, id int64) (*entity.Param, bool, error)
	// FindByKey 按键查询参数（含禁用项）
	FindByKey(ctx context.Context, key string) (*entity.Param, bool, error)
	// GetEnabledByKey 按键查询单条启用参数，供缓存不可用时的读路径降级
	GetEnabledByKey(ctx context.Context, key string) (*entity.Param, bool, error)
	// LoadEnabled 加载全部启用参数，供投影全量构建
	LoadEnabled(ctx context.Context) ([]*entity.Param, error)
	// CreateParam 创建参数；param_key 唯一冲突时返回 ErrDuplicatedKey
	CreateParam(ctx context.Context, param *entity.Param, operatorID int64) error
	// UpdateParam 更新参数值、名称与备注（键与值类型不可变）
	UpdateParam(ctx context.Context, param *entity.Param, operatorID int64) error
	// UpdateParamStatus 设置参数启停
	UpdateParamStatus(ctx context.Context, id int64, status enums.Status, operatorID int64) error
	// DeleteParam 物理删除参数
	DeleteParam(ctx context.Context, id int64) error
}

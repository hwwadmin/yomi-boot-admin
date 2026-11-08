// Package repo 系统参数仓储实现
package repo

import (
	"context"
	"strings"
	"yomi-admin-backend/domains/sysparam/internal/repo_contract"
	"yomi-admin-backend/domains/sysparam/pkg/shared/entity"
	"yomi-admin-backend/domains/sysparam/pkg/shared/enums"

	yomidb "github.com/hwwadmin/yomi-boot/component/common/db"
	"github.com/hwwadmin/yomi-boot/component/common/db/tx"
	"github.com/hwwadmin/yomi-boot/component/core/errors"
	sharedenums "github.com/hwwadmin/yomi-boot/component/core/shared/enums"
	"github.com/hwwadmin/yomi-boot/component/core/utils/ptrutil"
	"yomi-admin-backend/persist/orm"

	"gorm.io/gorm"
)

// SysParamRepo 系统参数仓储实现
type SysParamRepo struct {
	orm.BaseRepository
	dbType sharedenums.DbType
}

var _ repo_contract.SysParamRepo = (*SysParamRepo)(nil)

// NewSysParamRepo 创建系统参数仓储
func NewSysParamRepo(d *gorm.DB, dbType sharedenums.DbType) *SysParamRepo {
	return &SysParamRepo{
		BaseRepository: orm.BaseRepository{DB: d},
		dbType:         dbType,
	}
}

// Page 分页查询参数；按主键倒序，雪花 ID 单调递增，因此新参数排在最前
func (r *SysParamRepo) Page(ctx context.Context, filter *repo_contract.ParamFilter) ([]*entity.Param, int64, error) {
	if filter == nil {
		return nil, 0, errors.New("参数分页过滤条件不能为空")
	}

	query := tx.Exec(ctx, r.DB).Model(&sysParamModel{})
	query = applyParamFilter(query, filter)

	var rows []*sysParamModel
	total, err := r.CountPage(ctx, query, filter.Page, filter.PageSize, &rows, orm.OrderDesc("id"))
	if err != nil {
		return nil, 0, errors.Wrap(err, "分页查询系统参数失败")
	}

	items := make([]*entity.Param, 0, len(rows))
	for _, row := range rows {
		items = append(items, toParamEntity(row))
	}
	return items, total, nil
}

// FindByID 按主键查询参数
func (r *SysParamRepo) FindByID(ctx context.Context, id int64) (*entity.Param, bool, error) {
	var row sysParamModel
	found, err := r.FirstByID(ctx, id, &row)
	if err != nil {
		return nil, false, errors.Wrap(err, "按主键查询系统参数失败")
	}
	if !found {
		return nil, false, nil
	}
	return toParamEntity(&row), true, nil
}

// FindByKey 按键查询参数（含禁用项）
func (r *SysParamRepo) FindByKey(ctx context.Context, key string) (*entity.Param, bool, error) {
	var row sysParamModel
	found, err := r.FirstByField(ctx, "param_key", key, &row)
	if err != nil {
		return nil, false, errors.Wrap(err, "按键查询系统参数失败")
	}
	if !found {
		return nil, false, nil
	}
	return toParamEntity(&row), true, nil
}

// GetEnabledByKey 按键查询单条启用参数，是业务单键读路径的唯一查询
func (r *SysParamRepo) GetEnabledByKey(ctx context.Context, key string) (*entity.Param, bool, error) {
	var row sysParamModel
	err := tx.Exec(ctx, r.DB).
		Where("param_key = ? AND status = ?", key, enums.StatusEnabled).
		First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, errors.Wrap(err, "查询系统参数失败")
	}
	return toParamEntity(&row), true, nil
}

// FindEnabledByKeys 批量按键查询启用参数，供业务批量读路径使用；键列表为空时不查库直接返回
func (r *SysParamRepo) FindEnabledByKeys(ctx context.Context, keys []string) ([]*entity.Param, error) {
	if len(keys) == 0 {
		return nil, nil
	}

	var rows []*sysParamModel
	if err := tx.Exec(ctx, r.DB).
		Where("param_key IN ? AND status = ?", keys, enums.StatusEnabled).
		Find(&rows).Error; err != nil {
		return nil, errors.Wrap(err, "批量查询系统参数失败")
	}

	items := make([]*entity.Param, 0, len(rows))
	for _, row := range rows {
		items = append(items, toParamEntity(row))
	}
	return items, nil
}

// CreateParam 创建参数；param_key 唯一冲突时返回 ErrDuplicatedKey。
// 写入成功后把数据库生成的时间回填到实体，保证返回给调用方的实体与库内一致。
func (r *SysParamRepo) CreateParam(ctx context.Context, param *entity.Param, operatorID int64) error {
	row := toParamModel(param, operatorID)
	if row == nil {
		return errors.New("参数实体不能为空")
	}
	if err := r.Create(ctx, row); err != nil {
		if yomidb.IsDuplicateKey(err, r.dbType) {
			return errors.Sentinel(repo_contract.ErrDuplicatedKey)
		}
		return errors.Wrap(err, "创建系统参数失败")
	}
	param.CreateTime = row.CreateTime.UnixMilli()
	param.UpdateTime = row.UpdateTime.UnixMilli()
	return nil
}

// UpdateParam 更新参数值、名称与备注（键与值类型不可变）
func (r *SysParamRepo) UpdateParam(ctx context.Context, param *entity.Param, operatorID int64) error {
	updates := map[string]any{
		"param_value": param.Value,
		"name":        param.Name,
		"remark":      param.Remark,
		"update_by":   ptrutil.PtrFromInt64(operatorID),
	}
	if err := r.UpdateByID(ctx, param.ID, &sysParamModel{}, updates); err != nil {
		return errors.Wrap(err, "更新系统参数失败")
	}
	return nil
}

// UpdateParamStatus 设置参数启停
func (r *SysParamRepo) UpdateParamStatus(ctx context.Context, id int64, status enums.Status, operatorID int64) error {
	updates := map[string]any{
		"status":    status,
		"update_by": ptrutil.PtrFromInt64(operatorID),
	}
	if err := r.UpdateByID(ctx, id, &sysParamModel{}, updates); err != nil {
		return errors.Wrap(err, "更新系统参数状态失败")
	}
	return nil
}

// DeleteParam 物理删除参数
func (r *SysParamRepo) DeleteParam(ctx context.Context, id int64) error {
	if err := r.DeleteByID(ctx, id, &sysParamModel{}); err != nil {
		return errors.Wrap(err, "删除系统参数失败")
	}
	return nil
}

// applyParamFilter 按过滤条件拼装查询
func applyParamFilter(query *gorm.DB, filter *repo_contract.ParamFilter) *gorm.DB {
	if keyword := strings.TrimSpace(filter.Keyword); keyword != "" {
		like := orm.LikeFull(keyword)
		query = query.Where("param_key LIKE ? OR name LIKE ? OR remark LIKE ?", like, like, like)
	}
	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}
	if filter.ValueType != "" {
		query = query.Where("value_type = ?", filter.ValueType)
	}
	return query
}

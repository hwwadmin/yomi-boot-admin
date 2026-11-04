// Package repo HTTP 访问记录仓储实现
package repo

import (
	"context"
	"strings"
	"time"
	"yomi-admin-backend/domains/httprecord/internal/repo_contract"
	"yomi-admin-backend/domains/httprecord/pkg/shared/entity"
	"yomi-admin-backend/domains/httprecord/pkg/shared/enums"

	"github.com/hwwadmin/yomi-boot/component/common/db/orm"
	"github.com/hwwadmin/yomi-boot/component/common/db/tx"
	"github.com/hwwadmin/yomi-boot/component/core/errors"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// RecordRepo HTTP 访问记录仓储实现
type RecordRepo struct {
	orm.BaseRepository
}

var _ repo_contract.RecordRepo = (*RecordRepo)(nil)

// NewRecordRepo 创建访问记录仓储
func NewRecordRepo(d *gorm.DB) *RecordRepo {
	return &RecordRepo{BaseRepository: orm.BaseRepository{DB: d}}
}

// Create 写入访问记录；trace_id + request_at 重复时忽略，保证消费者重投幂等
func (r *RecordRepo) Create(ctx context.Context, record *entity.Record) error {
	row := toRecordModel(record)
	if row == nil {
		return errors.Sentinel(repo_contract.ErrInvalidRecord)
	}
	err := tx.Exec(ctx, r.DB).
		Clauses(clause.OnConflict{DoNothing: true}).
		Create(row).Error
	if err != nil {
		return errors.Wrap(err, "写入访问记录失败")
	}
	return nil
}

// FindByID 按主键查询访问记录
func (r *RecordRepo) FindByID(ctx context.Context, id int64) (*entity.Record, bool, error) {
	var row httpRecordModel
	found, err := r.FirstByID(ctx, id, &row)
	if err != nil {
		return nil, false, errors.Wrap(err, "按主键查询访问记录失败")
	}
	if !found {
		return nil, false, nil
	}
	return toRecordEntity(&row), true, nil
}

// Page 分页查询访问记录；按请求时间与主键倒序
func (r *RecordRepo) Page(ctx context.Context, filter *repo_contract.RecordFilter) ([]*entity.Record, int64, error) {
	if filter == nil {
		return nil, 0, errors.Sentinel(repo_contract.ErrInvalidRecord)
	}

	query := tx.Exec(ctx, r.DB).Model(&httpRecordModel{})
	query = applyRecordFilter(query, filter)

	var rows []*httpRecordModel
	total, err := r.CountPage(ctx, query, filter.Page, filter.PageSize, &rows,
		orm.OrderDesc("request_at"), orm.OrderDesc("id"))
	if err != nil {
		return nil, 0, errors.Wrap(err, "分页查询访问记录失败")
	}

	items := make([]*entity.Record, 0, len(rows))
	for _, row := range rows {
		items = append(items, toRecordEntity(row))
	}
	return items, total, nil
}

// applyRecordFilter 按过滤条件拼装查询
func applyRecordFilter(query *gorm.DB, filter *repo_contract.RecordFilter) *gorm.DB {
	if keyword := strings.TrimSpace(filter.Keyword); keyword != "" {
		like := orm.LikeFull(keyword)
		query = query.Where("path LIKE ? OR route LIKE ? OR trace_id LIKE ? OR client_ip LIKE ?",
			like, like, like, like)
	}
	if method := strings.TrimSpace(filter.Method); method != "" {
		query = query.Where("method = ?", method)
	}
	if min, max, ok := enums.StatusCodeRange(enums.StatusClass(filter.StatusClass)); ok {
		query = query.Where("status_code >= ? AND status_code < ?", min, max)
	}
	if filter.UserID > 0 {
		query = query.Where("user_id = ?", filter.UserID)
	}
	if filter.StartTime > 0 {
		query = query.Where("request_at >= ?", time.UnixMilli(filter.StartTime).UTC())
	}
	if filter.EndTime > 0 {
		query = query.Where("request_at <= ?", time.UnixMilli(filter.EndTime).UTC())
	}
	return query
}

// Package repo 操作审计日志仓储实现
package repo

import (
	"context"
	"strings"
	"time"
	"yomi-admin-backend/domains/auditlog/internal/repo_contract"
	"yomi-admin-backend/domains/auditlog/pkg/shared/entity"

	"gitee.com/hq_project/yomi-boot/component/common/db/orm"
	"gitee.com/hq_project/yomi-boot/component/common/db/tx"
	"gitee.com/hq_project/yomi-boot/component/core/errors"

	"gorm.io/gorm"
)

// AuditLogRepo 操作审计日志仓储实现
type AuditLogRepo struct {
	orm.BaseRepository
}

var _ repo_contract.AuditLogRepo = (*AuditLogRepo)(nil)

// NewAuditLogRepo 创建审计日志仓储
func NewAuditLogRepo(d *gorm.DB) *AuditLogRepo {
	return &AuditLogRepo{BaseRepository: orm.BaseRepository{DB: d}}
}

// Create 写入一条审计记录。
//
// 审计写入必须绕开业务事务：直接使用原始 DB 句柄，不走 tx.Exec，也不走
// BaseRepository.Create（内部同样是 tx.Exec）。即使调用方带着事务上下文，
// 这条 insert 也不会加入业务事务。
func (r *AuditLogRepo) Create(ctx context.Context, log *entity.Log) error {
	row := toLogModel(log)
	if row == nil {
		return errors.New("审计日志实体不能为空")
	}
	if err := r.DB.WithContext(ctx).Create(row).Error; err != nil {
		return errors.Wrap(err, "写入审计日志失败")
	}
	return nil
}

// Page 分页查询审计日志；按事件时间与主键倒序
func (r *AuditLogRepo) Page(ctx context.Context, filter *repo_contract.Filter) ([]*entity.Log, int64, error) {
	if filter == nil {
		return nil, 0, errors.New("审计日志分页过滤条件不能为空")
	}

	query := applyLogFilter(tx.Exec(ctx, r.DB).Model(&auditLogModel{}), filter)

	var rows []*auditLogModel
	total, err := r.CountPage(ctx, query, filter.Page, filter.PageSize, &rows,
		orm.OrderDesc("create_time"), orm.OrderDesc("id"))
	if err != nil {
		return nil, 0, errors.Wrap(err, "分页查询审计日志失败")
	}

	items := make([]*entity.Log, 0, len(rows))
	for _, row := range rows {
		items = append(items, toLogEntity(row))
	}
	return items, total, nil
}

// applyLogFilter 按过滤条件拼装查询
func applyLogFilter(query *gorm.DB, filter *repo_contract.Filter) *gorm.DB {
	// 关键字只匹配落库快照（操作者名、来源 IP）：目标用户名不落库，由查询侧按 ID 实时回填，无法参与 SQL 过滤
	if keyword := strings.TrimSpace(filter.Keyword); keyword != "" {
		like := orm.LikeFull(keyword)
		query = query.Where("operator_name LIKE ? OR client_ip LIKE ?", like, like)
	}
	if filter.EventCode != "" {
		query = query.Where("event_code = ?", filter.EventCode)
	}
	if filter.Phase != "" {
		query = query.Where("phase = ?", filter.Phase)
	}
	if filter.Result != "" {
		query = query.Where("result = ?", filter.Result)
	}
	if filter.AuditID > 0 {
		query = query.Where("audit_id = ?", filter.AuditID)
	}
	if filter.StartTime > 0 {
		query = query.Where("create_time >= ?", time.UnixMilli(filter.StartTime).UTC())
	}
	if filter.EndTime > 0 {
		query = query.Where("create_time <= ?", time.UnixMilli(filter.EndTime).UTC())
	}
	return query
}

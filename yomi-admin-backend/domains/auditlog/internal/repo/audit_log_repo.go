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

// Page 分页查询审计日志；按事件时间与主键倒序。
//
// 深翻页走"索引子查询 + 主键回表"：子查询只取本页的主键，排序与跳过都发生在
// (create_time, id) 索引里，不取整行；外层再按主键取这几十行。直接
// LIMIT/OFFSET 大偏移量时，被跳过的每一行都要回表取整行再丢掉，页数一深就慢。
// 子查询列必须改名（id AS page_id），否则外层 ORDER BY 里的 id 会和派生表撞名。
func (r *AuditLogRepo) Page(ctx context.Context, filter *repo_contract.Filter) ([]*entity.Log, int64, error) {
	if filter == nil {
		return nil, 0, errors.New("审计日志分页过滤条件不能为空")
	}

	orders := []orm.OrderClause{orm.OrderDesc("create_time"), orm.OrderDesc("id")}
	base := applyLogFilter(r.scopedQuery(ctx), filter)

	var total int64
	if err := base.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		return nil, 0, errors.Wrap(err, "统计审计日志失败")
	}
	if total == 0 {
		return []*entity.Log{}, 0, nil
	}

	pageIDs, err := orm.ApplyPage(base.Session(&gorm.Session{}).Select("id AS page_id"),
		filter.Page, filter.PageSize, orders...)
	if err != nil {
		return nil, 0, errors.Wrap(err, "审计日志分页参数非法")
	}

	query, err := orm.ApplyPage(
		r.scopedQuery(ctx).
			Select("t_audit_log.*").
			Joins("JOIN (?) AS page_ids ON page_ids.page_id = t_audit_log.id", pageIDs),
		1, filter.PageSize, orders...)
	if err != nil {
		return nil, 0, errors.Wrap(err, "审计日志分页参数非法")
	}

	var rows []*auditLogModel
	if err := query.Find(&rows).Error; err != nil {
		return nil, 0, errors.Wrap(err, "分页查询审计日志失败")
	}

	items := make([]*entity.Log, 0, len(rows))
	for _, row := range rows {
		items = append(items, toLogEntity(row))
	}
	return items, total, nil
}

// scopedQuery 返回审计日志的基础查询。
//
// 审计只追加、不软删（见建表脚本），这里显式 Unscoped：若带上 GORM 自动加的
// delete_time IS NULL，子查询就必须回表逐行判断，索引覆盖失效，深翻页优化白做。
func (r *AuditLogRepo) scopedQuery(ctx context.Context) *gorm.DB {
	return tx.Exec(ctx, r.DB).Unscoped().Model(&auditLogModel{})
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

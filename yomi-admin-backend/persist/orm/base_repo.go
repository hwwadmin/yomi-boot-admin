package orm

import (
	"context"
	"reflect"

	"github.com/hwwadmin/yomi-boot/component/common/db/tx"
	"github.com/hwwadmin/yomi-boot/component/core/errors"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// BaseRepository 是各域仓储共用的读写辅助。
// DB 必须是数据源根句柄，事务解析依赖该句柄。
type BaseRepository struct {
	DB *gorm.DB
}

// idChunkSize 是批量 ID 查询的单片上限，避免触发数据库占位符上限。
const idChunkSize = 500

func uniqueIDs(ids []int64) []int64 {
	seen := make(map[int64]struct{}, len(ids))
	out := make([]int64, 0, len(ids))
	for _, id := range ids {
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	return out
}

// Create 持久化一条记录。
func (r *BaseRepository) Create(ctx context.Context, model any) error {
	if err := tx.Exec(ctx, r.DB).Create(model).Error; err != nil {
		return errors.Wrap(err, "创建记录失败")
	}
	return nil
}

// CreateInBatches 批量持久化记录；batchSize 必须大于 0。
func (r *BaseRepository) CreateInBatches(ctx context.Context, models any, batchSize int) error {
	if batchSize <= 0 {
		return errors.Newf("batchSize 必须大于 0: %d", batchSize)
	}
	if err := tx.Exec(ctx, r.DB).CreateInBatches(models, batchSize).Error; err != nil {
		return errors.Wrap(err, "批量创建记录失败")
	}
	return nil
}

// FirstByID 根据主键 id 查询单条记录。
func (r *BaseRepository) FirstByID(ctx context.Context, id int64, dest any) (bool, error) {
	err := tx.Exec(ctx, r.DB).Where("id = ?", id).First(dest).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, nil
	}
	if err != nil {
		return false, errors.Wrap(err, "根据主键查询失败")
	}
	return true, nil
}

// FirstByField 根据指定字段查询单条记录。
func (r *BaseRepository) FirstByField(ctx context.Context, field string, value any, dest any) (bool, error) {
	if err := ValidateSQLIdentifier(field); err != nil {
		return false, errors.Wrap(err, "非法字段名")
	}
	err := tx.Exec(ctx, r.DB).Where(clause.Eq{Column: clause.Column{Name: field}, Value: value}).First(dest).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, nil
	}
	if err != nil {
		return false, errors.Wrap(err, "根据字段查询失败")
	}
	return true, nil
}

// FindByIDs 根据主键列表批量查询，dest 须为非空切片指针；超过 idChunkSize 自动分批并合并结果。
func (r *BaseRepository) FindByIDs(ctx context.Context, ids []int64, model any, dest any) error {
	if len(ids) == 0 {
		return nil
	}
	v := reflect.ValueOf(dest)
	if v.Kind() != reflect.Pointer || v.IsNil() || v.Elem().Kind() != reflect.Slice {
		return errors.New("FindByIDs dest 须为非空切片指针")
	}
	if len(ids) > idChunkSize {
		ids = uniqueIDs(ids)
	}

	sliceType := v.Elem().Type()
	combined := reflect.MakeSlice(sliceType, 0, len(ids))
	for start := 0; start < len(ids); start += idChunkSize {
		end := start + idChunkSize
		if end > len(ids) {
			end = len(ids)
		}
		chunk := reflect.New(sliceType)
		if err := tx.Exec(ctx, r.DB).Model(model).Where("id IN ?", ids[start:end]).Find(chunk.Interface()).Error; err != nil {
			return errors.Wrap(err, "根据主键列表查询失败")
		}
		combined = reflect.AppendSlice(combined, chunk.Elem())
	}
	v.Elem().Set(combined)
	return nil
}

// UpdateByID 根据主键更新记录。
func (r *BaseRepository) UpdateByID(ctx context.Context, id int64, model any, updates map[string]any) error {
	if err := tx.Exec(ctx, r.DB).Model(model).Where("id = ?", id).Updates(updates).Error; err != nil {
		return errors.Wrap(err, "根据主键更新记录失败")
	}
	return nil
}

// DeleteByID 根据主键硬删除记录（跳过 GORM 软删除）。
func (r *BaseRepository) DeleteByID(ctx context.Context, id int64, model any) error {
	if err := tx.Exec(ctx, r.DB).Unscoped().Where("id = ?", id).Delete(model).Error; err != nil {
		return errors.Wrap(err, "根据主键删除记录失败")
	}
	return nil
}

// CountPage 先统计再分页查询。
func (r *BaseRepository) CountPage(ctx context.Context, query *gorm.DB, page, pageSize int, dest any, orders ...OrderClause) (int64, error) {
	// 调用方传入的 query 通常 clone=0，Count 会把 SQL 写入 Statement，
	// 后续 Find 会复用 "SELECT count(*)"。先取 Session 让两次查询各自克隆干净语句。
	base := query.Session(&gorm.Session{})

	var total int64
	if err := base.Count(&total).Error; err != nil {
		return 0, errors.Wrap(err, "分页统计失败")
	}
	if total == 0 {
		if err := resetPageDest(dest); err != nil {
			return 0, errors.Wrap(err, "清空分页结果失败")
		}
		return 0, nil
	}
	paged, err := ApplyPage(base, page, pageSize, orders...)
	if err != nil {
		return 0, errors.Wrap(err, "分页排序参数非法")
	}
	if err := paged.Find(dest).Error; err != nil {
		return 0, errors.Wrap(err, "分页查询失败")
	}
	return total, nil
}

func resetPageDest(dest any) error {
	v := reflect.ValueOf(dest)
	if v.Kind() != reflect.Pointer || v.IsNil() {
		return errors.New("分页 dest 须为非空指针")
	}
	elem := v.Elem()
	if !elem.CanSet() {
		return errors.New("分页 dest 不可写")
	}
	if elem.Kind() == reflect.Slice {
		elem.Set(reflect.MakeSlice(elem.Type(), 0, 0))
		return nil
	}
	elem.Set(reflect.Zero(elem.Type()))
	return nil
}

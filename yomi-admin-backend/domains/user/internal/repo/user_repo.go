// Package repo 用户域仓储实现
package repo

import (
	"context"
	"strings"
	"time"
	"yomi-admin-backend/domains/user/internal/model"
	"yomi-admin-backend/domains/user/internal/repo_contract"
	"yomi-admin-backend/domains/user/pkg/shared/entity"

	userenums "yomi-admin-backend/domains/user/pkg/shared/enums"

	yomidb "gitee.com/hq_project/yomi-boot/component/common/db"
	"gitee.com/hq_project/yomi-boot/component/common/db/orm"
	"gitee.com/hq_project/yomi-boot/component/common/db/tx"
	"gitee.com/hq_project/yomi-boot/component/core/errors"
	sharedenums "gitee.com/hq_project/yomi-boot/component/core/shared/enums"
	"gitee.com/hq_project/yomi-boot/component/core/utils/ptrutil"

	"gorm.io/gorm"
)

// UserRepo 用户仓储实现
type UserRepo struct {
	orm.BaseRepository
	dbType sharedenums.DbType
}

var _ repo_contract.UserRepo = (*UserRepo)(nil)

// NewUserRepo 创建用户仓储
func NewUserRepo(d *gorm.DB, dbType sharedenums.DbType) *UserRepo {
	return &UserRepo{
		BaseRepository: orm.BaseRepository{DB: d},
		dbType:         dbType,
	}
}

// FindByID 按主键查询
func (r *UserRepo) FindByID(ctx context.Context, id int64) (*entity.User, bool, error) {
	var row model.User
	found, err := r.FirstByID(ctx, id, &row)
	if err != nil {
		return nil, false, errors.Wrap(err, "按主键查询用户失败")
	}
	if !found {
		return nil, false, nil
	}
	return toEntity(&row), true, nil
}

// FindCredentialByUsername 按登录名查询验密凭证
func (r *UserRepo) FindCredentialByUsername(ctx context.Context, username string) (*repo_contract.UserCredential, bool, error) {
	var row model.User
	found, err := r.FirstByField(ctx, "username", username, &row)
	if err != nil {
		return nil, false, errors.Wrap(err, "按登录名查询用户失败")
	}
	if !found {
		return nil, false, nil
	}
	return &repo_contract.UserCredential{
		User:         toEntity(&row),
		PasswordHash: row.PasswordHash,
	}, true, nil
}

// ExistsByUsername 判断用户名是否存在（包含已软删除记录）
func (r *UserRepo) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	var count int64
	err := tx.Exec(ctx, r.DB).
		Unscoped().
		Model(&model.User{}).
		Where("username = ?", username).
		Count(&count).Error
	if err != nil {
		return false, errors.Wrap(err, "检查用户名是否存在失败")
	}
	return count > 0, nil
}

// Page 分页查询用户
func (r *UserRepo) Page(ctx context.Context, keyword string, page, pageSize int) ([]*entity.User, int64, error) {
	query := tx.Exec(ctx, r.DB).Model(&model.User{})
	if keyword = strings.TrimSpace(keyword); keyword != "" {
		like := orm.LikeFull(keyword)
		query = query.Where("username LIKE ? OR nickname LIKE ?", like, like)
	}

	var rows []*model.User
	total, err := r.CountPage(ctx, query, page, pageSize, &rows, orm.OrderDesc("id"))
	if err != nil {
		return nil, 0, errors.Wrap(err, "分页查询用户失败")
	}

	items := make([]*entity.User, 0, len(rows))
	for _, row := range rows {
		items = append(items, toEntity(row))
	}
	return items, total, nil
}

// Create 创建用户并回填自增 ID
func (r *UserRepo) Create(ctx context.Context, user *entity.User, passwordHash string, operatorID int64) error {
	row := &model.User{
		Username:     user.Username,
		PasswordHash: passwordHash,
		Nickname:     user.Nickname,
		Email:        user.Email,
		Phone:        user.Phone,
		Status:       int8(user.Status),
	}
	row.CreateBy = ptrutil.PtrFromInt64(operatorID)
	row.UpdateBy = ptrutil.PtrFromInt64(operatorID)

	if err := r.BaseRepository.Create(ctx, row); err != nil {
		if yomidb.IsDuplicateKey(err, r.dbType) {
			return errors.Sentinel(repo_contract.ErrDuplicatedKey)
		}
		return err
	}

	user.ID = row.ID
	user.CreatedAt = row.CreateTime
	return nil
}

// Update 更新用户资料与状态
func (r *UserRepo) Update(ctx context.Context, user *entity.User, operatorID int64) error {
	updates := map[string]any{
		"nickname":  user.Nickname,
		"email":     user.Email,
		"phone":     user.Phone,
		"status":    int8(user.Status),
		"update_by": ptrutil.PtrFromInt64(operatorID),
	}
	if err := r.UpdateByID(ctx, user.ID, &model.User{}, updates); err != nil {
		return errors.Wrap(err, "更新用户失败")
	}
	return nil
}

// SoftDelete 软删除用户
func (r *UserRepo) SoftDelete(ctx context.Context, id int64, operatorID int64) error {
	updates := map[string]any{
		"delete_time": time.Now(),
		"delete_by":   ptrutil.PtrFromInt64(operatorID),
	}
	if err := r.UpdateByID(ctx, id, &model.User{}, updates); err != nil {
		return errors.Wrap(err, "删除用户失败")
	}
	return nil
}

// toEntity 将持久化模型转换为领域实体
func toEntity(row *model.User) *entity.User {
	return &entity.User{
		ID:        row.ID,
		Username:  row.Username,
		Nickname:  row.Nickname,
		Email:     row.Email,
		Phone:     row.Phone,
		Status:    userenums.UserStatus(row.Status),
		CreatedAt: row.CreateTime,
	}
}

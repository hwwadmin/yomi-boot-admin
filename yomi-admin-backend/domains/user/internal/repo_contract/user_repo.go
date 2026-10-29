// Package repo_contract 用户域仓储契约
package repo_contract

import (
	"context"
	"yomi-admin-backend/domains/user/pkg/shared/entity"
)

// UserCredential 登录验密凭证（含密码哈希，不进入对外实体）
type UserCredential struct {
	User         *entity.User
	PasswordHash string
}

// UserRepo 用户仓储
type UserRepo interface {
	// FindByID 按主键查询
	FindByID(ctx context.Context, id int64) (*entity.User, bool, error)
	// FindCredentialByUsername 按登录名查询验密凭证
	FindCredentialByUsername(ctx context.Context, username string) (*UserCredential, bool, error)
	// FindCredentialByID 按主键查询验密凭证
	FindCredentialByID(ctx context.Context, id int64) (*UserCredential, bool, error)
	// ExistsByUsername 判断用户名是否存在（包含已软删除记录）
	ExistsByUsername(ctx context.Context, username string) (bool, error)
	// Page 分页查询用户
	Page(ctx context.Context, keyword string, page, pageSize int) ([]*entity.User, int64, error)
	// Create 创建用户（主键由领域服务生成雪花 ID）
	Create(ctx context.Context, user *entity.User, passwordHash string, operatorID int64) error
	// Update 更新用户资料与状态
	Update(ctx context.Context, user *entity.User, operatorID int64) error
	// UpdatePassword 只更新密码哈希，不动其他字段
	UpdatePassword(ctx context.Context, id int64, passwordHash string, operatorID int64) error
	// SoftDelete 软删除用户
	SoftDelete(ctx context.Context, id int64, operatorID int64) error
}

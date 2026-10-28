// Package service_contract 用户域服务契约
package service_contract

import (
	"context"
	"yomi-admin-backend/domains/user/pkg/shared/entity"
	"yomi-admin-backend/domains/user/pkg/shared/input"
	"yomi-admin-backend/domains/user/pkg/shared/output"
)

// UserService 用户领域服务
type UserService interface {
	// Authenticate 校验账号密码，返回用户实体（不含 Token）
	Authenticate(ctx context.Context, in *input.AuthenticateInput) (*output.AuthenticateOutput, error)
	// GetByID 按用户 ID 查询
	GetByID(ctx context.Context, id int64) (*entity.User, error)
	// Page 分页查询用户
	Page(ctx context.Context, in *input.PageUsersInput) (*output.PageUsersOutput, error)
	// Create 创建用户
	Create(ctx context.Context, in *input.CreateUserInput) (*entity.User, error)
	// Update 更新用户资料与状态；调用方负责最后一名超管与自操作校验
	Update(ctx context.Context, in *input.UpdateUserInput) error
	// Delete 软删除用户；调用方负责最后一名超管与自操作校验
	Delete(ctx context.Context, in *input.DeleteUserInput) error
}

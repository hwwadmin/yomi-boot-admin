// Package service_contract RBAC 域服务契约
package service_contract

import (
	"context"
	"yomi-admin-backend/domains/rbac/pkg/shared/entity"
	"yomi-admin-backend/domains/rbac/pkg/shared/input"
	"yomi-admin-backend/domains/rbac/pkg/shared/output"
)

// RbacAdminService 角色与权限管理
type RbacAdminService interface {
	// PageRoles 分页查询角色
	PageRoles(ctx context.Context, in *input.PageRolesInput) (*output.PageRolesOutput, error)
	// GetRole 查询角色详情（含已绑权限 ID）
	GetRole(ctx context.Context, id int64) (*entity.RoleDetail, error)
	// CreateRole 创建角色
	CreateRole(ctx context.Context, in *input.CreateRoleInput) (*entity.Role, error)
	// UpdateRole 更新角色资料与状态；内置超管不可禁用
	UpdateRole(ctx context.Context, in *input.UpdateRoleInput) error
	// BindRolePermissions 覆盖角色权限
	BindRolePermissions(ctx context.Context, in *input.BindRolePermissionsInput) error

	// PagePermissions 分页查询权限目录
	PagePermissions(ctx context.Context, in *input.PagePermissionsInput) (*output.PagePermissionsOutput, error)
	// SetPermissionStatus 设置权限启停
	SetPermissionStatus(ctx context.Context, in *input.SetPermissionStatusInput) error

	// ListUserRoles 查询用户角色绑定视图
	ListUserRoles(ctx context.Context, userID int64) (*output.UserRolesOutput, error)
	// BindUserRoles 覆盖用户角色
	BindUserRoles(ctx context.Context, in *input.BindUserRolesInput) error
	// ListRolesByUserIDs 按用户批量查询已绑角色
	ListRolesByUserIDs(ctx context.Context, userIDs []int64) (map[int64][]*entity.Role, error)
	// ListUserIDsWithEnabledSuperAdmin 列出拥有启用态超管角色的用户 ID
	ListUserIDsWithEnabledSuperAdmin(ctx context.Context) ([]int64, error)
}

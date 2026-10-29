// Package service_contract RBAC 域服务契约
package service_contract

import (
	"context"
	"yomi-admin-backend/domains/rbac/pkg/shared/entity"
	"yomi-admin-backend/domains/rbac/pkg/shared/input"
	"yomi-admin-backend/domains/rbac/pkg/shared/output"
)

// RbacAdminService 角色、权限组与权限管理
type RbacAdminService interface {
	// PageRoles 分页查询角色
	PageRoles(ctx context.Context, in *input.PageRolesInput) (*output.PageRolesOutput, error)
	// GetRole 查询角色详情（含已绑权限组与聚合权限 ID）
	GetRole(ctx context.Context, id int64) (*entity.RoleDetail, error)
	// CreateRole 创建角色
	CreateRole(ctx context.Context, in *input.CreateRoleInput) (*entity.Role, error)
	// UpdateRole 更新角色资料；编码不可变
	UpdateRole(ctx context.Context, in *input.UpdateRoleInput) error
	// DeleteRole 硬删除角色；仍被用户绑定时拒绝，并清理角色权限组关系
	DeleteRole(ctx context.Context, in *input.DeleteRoleInput) error
	// SetRoleStatus 设置角色启停；内置超管不可禁用
	SetRoleStatus(ctx context.Context, in *input.SetRoleStatusInput) error
	// BindRoleGroups 覆盖角色权限组
	BindRoleGroups(ctx context.Context, in *input.BindRoleGroupsInput) error

	// PagePermissionGroups 分页查询权限组
	PagePermissionGroups(ctx context.Context, in *input.PagePermissionGroupsInput) (*output.PagePermissionGroupsOutput, error)
	// GetPermissionGroup 查询权限组详情（含已绑权限 ID）
	GetPermissionGroup(ctx context.Context, id int64) (*entity.PermissionGroupDetail, error)
	// CreatePermissionGroup 创建权限组
	CreatePermissionGroup(ctx context.Context, in *input.CreatePermissionGroupInput) (*entity.PermissionGroup, error)
	// UpdatePermissionGroup 更新权限组资料；编码不可变
	UpdatePermissionGroup(ctx context.Context, in *input.UpdatePermissionGroupInput) error
	// DeletePermissionGroup 硬删除权限组；仍被角色绑定时拒绝，并清理权限组权限关系
	DeletePermissionGroup(ctx context.Context, in *input.DeletePermissionGroupInput) error
	// SetPermissionGroupStatus 设置权限组启停
	SetPermissionGroupStatus(ctx context.Context, in *input.SetPermissionGroupStatusInput) error
	// BindGroupPermissions 覆盖权限组权限
	BindGroupPermissions(ctx context.Context, in *input.BindGroupPermissionsInput) error

	// PagePermissions 分页查询权限目录
	PagePermissions(ctx context.Context, in *input.PagePermissionsInput) (*output.PagePermissionsOutput, error)
	// CreatePermission 创建权限
	CreatePermission(ctx context.Context, in *input.CreatePermissionInput) (*entity.Permission, error)
	// UpdatePermission 更新权限资料；编码与类型不可变
	UpdatePermission(ctx context.Context, in *input.UpdatePermissionInput) error
	// DeletePermission 硬删除权限，并清理权限组权限关系
	DeletePermission(ctx context.Context, in *input.DeletePermissionInput) error
	// SetPermissionStatus 设置权限启停
	SetPermissionStatus(ctx context.Context, in *input.SetPermissionStatusInput) error

	// ListUserRoles 查询用户角色绑定视图
	ListUserRoles(ctx context.Context, userID int64) (*output.UserRolesOutput, error)
	// BindUserRoles 覆盖用户角色；调用方负责超管授予/回收、自解绑与最后一名超管校验
	BindUserRoles(ctx context.Context, in *input.BindUserRolesInput) error
	// DeleteUserRoles 硬删除用户全部角色绑定
	DeleteUserRoles(ctx context.Context, userID int64) error
	// ListRolesByUserIDs 按用户批量查询已绑角色
	ListRolesByUserIDs(ctx context.Context, userIDs []int64) (map[int64][]*entity.Role, error)
	// ListRolesByIDs 按主键批量查询角色
	ListRolesByIDs(ctx context.Context, ids []int64) ([]*entity.Role, error)
	// ListUserIDsWithEnabledSuperAdmin 列出拥有启用态超管角色的用户 ID
	ListUserIDsWithEnabledSuperAdmin(ctx context.Context) ([]int64, error)
}

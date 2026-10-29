// Package input RBAC 域服务入参
package input

import "yomi-admin-backend/domains/rbac/pkg/shared/enums"

// PageRolesInput 角色分页查询入参
type PageRolesInput struct {
	Keyword  string
	Status   enums.Status
	Page     int
	PageSize int
}

// CreateRoleInput 创建角色入参
type CreateRoleInput struct {
	Code       string
	Name       string
	Remark     string
	OperatorID int64
}

// UpdateRoleInput 更新角色资料入参；编码不可变，启停走独立入参
type UpdateRoleInput struct {
	ID         int64
	Name       string
	Remark     string
	OperatorID int64
}

// DeleteRoleInput 删除角色入参
type DeleteRoleInput struct {
	ID         int64
	OperatorID int64
}

// SetRoleStatusInput 设置角色启停入参
type SetRoleStatusInput struct {
	ID         int64
	Status     enums.Status
	OperatorID int64
}

// BindRoleGroupsInput 覆盖角色权限组入参
type BindRoleGroupsInput struct {
	RoleID     int64
	GroupIDs   []int64
	OperatorID int64
}

// PagePermissionGroupsInput 权限组分页查询入参
type PagePermissionGroupsInput struct {
	Keyword  string
	Status   enums.Status
	Page     int
	PageSize int
}

// CreatePermissionGroupInput 创建权限组入参
type CreatePermissionGroupInput struct {
	Code       string
	Name       string
	Remark     string
	OperatorID int64
}

// UpdatePermissionGroupInput 更新权限组资料入参；编码不可变，启停走独立入参
type UpdatePermissionGroupInput struct {
	ID         int64
	Name       string
	Remark     string
	OperatorID int64
}

// DeletePermissionGroupInput 删除权限组入参
type DeletePermissionGroupInput struct {
	ID         int64
	OperatorID int64
}

// SetPermissionGroupStatusInput 设置权限组启停入参
type SetPermissionGroupStatusInput struct {
	ID         int64
	Status     enums.Status
	OperatorID int64
}

// BindGroupPermissionsInput 覆盖权限组权限入参
type BindGroupPermissionsInput struct {
	GroupID       int64
	PermissionIDs []int64
	OperatorID    int64
}

// PagePermissionsInput 权限分页查询入参
type PagePermissionsInput struct {
	Keyword  string
	Type     enums.PermissionType
	Status   enums.Status
	Page     int
	PageSize int
}

// CreatePermissionInput 创建权限入参
type CreatePermissionInput struct {
	Code       string
	Name       string
	Type       enums.PermissionType
	UIPath     string
	HTTPMethod string
	HTTPPath   string
	Remark     string
	OperatorID int64
}

// UpdatePermissionInput 更新权限入参；编码与类型不可变，启停走独立入参
type UpdatePermissionInput struct {
	ID         int64
	Name       string
	UIPath     string
	HTTPMethod string
	HTTPPath   string
	Remark     string
	OperatorID int64
}

// DeletePermissionInput 删除权限入参
type DeletePermissionInput struct {
	ID         int64
	OperatorID int64
}

// SetPermissionStatusInput 设置权限启停入参
type SetPermissionStatusInput struct {
	ID         int64
	Status     enums.Status
	OperatorID int64
}

// BindUserRolesInput 覆盖用户角色入参
type BindUserRolesInput struct {
	UserID     int64
	RoleIDs    []int64
	OperatorID int64
}

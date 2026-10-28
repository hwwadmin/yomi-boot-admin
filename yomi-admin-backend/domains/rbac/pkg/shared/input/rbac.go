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

// UpdateRoleInput 更新角色入参；编码不可变
type UpdateRoleInput struct {
	ID         int64
	Name       string
	Remark     string
	Status     enums.Status
	OperatorID int64
}

// BindRolePermissionsInput 覆盖角色权限入参
type BindRolePermissionsInput struct {
	RoleID        int64
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

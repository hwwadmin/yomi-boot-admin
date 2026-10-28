// Package output RBAC 域服务出参
package output

import "yomi-admin-backend/domains/rbac/pkg/shared/entity"

// PageRolesOutput 角色分页出参
type PageRolesOutput struct {
	Items []*entity.Role
	Total int64
}

// PagePermissionsOutput 权限分页出参
type PagePermissionsOutput struct {
	Items []*entity.Permission
	Total int64
}

// UserRolesOutput 用户角色绑定视图
type UserRolesOutput struct {
	Roles   []*entity.Role // 全部角色，供分配界面展示
	RoleIDs []int64        // 已绑定角色 ID
}

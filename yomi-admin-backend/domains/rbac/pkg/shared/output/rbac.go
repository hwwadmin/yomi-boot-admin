// Package output RBAC 域服务出参
package output

import "yomi-admin-backend/domains/rbac/pkg/shared/entity"

// PageRolesOutput 角色分页出参
type PageRolesOutput struct {
	Items []*entity.Role
	Total int64
}

// PagePermissionGroupsOutput 权限组分页出参
type PagePermissionGroupsOutput struct {
	Items []*entity.PermissionGroup
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

// CacheRefreshOutput 鉴权投影刷新结果
type CacheRefreshOutput struct {
	Version              string // 投影版本号
	BuiltAt              int64  // 投影构建时间（Unix 毫秒）
	RoleCount            int    // 角色数
	PermissionCount      int    // 权限数
	PermissionGroupCount int    // 权限组数
	RelationCount        int    // 三类关系条目总数
	DurationMS           int64  // 刷新耗时（毫秒）
}

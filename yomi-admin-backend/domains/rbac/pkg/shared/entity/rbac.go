// Package entity RBAC 域实体
package entity

import "yomi-admin-backend/domains/rbac/pkg/shared/enums"

// Role 表示角色
type Role struct {
	ID         int64
	Code       string
	Name       string
	Builtin    bool
	Status     enums.Status
	Remark     string
	CreateTime int64 // Unix 毫秒时间戳
	UpdateTime int64 // Unix 毫秒时间戳
}

// Permission 表示权限（菜单或接口）
type Permission struct {
	ID         int64
	Code       string
	Name       string
	Type       enums.PermissionType
	UIPath     string // 菜单路径，Type=PERMISSION_TYPE_MENU 时有值
	HTTPMethod string // HTTP 方法，Type=PERMISSION_TYPE_API 时有值
	HTTPPath   string // gin 路由模板，Type=PERMISSION_TYPE_API 时有值
	Status     enums.Status
	Remark     string
	CreateTime int64 // Unix 毫秒时间戳
	UpdateTime int64 // Unix 毫秒时间戳
}

// PermissionGroup 表示权限组，是权限的复用单元
type PermissionGroup struct {
	ID         int64
	Code       string
	Name       string
	Status     enums.Status
	Remark     string
	CreateTime int64 // Unix 毫秒时间戳
	UpdateTime int64 // Unix 毫秒时间戳
}

// RoleDetail 表示角色详情；PermissionIDs 是所绑权限组的去重聚合（含停用项），
// EffectivePermissionIDs 只统计启用权限组下的启用权限，供界面标注未生效项。
type RoleDetail struct {
	Role                   *Role
	GroupIDs               []int64
	PermissionIDs          []int64
	EffectivePermissionIDs []int64
}

// PermissionGroupDetail 表示权限组详情（含已绑权限 ID）
type PermissionGroupDetail struct {
	Group         *PermissionGroup
	PermissionIDs []int64
}

// UserRoleRel 表示用户-角色关系
type UserRoleRel struct {
	ID     int64
	UserID int64
	RoleID int64
}

// RoleGroupRel 表示角色-权限组关系
type RoleGroupRel struct {
	ID      int64
	RoleID  int64
	GroupID int64
}

// GroupPermissionRel 表示权限组-权限关系
type GroupPermissionRel struct {
	ID           int64
	GroupID      int64
	PermissionID int64
}

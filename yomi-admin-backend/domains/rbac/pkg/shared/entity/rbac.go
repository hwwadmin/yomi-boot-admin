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

// RoleDetail 表示角色详情（含已绑权限 ID）
type RoleDetail struct {
	Role          *Role
	PermissionIDs []int64
}

// UserRoleRel 表示用户-角色关系
type UserRoleRel struct {
	ID     int64
	UserID int64
	RoleID int64
}

// RolePermissionRel 表示角色-权限关系
type RolePermissionRel struct {
	ID           int64
	RoleID       int64
	PermissionID int64
}

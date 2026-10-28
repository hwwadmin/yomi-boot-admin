// Package repo RBAC 域仓储实现
package repo

import (
	"yomi-admin-backend/domains/rbac/pkg/shared/enums"

	"gitee.com/hq_project/yomi-boot/component/common/db/orm"
)

// 表名常量
const (
	tableRole              = "t_role"
	tablePermission        = "t_permission"
	tableUserRoleRel       = "t_user_role_rel"
	tableRolePermissionRel = "t_role_permission_rel"
)

// roleModel 是 t_role 的持久化模型
type roleModel struct {
	Code    string       `gorm:"column:code;size:64;uniqueIndex:uk_code"`
	Name    string       `gorm:"column:name;size:64"`
	Builtin bool         `gorm:"column:builtin"`
	Status  enums.Status `gorm:"column:status;size:20"`
	Remark  string       `gorm:"column:remark;size:255"`
	orm.BaseModel
}

// TableName 返回表名
func (roleModel) TableName() string {
	return tableRole
}

// permissionModel 是 t_permission 的持久化模型
type permissionModel struct {
	Code           string               `gorm:"column:code;size:96;uniqueIndex:uk_code"`
	Name           string               `gorm:"column:name;size:64"`
	PermissionType enums.PermissionType `gorm:"column:permission_type;size:20"`
	UIPath         *string              `gorm:"column:ui_path;size:191;uniqueIndex:uk_ui_path"`
	HTTPMethod     *string              `gorm:"column:http_method;size:16;uniqueIndex:uk_http_route"`
	HTTPPath       *string              `gorm:"column:http_path;size:191;uniqueIndex:uk_http_route"`
	Status         enums.Status         `gorm:"column:status;size:20"`
	Remark         string               `gorm:"column:remark;size:255"`
	orm.BaseModel
}

// TableName 返回表名
func (permissionModel) TableName() string {
	return tablePermission
}

// userRoleRelModel 是 t_user_role_rel 的持久化模型
type userRoleRelModel struct {
	UserID int64 `gorm:"column:user_id;uniqueIndex:uk_user_role"`
	RoleID int64 `gorm:"column:role_id;uniqueIndex:uk_user_role"`
	orm.BaseModel
}

// TableName 返回表名
func (userRoleRelModel) TableName() string {
	return tableUserRoleRel
}

// rolePermissionRelModel 是 t_role_permission_rel 的持久化模型
type rolePermissionRelModel struct {
	RoleID       int64 `gorm:"column:role_id;uniqueIndex:uk_role_permission"`
	PermissionID int64 `gorm:"column:permission_id;uniqueIndex:uk_role_permission"`
	orm.BaseModel
}

// TableName 返回表名
func (rolePermissionRelModel) TableName() string {
	return tableRolePermissionRel
}

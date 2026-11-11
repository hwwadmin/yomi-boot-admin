// Package repo RBAC 域仓储实现
package repo

import (
	"yomi-admin-backend/domains/rbac/pkg/shared/enums"
	"yomi-admin-backend/persist/orm"
)

// 表名常量
const (
	tableRole               = "t_rbac_role"
	tablePermission         = "t_rbac_permission"
	tablePermissionGroup    = "t_rbac_permission_group"
	tableUserRoleRel        = "t_rbac_user_role_rel"
	tableRoleGroupRel       = "t_rbac_role_group_rel"
	tableGroupPermissionRel = "t_rbac_group_permission_rel"
)

// roleModel 是 t_rbac_role 的持久化模型
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

// permissionModel 是 t_rbac_permission 的持久化模型
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

// permissionGroupModel 是 t_rbac_permission_group 的持久化模型
type permissionGroupModel struct {
	Code   string       `gorm:"column:code;size:64;uniqueIndex:uk_code"`
	Name   string       `gorm:"column:name;size:64"`
	Status enums.Status `gorm:"column:status;size:20"`
	Remark string       `gorm:"column:remark;size:255"`
	orm.BaseModel
}

// TableName 返回表名
func (permissionGroupModel) TableName() string {
	return tablePermissionGroup
}

// userRoleRelModel 是 t_rbac_user_role_rel 的持久化模型
type userRoleRelModel struct {
	UserID int64 `gorm:"column:user_id;uniqueIndex:uk_user_role"`
	RoleID int64 `gorm:"column:role_id;uniqueIndex:uk_user_role"`
	orm.BaseModel
}

// TableName 返回表名
func (userRoleRelModel) TableName() string {
	return tableUserRoleRel
}

// roleGroupRelModel 是 t_rbac_role_group_rel 的持久化模型
type roleGroupRelModel struct {
	RoleID  int64 `gorm:"column:role_id;uniqueIndex:uk_role_group"`
	GroupID int64 `gorm:"column:group_id;uniqueIndex:uk_role_group"`
	orm.BaseModel
}

// TableName 返回表名
func (roleGroupRelModel) TableName() string {
	return tableRoleGroupRel
}

// groupPermissionRelModel 是 t_rbac_group_permission_rel 的持久化模型
type groupPermissionRelModel struct {
	GroupID      int64 `gorm:"column:group_id;uniqueIndex:uk_group_permission"`
	PermissionID int64 `gorm:"column:permission_id;uniqueIndex:uk_group_permission"`
	orm.BaseModel
}

// TableName 返回表名
func (groupPermissionRelModel) TableName() string {
	return tableGroupPermissionRel
}

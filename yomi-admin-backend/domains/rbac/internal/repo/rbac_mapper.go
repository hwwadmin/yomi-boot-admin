// Package repo RBAC 域仓储实现
package repo

import (
	"yomi-admin-backend/domains/rbac/pkg/shared/entity"

	"github.com/hwwadmin/yomi-boot/component/core/utils/ptrutil"
)

// toRoleEntity 将角色持久化模型转换为领域实体
func toRoleEntity(row *roleModel) *entity.Role {
	if row == nil {
		return nil
	}
	return &entity.Role{
		ID:         row.ID,
		Code:       row.Code,
		Name:       row.Name,
		Builtin:    row.Builtin,
		Status:     row.Status,
		Remark:     row.Remark,
		CreateTime: row.CreateTime.UnixMilli(),
		UpdateTime: row.UpdateTime.UnixMilli(),
	}
}

// toPermissionEntity 将权限持久化模型转换为领域实体
func toPermissionEntity(row *permissionModel) *entity.Permission {
	if row == nil {
		return nil
	}
	return &entity.Permission{
		ID:         row.ID,
		Code:       row.Code,
		Name:       row.Name,
		Type:       row.PermissionType,
		UIPath:     ptrutil.StringFromPtr(row.UIPath),
		HTTPMethod: ptrutil.StringFromPtr(row.HTTPMethod),
		HTTPPath:   ptrutil.StringFromPtr(row.HTTPPath),
		Status:     row.Status,
		Remark:     row.Remark,
		CreateTime: row.CreateTime.UnixMilli(),
		UpdateTime: row.UpdateTime.UnixMilli(),
	}
}

// toPermissionGroupEntity 将权限组持久化模型转换为领域实体
func toPermissionGroupEntity(row *permissionGroupModel) *entity.PermissionGroup {
	if row == nil {
		return nil
	}
	return &entity.PermissionGroup{
		ID:         row.ID,
		Code:       row.Code,
		Name:       row.Name,
		Status:     row.Status,
		Remark:     row.Remark,
		CreateTime: row.CreateTime.UnixMilli(),
		UpdateTime: row.UpdateTime.UnixMilli(),
	}
}

// toRoleModel 将角色实体转换为持久化模型
func toRoleModel(role *entity.Role, operatorID int64) *roleModel {
	row := &roleModel{
		Code:    role.Code,
		Name:    role.Name,
		Builtin: role.Builtin,
		Status:  role.Status,
		Remark:  role.Remark,
	}
	row.ID = role.ID
	row.CreateBy = ptrutil.PtrFromInt64(operatorID)
	row.UpdateBy = ptrutil.PtrFromInt64(operatorID)
	return row
}

// toPermissionGroupModel 将权限组实体转换为持久化模型
func toPermissionGroupModel(group *entity.PermissionGroup, operatorID int64) *permissionGroupModel {
	row := &permissionGroupModel{
		Code:   group.Code,
		Name:   group.Name,
		Status: group.Status,
		Remark: group.Remark,
	}
	row.ID = group.ID
	row.CreateBy = ptrutil.PtrFromInt64(operatorID)
	row.UpdateBy = ptrutil.PtrFromInt64(operatorID)
	return row
}

// toPermissionModel 将权限实体转换为持久化模型
func toPermissionModel(permission *entity.Permission, operatorID int64) *permissionModel {
	row := &permissionModel{
		Code:           permission.Code,
		Name:           permission.Name,
		PermissionType: permission.Type,
		UIPath:         ptrutil.PtrFromString(permission.UIPath),
		HTTPMethod:     ptrutil.PtrFromString(permission.HTTPMethod),
		HTTPPath:       ptrutil.PtrFromString(permission.HTTPPath),
		Status:         permission.Status,
		Remark:         permission.Remark,
	}
	row.ID = permission.ID
	row.CreateBy = ptrutil.PtrFromInt64(operatorID)
	row.UpdateBy = ptrutil.PtrFromInt64(operatorID)
	return row
}

// toUserRoleRelModels 将用户角色关系转换为持久化模型
func toUserRoleRelModels(rels []*entity.UserRoleRel, operatorID int64) []*userRoleRelModel {
	rows := make([]*userRoleRelModel, 0, len(rels))
	for _, rel := range rels {
		row := &userRoleRelModel{UserID: rel.UserID, RoleID: rel.RoleID}
		row.ID = rel.ID
		row.CreateBy = ptrutil.PtrFromInt64(operatorID)
		row.UpdateBy = ptrutil.PtrFromInt64(operatorID)
		rows = append(rows, row)
	}
	return rows
}

// toRoleGroupRelModels 将角色权限组关系转换为持久化模型
func toRoleGroupRelModels(rels []*entity.RoleGroupRel, operatorID int64) []*roleGroupRelModel {
	rows := make([]*roleGroupRelModel, 0, len(rels))
	for _, rel := range rels {
		row := &roleGroupRelModel{RoleID: rel.RoleID, GroupID: rel.GroupID}
		row.ID = rel.ID
		row.CreateBy = ptrutil.PtrFromInt64(operatorID)
		row.UpdateBy = ptrutil.PtrFromInt64(operatorID)
		rows = append(rows, row)
	}
	return rows
}

// toGroupPermissionRelModels 将权限组权限关系转换为持久化模型
func toGroupPermissionRelModels(rels []*entity.GroupPermissionRel, operatorID int64) []*groupPermissionRelModel {
	rows := make([]*groupPermissionRelModel, 0, len(rels))
	for _, rel := range rels {
		row := &groupPermissionRelModel{GroupID: rel.GroupID, PermissionID: rel.PermissionID}
		row.ID = rel.ID
		row.CreateBy = ptrutil.PtrFromInt64(operatorID)
		row.UpdateBy = ptrutil.PtrFromInt64(operatorID)
		rows = append(rows, row)
	}
	return rows
}

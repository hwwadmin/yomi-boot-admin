// Package repo RBAC 域仓储实现
package repo

import (
	"yomi-admin-backend/domains/rbac/pkg/shared/entity"

	"gitee.com/hq_project/yomi-boot/component/core/utils/ptrutil"
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

// toRolePermissionRelModels 将角色权限关系转换为持久化模型
func toRolePermissionRelModels(rels []*entity.RolePermissionRel, operatorID int64) []*rolePermissionRelModel {
	rows := make([]*rolePermissionRelModel, 0, len(rels))
	for _, rel := range rels {
		row := &rolePermissionRelModel{RoleID: rel.RoleID, PermissionID: rel.PermissionID}
		row.ID = rel.ID
		row.CreateBy = ptrutil.PtrFromInt64(operatorID)
		row.UpdateBy = ptrutil.PtrFromInt64(operatorID)
		rows = append(rows, row)
	}
	return rows
}

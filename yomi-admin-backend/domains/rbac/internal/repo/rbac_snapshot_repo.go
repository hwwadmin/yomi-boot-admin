// Package repo RBAC 域仓储实现
package repo

import (
	"context"
	"yomi-admin-backend/domains/rbac/internal/repo_contract"
	"yomi-admin-backend/domains/rbac/pkg/shared/entity"

	"github.com/hwwadmin/yomi-boot/component/common/db/tx"
	"github.com/hwwadmin/yomi-boot/component/core/errors"
)

var _ repo_contract.AuthGraphReader = (*RbacRepo)(nil)

// LoadAuthGraph 全量读取鉴权投影所需的六表。
// 调用方须在事务内调用：六次读取共享同一事务快照，避免投影出现"角色已绑组、组内尚无权限"的中间态。
func (r *RbacRepo) LoadAuthGraph(ctx context.Context) (*repo_contract.AuthGraph, error) {
	graph := &repo_contract.AuthGraph{
		UserRoles:        make(map[int64][]int64),
		RoleGroups:       make(map[int64][]int64),
		GroupPermissions: make(map[int64][]int64),
	}

	var roleRows []*roleModel
	if err := tx.Exec(ctx, r.DB).Order("id ASC").Find(&roleRows).Error; err != nil {
		return nil, errors.Wrap(err, "读取角色失败")
	}
	graph.Roles = make([]*entity.Role, 0, len(roleRows))
	for _, row := range roleRows {
		graph.Roles = append(graph.Roles, toRoleEntity(row))
	}

	var permissionRows []*permissionModel
	if err := tx.Exec(ctx, r.DB).Order("id ASC").Find(&permissionRows).Error; err != nil {
		return nil, errors.Wrap(err, "读取权限失败")
	}
	graph.Permissions = make([]*entity.Permission, 0, len(permissionRows))
	for _, row := range permissionRows {
		graph.Permissions = append(graph.Permissions, toPermissionEntity(row))
	}

	var groupRows []*permissionGroupModel
	if err := tx.Exec(ctx, r.DB).Order("id ASC").Find(&groupRows).Error; err != nil {
		return nil, errors.Wrap(err, "读取权限组失败")
	}
	graph.PermissionGroups = make([]*entity.PermissionGroup, 0, len(groupRows))
	for _, row := range groupRows {
		graph.PermissionGroups = append(graph.PermissionGroups, toPermissionGroupEntity(row))
	}

	var userRoleRows []*userRoleRelModel
	if err := tx.Exec(ctx, r.DB).Order("id ASC").Find(&userRoleRows).Error; err != nil {
		return nil, errors.Wrap(err, "读取用户角色关系失败")
	}
	for _, row := range userRoleRows {
		graph.UserRoles[row.UserID] = append(graph.UserRoles[row.UserID], row.RoleID)
	}

	var roleGroupRows []*roleGroupRelModel
	if err := tx.Exec(ctx, r.DB).Order("id ASC").Find(&roleGroupRows).Error; err != nil {
		return nil, errors.Wrap(err, "读取角色权限组关系失败")
	}
	for _, row := range roleGroupRows {
		graph.RoleGroups[row.RoleID] = append(graph.RoleGroups[row.RoleID], row.GroupID)
	}

	var groupPermissionRows []*groupPermissionRelModel
	if err := tx.Exec(ctx, r.DB).Order("id ASC").Find(&groupPermissionRows).Error; err != nil {
		return nil, errors.Wrap(err, "读取权限组权限关系失败")
	}
	for _, row := range groupPermissionRows {
		graph.GroupPermissions[row.GroupID] = append(graph.GroupPermissions[row.GroupID], row.PermissionID)
	}

	return graph, nil
}

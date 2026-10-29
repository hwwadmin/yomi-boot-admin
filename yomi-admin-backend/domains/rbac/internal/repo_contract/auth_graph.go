// Package repo_contract RBAC 域仓储契约
package repo_contract

import (
	"context"
	"yomi-admin-backend/domains/rbac/pkg/shared/entity"
)

// AuthGraph 是鉴权投影所需的六表全量数据；关系表按左侧 ID 聚合且保持库内顺序
type AuthGraph struct {
	Roles            []*entity.Role
	Permissions      []*entity.Permission
	PermissionGroups []*entity.PermissionGroup
	UserRoles        map[int64][]int64
	RoleGroups       map[int64][]int64
	GroupPermissions map[int64][]int64
}

// AuthGraphReader 全量读取鉴权投影数据
type AuthGraphReader interface {
	// LoadAuthGraph 读取六表全量数据；须在调用方事务内执行，保证六次读取处于同一快照
	LoadAuthGraph(ctx context.Context) (*AuthGraph, error)
}

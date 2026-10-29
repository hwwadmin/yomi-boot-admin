// Package repo_contract RBAC 域仓储契约
package repo_contract

import (
	"context"
	"yomi-admin-backend/domains/rbac/pkg/shared/entity"
	"yomi-admin-backend/domains/rbac/pkg/shared/enums"
)

// RbacRepo 角色、权限组、权限与关系仓储
type RbacRepo interface {
	// FindRoleByID 按主键查询角色
	FindRoleByID(ctx context.Context, id int64) (*entity.Role, bool, error)
	// FindRoleByCode 按编码查询角色
	FindRoleByCode(ctx context.Context, code string) (*entity.Role, bool, error)
	// FindRolesByIDs 按主键批量查询角色
	FindRolesByIDs(ctx context.Context, ids []int64) ([]*entity.Role, error)
	// ListAllRoles 列出全部角色（含禁用）
	ListAllRoles(ctx context.Context) ([]*entity.Role, error)
	// PageRoles 分页查询角色
	PageRoles(ctx context.Context, keyword string, status enums.Status, page, pageSize int) ([]*entity.Role, int64, error)
	// CreateRole 创建角色
	CreateRole(ctx context.Context, role *entity.Role, operatorID int64) error
	// UpdateRole 更新角色资料与状态
	UpdateRole(ctx context.Context, role *entity.Role, operatorID int64) error

	// FindPermissionGroupByID 按主键查询权限组
	FindPermissionGroupByID(ctx context.Context, id int64) (*entity.PermissionGroup, bool, error)
	// FindPermissionGroupByCode 按编码查询权限组
	FindPermissionGroupByCode(ctx context.Context, code string) (*entity.PermissionGroup, bool, error)
	// FindPermissionGroupsByIDs 按主键批量查询权限组
	FindPermissionGroupsByIDs(ctx context.Context, ids []int64) ([]*entity.PermissionGroup, error)
	// PagePermissionGroups 分页查询权限组
	PagePermissionGroups(ctx context.Context, keyword string, status enums.Status, page, pageSize int) ([]*entity.PermissionGroup, int64, error)
	// CreatePermissionGroup 创建权限组
	CreatePermissionGroup(ctx context.Context, group *entity.PermissionGroup, operatorID int64) error
	// UpdatePermissionGroup 更新权限组资料与状态
	UpdatePermissionGroup(ctx context.Context, group *entity.PermissionGroup, operatorID int64) error

	// FindPermissionByID 按主键查询权限
	FindPermissionByID(ctx context.Context, id int64) (*entity.Permission, bool, error)
	// FindPermissionsByIDs 按主键批量查询权限
	FindPermissionsByIDs(ctx context.Context, ids []int64) ([]*entity.Permission, error)
	// PagePermissions 分页查询权限目录
	PagePermissions(ctx context.Context, keyword string, permissionType enums.PermissionType, status enums.Status, page, pageSize int) ([]*entity.Permission, int64, error)
	// UpdatePermissionStatus 更新权限启停
	UpdatePermissionStatus(ctx context.Context, id int64, status enums.Status, operatorID int64) error

	// ListGroupIDsByRoleID 列出角色已绑权限组 ID
	ListGroupIDsByRoleID(ctx context.Context, roleID int64) ([]int64, error)
	// ReplaceRoleGroups 覆盖角色权限组绑定
	ReplaceRoleGroups(ctx context.Context, roleID int64, rels []*entity.RoleGroupRel, operatorID int64) error
	// ListPermissionIDsByRoleID 列出角色所绑权限组的权限 ID（去重聚合）
	ListPermissionIDsByRoleID(ctx context.Context, roleID int64) ([]int64, error)

	// ListPermissionIDsByGroupID 列出权限组已绑权限 ID
	ListPermissionIDsByGroupID(ctx context.Context, groupID int64) ([]int64, error)
	// ReplaceGroupPermissions 覆盖权限组权限绑定
	ReplaceGroupPermissions(ctx context.Context, groupID int64, rels []*entity.GroupPermissionRel, operatorID int64) error

	// ListRoleIDsByUserID 列出用户已绑角色 ID
	ListRoleIDsByUserID(ctx context.Context, userID int64) ([]int64, error)
	// ListRolesByUserIDs 按用户批量查询已绑角色
	ListRolesByUserIDs(ctx context.Context, userIDs []int64) (map[int64][]*entity.Role, error)
	// ReplaceUserRoles 覆盖用户角色绑定
	ReplaceUserRoles(ctx context.Context, userID int64, rels []*entity.UserRoleRel, operatorID int64) error

	// UserHasEnabledSuperAdmin 判断用户是否拥有启用态超管角色
	UserHasEnabledSuperAdmin(ctx context.Context, userID int64) (bool, error)
	// ListUserIDsWithEnabledSuperAdmin 列出拥有启用态超管角色的用户 ID
	ListUserIDsWithEnabledSuperAdmin(ctx context.Context) ([]int64, error)

	// ExistsRouteAccess 判断用户是否可访问指定 HTTP 方法与路由模板
	ExistsRouteAccess(ctx context.Context, userID int64, method, route string) (bool, error)
	// ListMenusByUserID 列出用户可见的启用菜单
	ListMenusByUserID(ctx context.Context, userID int64) ([]*entity.Permission, error)
	// ListEnabledMenus 列出全部启用菜单
	ListEnabledMenus(ctx context.Context) ([]*entity.Permission, error)
	// ListPermissionCodesByUserID 列出用户拥有的启用权限编码
	ListPermissionCodesByUserID(ctx context.Context, userID int64) ([]string, error)
	// ListEnabledPermissionCodes 列出全部启用权限编码
	ListEnabledPermissionCodes(ctx context.Context) ([]string, error)
}

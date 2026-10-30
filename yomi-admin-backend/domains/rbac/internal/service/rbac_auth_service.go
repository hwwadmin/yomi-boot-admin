// Package service RBAC 授权查询实现
package service

import (
	"context"
	"strings"
	"yomi-admin-backend/domains/rbac/internal/projection"
	"yomi-admin-backend/domains/rbac/internal/repo_contract"
	"yomi-admin-backend/domains/rbac/internal/snapshot"
	"yomi-admin-backend/domains/rbac/pkg/shared/entity"
	"yomi-admin-backend/domains/rbac/service_contract"
)

// RbacAuthService 授权查询实现
//
// 鉴权热路径优先读取 Redis 投影做内存求值；投影不可用时降级到 SQL 直查，
// 保证缓存故障只会影响性能，不会误判权限。
type RbacAuthService struct {
	repo       repo_contract.RbacRepo
	projection *projection.Store
}

var _ service_contract.RbacAuthService = (*RbacAuthService)(nil)

// NewRbacAuthService 创建授权查询服务；authProjection 为 nil 时只走 SQL
func NewRbacAuthService(repo repo_contract.RbacRepo, authProjection *projection.Store) *RbacAuthService {
	return &RbacAuthService{repo: repo, projection: authProjection}
}

// CanAccessAPI 判断用户能否访问指定 HTTP 方法与接口路径模板
func (s *RbacAuthService) CanAccessAPI(ctx context.Context, userID int64, method, path string) (bool, error) {
	if userID <= 0 {
		return false, nil
	}

	if authSnapshot, ok := s.loadSnapshot(ctx); ok {
		return authSnapshot.CanAccessAPI(userID, method, path), nil
	}
	return s.canAccessAPIFromSQL(ctx, userID, method, path)
}

// ListMenus 列出用户可见的启用菜单；超管返回全部启用菜单
func (s *RbacAuthService) ListMenus(ctx context.Context, userID int64) ([]*entity.Permission, error) {
	if userID <= 0 {
		return []*entity.Permission{}, nil
	}

	if authSnapshot, ok := s.loadSnapshot(ctx); ok {
		return authSnapshot.ListMenus(userID), nil
	}
	return s.listMenusFromSQL(ctx, userID)
}

// ListPermissionCodes 列出用户拥有的启用权限编码；超管返回 ["*"]
func (s *RbacAuthService) ListPermissionCodes(ctx context.Context, userID int64) ([]string, error) {
	if userID <= 0 {
		return []string{}, nil
	}

	if authSnapshot, ok := s.loadSnapshot(ctx); ok {
		return authSnapshot.ListPermissionCodes(userID), nil
	}
	return s.listPermissionCodesFromSQL(ctx, userID)
}

// ListRoles 列出用户已绑角色（含停用），用于身份视图展示
func (s *RbacAuthService) ListRoles(ctx context.Context, userID int64) ([]*entity.Role, error) {
	if userID <= 0 {
		return []*entity.Role{}, nil
	}

	rolesByUser, err := s.repo.ListRolesByUserIDs(ctx, []int64{userID})
	if err != nil {
		return nil, err
	}
	roles := rolesByUser[userID]
	if roles == nil {
		return []*entity.Role{}, nil
	}
	return roles, nil
}

// loadSnapshot 读取授权投影；未启用或读取失败时返回 false 由调用方降级
func (s *RbacAuthService) loadSnapshot(ctx context.Context) (*snapshot.Snapshot, bool) {
	if s.projection == nil {
		return nil, false
	}
	authSnapshot, err := s.projection.Load(ctx)
	if err != nil {
		return nil, false
	}
	return authSnapshot, true
}

// canAccessAPIFromSQL 沿 启用角色 → 启用权限组 → 启用接口权限 链路直查
func (s *RbacAuthService) canAccessAPIFromSQL(ctx context.Context, userID int64, method, path string) (bool, error) {
	superAdmin, err := s.repo.UserHasEnabledSuperAdmin(ctx, userID)
	if err != nil {
		return false, err
	}
	if superAdmin {
		return true, nil
	}

	method = strings.ToUpper(strings.TrimSpace(method))
	path = strings.TrimSpace(path)
	if method == "" || path == "" {
		return false, nil
	}
	return s.repo.ExistsRouteAccess(ctx, userID, method, path)
}

// listMenusFromSQL 直查用户可见菜单；超管返回全部启用菜单
func (s *RbacAuthService) listMenusFromSQL(ctx context.Context, userID int64) ([]*entity.Permission, error) {
	superAdmin, err := s.repo.UserHasEnabledSuperAdmin(ctx, userID)
	if err != nil {
		return nil, err
	}
	if superAdmin {
		return s.repo.ListEnabledMenus(ctx)
	}
	return s.repo.ListMenusByUserID(ctx, userID)
}

// listPermissionCodesFromSQL 直查用户权限编码；超管返回 ["*"]
func (s *RbacAuthService) listPermissionCodesFromSQL(ctx context.Context, userID int64) ([]string, error) {
	superAdmin, err := s.repo.UserHasEnabledSuperAdmin(ctx, userID)
	if err != nil {
		return nil, err
	}
	if superAdmin {
		return []string{"*"}, nil
	}
	return s.repo.ListPermissionCodesByUserID(ctx, userID)
}

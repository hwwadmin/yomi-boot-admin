// Package service RBAC 授权查询实现
package service

import (
	"context"
	"strings"
	"yomi-admin-backend/domains/rbac/internal/repo_contract"
	"yomi-admin-backend/domains/rbac/pkg/shared/entity"
	"yomi-admin-backend/domains/rbac/service_contract"
)

// RbacAuthService 授权查询实现
type RbacAuthService struct {
	repo repo_contract.RbacRepo
}

var _ service_contract.RbacAuthService = (*RbacAuthService)(nil)

// NewRbacAuthService 创建授权查询服务
func NewRbacAuthService(repo repo_contract.RbacRepo) *RbacAuthService {
	return &RbacAuthService{repo: repo}
}

// CanAccessAPI 判断用户能否访问指定 HTTP 方法与接口路径模板
func (s *RbacAuthService) CanAccessAPI(ctx context.Context, userID int64, method, path string) (bool, error) {
	if userID <= 0 {
		return false, nil
	}

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

// ListMenus 列出用户可见的启用菜单；超管返回全部启用菜单
func (s *RbacAuthService) ListMenus(ctx context.Context, userID int64) ([]*entity.Permission, error) {
	if userID <= 0 {
		return []*entity.Permission{}, nil
	}

	superAdmin, err := s.repo.UserHasEnabledSuperAdmin(ctx, userID)
	if err != nil {
		return nil, err
	}
	if superAdmin {
		return s.repo.ListEnabledMenus(ctx)
	}
	return s.repo.ListMenusByUserID(ctx, userID)
}

// ListPermissionCodes 列出用户拥有的启用权限编码；超管返回 ["*"]
func (s *RbacAuthService) ListPermissionCodes(ctx context.Context, userID int64) ([]string, error) {
	if userID <= 0 {
		return []string{}, nil
	}

	superAdmin, err := s.repo.UserHasEnabledSuperAdmin(ctx, userID)
	if err != nil {
		return nil, err
	}
	if superAdmin {
		return []string{"*"}, nil
	}
	return s.repo.ListPermissionCodesByUserID(ctx, userID)
}

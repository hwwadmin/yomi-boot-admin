// Package service_contract RBAC 域服务契约
package service_contract

import (
	"context"
	"yomi-admin-backend/domains/rbac/pkg/shared/entity"
)

// RbacAuthService 授权查询（鉴权热路径）
type RbacAuthService interface {
	// CanAccessRoute 判断用户能否访问指定 HTTP 方法与 gin 路由模板
	CanAccessRoute(ctx context.Context, userID int64, method, route string) (bool, error)
	// ListMenus 列出用户可见的启用菜单；超管返回全部启用菜单
	ListMenus(ctx context.Context, userID int64) ([]*entity.Permission, error)
	// ListPermissionCodes 列出用户拥有的启用权限编码；超管返回 ["*"]
	ListPermissionCodes(ctx context.Context, userID int64) ([]string, error)
}

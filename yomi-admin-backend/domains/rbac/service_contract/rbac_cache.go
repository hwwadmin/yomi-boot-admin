// Package service_contract RBAC 域服务契约
package service_contract

import (
	"context"
	"yomi-admin-backend/domains/rbac/pkg/shared/output"
)

// RbacCacheService 维护鉴权投影缓存
type RbacCacheService interface {
	// RefreshCache 从数据库全量重建鉴权投影；缓存未启用时返回 ErrCacheDisabled
	RefreshCache(ctx context.Context) (*output.CacheRefreshOutput, error)
}

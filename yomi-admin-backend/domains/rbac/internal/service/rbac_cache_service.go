// Package service RBAC 授权缓存维护实现
package service

import (
	"context"
	"time"
	"yomi-admin-backend/domains/rbac/internal/projection"
	"yomi-admin-backend/domains/rbac/pkg/shared/output"
	"yomi-admin-backend/domains/rbac/service_contract"

	"github.com/hwwadmin/yomi-boot/component/core/errors"
)

// RbacCacheService 鉴权投影维护
type RbacCacheService struct {
	projection *projection.Store
}

var _ service_contract.RbacCacheService = (*RbacCacheService)(nil)

// NewRbacCacheService 创建投影维护服务；authProjection 为 nil 表示缓存未启用
func NewRbacCacheService(authProjection *projection.Store) *RbacCacheService {
	return &RbacCacheService{projection: authProjection}
}

// RefreshCache 全量重建鉴权投影
func (s *RbacCacheService) RefreshCache(ctx context.Context) (*output.CacheRefreshOutput, error) {
	if s.projection == nil {
		return nil, errors.Sentinel(service_contract.ErrCacheDisabled)
	}

	startedAt := time.Now()
	authSnapshot, err := s.projection.Refresh(ctx)
	if err != nil {
		switch {
		case errors.Is(err, projection.ErrBusy):
			return nil, errors.Sentinel(service_contract.ErrCacheRefreshBusy)
		case errors.Is(err, projection.ErrTooLarge):
			return nil, errors.Sentinel(service_contract.ErrCacheTooLarge)
		default:
			return nil, err
		}
	}

	return &output.CacheRefreshOutput{
		Version:              authSnapshot.Version,
		BuiltAt:              authSnapshot.BuiltAt,
		RoleCount:            authSnapshot.RoleCount(),
		PermissionCount:      authSnapshot.PermissionCount(),
		PermissionGroupCount: authSnapshot.PermissionGroupCount(),
		RelationCount:        authSnapshot.RelationCount(),
		DurationMS:           time.Since(startedAt).Milliseconds(),
	}, nil
}

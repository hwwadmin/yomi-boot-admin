package rbac

import (
	"context"
	rbaccontract "yomi-admin-backend/domains/rbac/service_contract"
)

// CacheRefreshOutput 是鉴权缓存刷新视图
type CacheRefreshOutput struct {
	Version              string `json:"version"`
	BuiltAt              int64  `json:"builtAt"`
	RoleCount            int    `json:"roleCount"`
	PermissionCount      int    `json:"permissionCount"`
	PermissionGroupCount int    `json:"permissionGroupCount"`
	RelationCount        int    `json:"relationCount"`
	DurationMS           int64  `json:"durationMs"`
}

// RefreshRbacCacheUsecase 鉴权缓存刷新编排
type RefreshRbacCacheUsecase struct {
	cache rbaccontract.RbacCacheService
}

// NewRefreshRbacCacheUsecase 创建鉴权缓存刷新用例
func NewRefreshRbacCacheUsecase(cache rbaccontract.RbacCacheService) *RefreshRbacCacheUsecase {
	return &RefreshRbacCacheUsecase{cache: cache}
}

// Execute 全量重建鉴权缓存
func (uc *RefreshRbacCacheUsecase) Execute(ctx context.Context) (*CacheRefreshOutput, error) {
	out, err := uc.cache.RefreshCache(ctx)
	if err != nil {
		return nil, err
	}
	return &CacheRefreshOutput{
		Version:              out.Version,
		BuiltAt:              out.BuiltAt,
		RoleCount:            out.RoleCount,
		PermissionCount:      out.PermissionCount,
		PermissionGroupCount: out.PermissionGroupCount,
		RelationCount:        out.RelationCount,
		DurationMS:           out.DurationMS,
	}, nil
}

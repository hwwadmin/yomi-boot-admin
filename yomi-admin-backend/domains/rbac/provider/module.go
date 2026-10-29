// Package provider RBAC 域装配入口（仅 apps/server/bootstrap 使用）
package provider

import (
	"context"
	"time"
	"yomi-admin-backend/domains/rbac/internal/projection"
	"yomi-admin-backend/domains/rbac/internal/repo"
	"yomi-admin-backend/domains/rbac/internal/service"
	"yomi-admin-backend/domains/rbac/service_contract"

	"gitee.com/hq_project/yomi-boot/component/common/db/tx"
	"gitee.com/hq_project/yomi-boot/component/common/distlock"
	"gitee.com/hq_project/yomi-boot/component/common/redis"
	"gitee.com/hq_project/yomi-boot/component/common/snowflake"
	"gitee.com/hq_project/yomi-boot/component/core/shared/enums"

	"gorm.io/gorm"
)

// Deps 装配依赖
type Deps struct {
	DB          *gorm.DB
	DBType      enums.DbType
	IDGenerator snowflake.Generator
	Redis       redis.RedisService
	Locks       distlock.DistLockFactory
	Tx          tx.Executor
	Cache       CacheOptions
}

// CacheOptions 是鉴权投影缓存配置
type CacheOptions struct {
	Enabled         bool          // 是否启用鉴权缓存
	Key             string        // 投影键
	LockKey         string        // 刷新锁键
	MaxBytes        int           // 投影体积上限
	LockWait        time.Duration // 刷新锁最长等待
	BootstrapWarmup bool          // 启动时投影不可用才构建
	RefreshInterval time.Duration // 定时刷新间隔，<=0 表示关闭
	MemorySync      time.Duration // 内存快照同步周期，<=0 表示关闭内存层
}

// Module 导出本域能力
type Module struct {
	Admin service_contract.RbacAdminService
	Auth  service_contract.RbacAuthService
	Cache service_contract.RbacCacheService
	// StartBackground 启动启动期刷新与定时刷新；缓存未启用时为空操作
	StartBackground func(ctx context.Context)
}

// Provide 装配 rbac 域
func Provide(deps Deps) (Module, error) {
	rbacRepo := repo.NewRbacRepo(deps.DB, deps.DBType)

	var authProjection *projection.Store
	if deps.Cache.Enabled {
		store, err := projection.NewStore(projection.Config{
			Key:             deps.Cache.Key,
			LockKey:         deps.Cache.LockKey,
			MaxBytes:        deps.Cache.MaxBytes,
			LockWait:        deps.Cache.LockWait,
			BootstrapWarmup: deps.Cache.BootstrapWarmup,
			RefreshInterval: deps.Cache.RefreshInterval,
			MemorySync:      deps.Cache.MemorySync,
		}, deps.Redis, deps.Locks, rbacRepo, deps.Tx, deps.IDGenerator)
		if err != nil {
			return Module{}, err
		}
		authProjection = store
	}

	return Module{
		Admin: service.NewRbacAdminService(rbacRepo, deps.IDGenerator),
		Auth:  service.NewRbacAuthService(rbacRepo, authProjection),
		Cache: service.NewRbacCacheService(authProjection),
		StartBackground: func(ctx context.Context) {
			if authProjection != nil {
				authProjection.StartBackground(ctx)
			}
		},
	}, nil
}

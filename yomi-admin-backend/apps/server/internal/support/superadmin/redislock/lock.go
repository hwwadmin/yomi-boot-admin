// Package redislock 是 support/superadmin.Locker 的 Redis 分布式锁实现。
//
// 锁覆盖 usecase 传入的整段闭包（检查 + 写入 + 事务提交），
// 并在锁丢失时通过 HeldContext 取消业务 context，让事务回滚。
package redislock

import (
	"context"
	"time"
	"yomi-admin-backend/apps/server/internal/support/superadmin"

	"gitee.com/hq_project/yomi-boot/component/common/distlock"
	"gitee.com/hq_project/yomi-boot/component/core/errors"
	"gitee.com/hq_project/yomi-boot/component/core/logger"
)

const (
	lockTTL            = 30 * time.Second
	lockWatchdogPeriod = 10 * time.Second
	lockMaxHold        = 2 * time.Minute
	lockAcquireTimeout = 5 * time.Second
	lockReleaseTimeout = 3 * time.Second
)

// Locker 用 Redis 分布式锁串行化超管变更。
type Locker struct {
	factory distlock.DistLockFactory
}

var _ superadmin.Locker = (*Locker)(nil)

// New 创建超管变更锁
func New(factory distlock.DistLockFactory) *Locker {
	return &Locker{factory: factory}
}

// WithLock 获取全局超管锁后执行 fn；抢不到锁时 fail closed。
func (l *Locker) WithLock(ctx context.Context, fn func(context.Context) error) error {
	lock := l.factory.GetLock(superadmin.LockKey,
		distlock.WithTTL(lockTTL),
		distlock.WithWatchdog(true),
		distlock.WithWatchdogInterval(lockWatchdogPeriod),
		distlock.WithMaxHoldDuration(lockMaxHold),
	)

	acquired, err := lock.TryLockWithTimeout(ctx, lockAcquireTimeout)
	if err != nil {
		logger.Error(ctx, "获取超管变更锁失败", "key", superadmin.LockKey, "error", err)
		return errors.Sentinel(superadmin.ErrSuperAdminLockUnavailable)
	}
	if !acquired {
		return errors.Sentinel(superadmin.ErrSuperAdminBusy)
	}

	defer func() {
		unlockCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), lockReleaseTimeout)
		defer cancel()
		if err := lock.Unlock(unlockCtx); err != nil {
			logger.Error(ctx, "释放超管变更锁失败", "key", superadmin.LockKey, "error", err)
		}
	}()

	heldCtx, stop := distlock.BindHeldContext(ctx, lock)
	defer stop()

	return fn(heldCtx)
}

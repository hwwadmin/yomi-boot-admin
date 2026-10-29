package bootstrap

import (
	"context"
	"time"
	"yomi-admin-backend/apps/server/internal/support/superadmin"

	"gitee.com/hq_project/yomi-boot/component/common/distlock"
	"gitee.com/hq_project/yomi-boot/component/core/errors"
	"gitee.com/hq_project/yomi-boot/component/core/logger"
)

const (
	superAdminLockTTL            = 30 * time.Second
	superAdminLockWatchdogPeriod = 10 * time.Second
	superAdminLockMaxHold        = 2 * time.Minute
	superAdminLockAcquireTimeout = 5 * time.Second
	superAdminLockReleaseTimeout = 3 * time.Second
)

// superAdminLocker 用 Redis 分布式锁串行化超管变更。
//
// 锁覆盖 usecase 传入的整段闭包（检查 + 写入 + 事务提交），
// 并在锁丢失时通过 HeldContext 取消业务 context，让事务回滚。
type superAdminLocker struct {
	factory distlock.DistLockFactory
}

var _ superadmin.Locker = (*superAdminLocker)(nil)

func newSuperAdminLocker(factory distlock.DistLockFactory) *superAdminLocker {
	return &superAdminLocker{factory: factory}
}

// WithLock 获取全局超管锁后执行 fn；抢不到锁时 fail closed。
func (l *superAdminLocker) WithLock(ctx context.Context, fn func(context.Context) error) error {
	lock := l.factory.GetLock(superadmin.LockKey,
		distlock.WithTTL(superAdminLockTTL),
		distlock.WithWatchdog(true),
		distlock.WithWatchdogInterval(superAdminLockWatchdogPeriod),
		distlock.WithMaxHoldDuration(superAdminLockMaxHold),
	)

	acquired, err := lock.TryLockWithTimeout(ctx, superAdminLockAcquireTimeout)
	if err != nil {
		logger.Error(ctx, "获取超管变更锁失败", "key", superadmin.LockKey, "error", err)
		return errors.Sentinel(superadmin.ErrSuperAdminLockUnavailable)
	}
	if !acquired {
		return errors.Sentinel(superadmin.ErrSuperAdminBusy)
	}

	defer func() {
		unlockCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), superAdminLockReleaseTimeout)
		defer cancel()
		if err := lock.Unlock(unlockCtx); err != nil {
			logger.Error(ctx, "释放超管变更锁失败", "key", superadmin.LockKey, "error", err)
		}
	}()

	heldCtx, stop := distlock.BindHeldContext(ctx, lock)
	defer stop()

	return fn(heldCtx)
}

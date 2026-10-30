package redislock

import (
	"context"
	"fmt"
	"testing"
	"time"
	"yomi-admin-backend/apps/server/internal/support/superadmin"

	"gitee.com/hq_project/yomi-boot/component/common/distlock"
	"gitee.com/hq_project/yomi-boot/component/core/errors"
)

// fakeLock 是可配置结果的锁测试桩
type fakeLock struct {
	acquired   bool
	acquireErr error
	unlockErr  error
	unlocks    int
}

func (f *fakeLock) Lock(context.Context) error { return f.acquireErr }

func (f *fakeLock) TryLock(context.Context) (bool, error) { return f.acquired, f.acquireErr }

func (f *fakeLock) TryLockWithTimeout(context.Context, time.Duration) (bool, error) {
	return f.acquired, f.acquireErr
}

func (f *fakeLock) Unlock(context.Context) error {
	f.unlocks++
	return f.unlockErr
}

func (f *fakeLock) HeldContext() context.Context { return nil }

func (f *fakeLock) Key() string { return superadmin.LockKey }

// fakeLockFactory 总是返回同一个测试锁
type fakeLockFactory struct {
	lock *fakeLock
}

func (f *fakeLockFactory) GetLock(string, ...distlock.LockOption) distlock.Lock {
	return f.lock
}

func TestLockerRunsFnAndUnlocks(t *testing.T) {
	lock := &fakeLock{acquired: true}
	locker := New(&fakeLockFactory{lock: lock})

	called := false
	err := locker.WithLock(context.Background(), func(context.Context) error {
		called = true
		return nil
	})
	if err != nil {
		t.Fatalf("WithLock 失败: %v", err)
	}
	if !called || lock.unlocks != 1 {
		t.Fatalf("闭包或释放未执行: called=%v unlocks=%d", called, lock.unlocks)
	}
}

func TestLockerFailsClosedWhenBusy(t *testing.T) {
	lock := &fakeLock{acquired: false}
	locker := New(&fakeLockFactory{lock: lock})

	called := false
	err := locker.WithLock(context.Background(), func(context.Context) error {
		called = true
		return nil
	})
	if !errors.Is(err, superadmin.ErrSuperAdminBusy) {
		t.Fatalf("期望 ErrSuperAdminBusy，实际 %v", err)
	}
	if called || lock.unlocks != 0 {
		t.Fatalf("未持锁不应执行闭包或释放: called=%v unlocks=%d", called, lock.unlocks)
	}
}

func TestLockerFailsClosedWhenRedisError(t *testing.T) {
	lock := &fakeLock{acquireErr: fmt.Errorf("redis down")}
	locker := New(&fakeLockFactory{lock: lock})

	called := false
	err := locker.WithLock(context.Background(), func(context.Context) error {
		called = true
		return nil
	})
	if !errors.Is(err, superadmin.ErrSuperAdminLockUnavailable) {
		t.Fatalf("期望 ErrSuperAdminLockUnavailable，实际 %v", err)
	}
	if called || lock.unlocks != 0 {
		t.Fatalf("未持锁不应执行闭包或释放: called=%v unlocks=%d", called, lock.unlocks)
	}
}

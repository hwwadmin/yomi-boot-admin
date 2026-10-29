package user

import "context"

// lastSuperGuard 最后一名启用超管保护策略
type lastSuperGuard interface {
	WithLock(ctx context.Context, fn func(context.Context) error) error
	EnsureNotLastSuper(ctx context.Context, targetUserID int64) error
}

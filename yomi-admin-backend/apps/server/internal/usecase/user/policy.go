package user

import "context"

// lastSuperGuard 最后一名启用超管保护策略
type lastSuperGuard interface {
	EnsureNotLastSuper(ctx context.Context, targetUserID int64) error
}

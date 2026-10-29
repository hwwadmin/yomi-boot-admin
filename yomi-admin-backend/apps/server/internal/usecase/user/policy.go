package user

import "context"

// superAdminGuard 超管安全策略：串行化超管相关变更，并保护超管账号不被越权处置。
type superAdminGuard interface {
	// WithLock 在超管变更锁内执行闭包，覆盖检查、写入与事务提交
	WithLock(ctx context.Context, fn func(context.Context) error) error
	// EnsureCanManageUser 校验能否停用或删除目标用户
	EnsureCanManageUser(ctx context.Context, operatorID, targetUserID int64) error
}

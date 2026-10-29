// Package superadmin 供 usecase 复用的超管安全策略：超管角色的授予、回收，以及超管账号的停用、删除与最后一名启用超管保护。
package superadmin

import "gitee.com/hq_project/yomi-boot/component/core/errors"

var (
	// ErrGrantSuperDenied 只有超管可以变更超管角色
	ErrGrantSuperDenied = errors.NewBiz("ADMIN_ONLY_SUPER_CAN_GRANT_SUPER", "只有超级管理员可以变更超级管理员角色")
	// ErrUnbindSelfSuper 不能解绑自己的超管角色
	ErrUnbindSelfSuper = errors.NewBiz("ADMIN_CANNOT_UNBIND_SELF_SUPER", "不能解绑自己的超级管理员角色")
	// ErrManageSuperDenied 只有超管可以停用或删除超管账号
	ErrManageSuperDenied = errors.NewBiz("ADMIN_ONLY_SUPER_CAN_MANAGE_SUPER", "只有超级管理员可以停用或删除超级管理员账号")
	// ErrLastSuperAdmin 不能移除最后一名启用状态的超管
	ErrLastSuperAdmin = errors.NewBiz("ADMIN_LAST_SUPER_ADMIN", "不能移除最后一名启用状态的超级管理员")
	// ErrSuperAdminBusy 超管变更操作正在进行中
	ErrSuperAdminBusy = errors.NewBiz("ADMIN_SUPER_ADMIN_BUSY", "超管变更操作进行中，请稍后重试")
	// ErrSuperAdminLockUnavailable 超管变更锁暂时不可用
	ErrSuperAdminLockUnavailable = errors.NewBiz("ADMIN_SUPER_ADMIN_LOCK_UNAVAILABLE", "超管变更暂时不可用，请稍后重试")
)

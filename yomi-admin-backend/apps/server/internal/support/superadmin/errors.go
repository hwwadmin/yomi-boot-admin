// Package superadmin 供 usecase 复用的超管安全策略：超管角色的授予、回收与最后一名启用超管保护。
package superadmin

import "gitee.com/hq_project/yomi-boot/component/core/errors"

var (
	// ErrGrantSuperDenied 只有超管可以变更超管角色
	ErrGrantSuperDenied = errors.NewBiz("ADMIN_ONLY_SUPER_CAN_GRANT_SUPER", "只有超级管理员可以变更超级管理员角色")
	// ErrUnbindSelfSuper 不能解绑自己的超管角色
	ErrUnbindSelfSuper = errors.NewBiz("ADMIN_CANNOT_UNBIND_SELF_SUPER", "不能解绑自己的超级管理员角色")
	// ErrLastSuperAdmin 不能移除最后一名启用状态的超管
	ErrLastSuperAdmin = errors.NewBiz("ADMIN_LAST_SUPER_ADMIN", "不能移除最后一名启用状态的超级管理员")
)

// Package superadmin 供 usecase 复用的超管安全策略：超管角色的授予、回收，以及超管账号的停用、删除与最后一名启用超管保护。
//
// 本包属于入口 support 层，只依赖各领域的公开契约与共享类型，不承载协议、存储或单域不变量。
package superadmin

import (
	"context"

	userentity "yomi-admin-backend/domains/user/pkg/shared/entity"
	userenums "yomi-admin-backend/domains/user/pkg/shared/enums"
	usercontract "yomi-admin-backend/domains/user/service_contract"

	rbacentity "yomi-admin-backend/domains/rbac/pkg/shared/entity"
	rbacenums "yomi-admin-backend/domains/rbac/pkg/shared/enums"

	"gitee.com/hq_project/yomi-boot/component/core/errors"
)

// UserReader 读取用户状态
type UserReader interface {
	GetByID(ctx context.Context, id int64) (*userentity.User, error)
}

// RoleReader 读取超管关系
type RoleReader interface {
	ListUserIDsWithEnabledSuperAdmin(ctx context.Context) ([]int64, error)
}

// Locker 串行化所有会改变启用态超管集合的操作。
//
// 实现必须保证：锁覆盖「检查 + 写入 + 提交」整段；抢不到锁时 fail closed。
type Locker interface {
	WithLock(ctx context.Context, fn func(context.Context) error) error
}

// LockKey 是超管变更操作共享的分布式锁 key。
//
// 多入口只复用领域时，新入口若也会改变超管状态，必须复用同一个 key，
// 否则入口之间的竞态会重新出现。
const LockKey = "yomi-admin:superadmin:guard"

// Guard 超管相关编排策略
type Guard struct {
	users UserReader
	rbac  RoleReader
	lock  Locker
}

// New 创建超管策略守卫
func New(users UserReader, rbac RoleReader, lock Locker) *Guard {
	return &Guard{users: users, rbac: rbac, lock: lock}
}

// WithLock 在超管变更锁内执行 fn；fn 必须包含检查、写入和事务提交。
func (g *Guard) WithLock(ctx context.Context, fn func(context.Context) error) error {
	return g.lock.WithLock(ctx, fn)
}

// EnsureCanManageUser 校验停用或删除目标用户是否允许。
//
// 必须在 WithLock 内调用，且锁需覆盖后续写入与提交。
//
// 规则：
//   - 目标是启用态超管时，操作者也必须是启用态超管，否则普通管理员可以把超管顶掉；
//   - 目标是最后一名启用态超管时拒绝，保证启用集合不会因此清空。
func (g *Guard) EnsureCanManageUser(ctx context.Context, operatorID, targetUserID int64) error {
	enabledSupers, err := g.enabledSuperUserIDs(ctx)
	if err != nil {
		return err
	}

	_, targetSuper := enabledSupers[targetUserID]
	if !targetSuper {
		return nil
	}
	if _, actorSuper := enabledSupers[operatorID]; !actorSuper {
		return errors.Sentinel(ErrManageSuperDenied)
	}
	if len(enabledSupers) == 1 {
		return errors.Sentinel(ErrLastSuperAdmin)
	}
	return nil
}

// EnsureCanResetPassword 校验操作者能否重置他人密码。
//
// 规则：只有启用态超管可以重置他人密码；重置自己的密码必须走"修改密码"并校验原密码。
// 重置密码不改变启用态超管集合，因此不需要持锁。
func (g *Guard) EnsureCanResetPassword(ctx context.Context, operatorID int64) error {
	if operatorID <= 0 {
		return errors.Sentinel(ErrResetPasswordDenied)
	}

	enabledSupers, err := g.enabledSuperUserIDs(ctx)
	if err != nil {
		return err
	}
	if _, ok := enabledSupers[operatorID]; !ok {
		return errors.Sentinel(ErrResetPasswordDenied)
	}
	return nil
}

// EnsureCanBindRoles 校验本次用户角色变更是否允许。
//
// 必须在 WithLock 内调用，且锁需覆盖后续写入与提交。
//
// roles 由调用方传入，且必须是已规范化并校验过的角色（存在、启用）；
// 本方法只判断超管身份与集合变化，不重复实现“ID 规范化 / 存在性 / 启用态”规则。
//
// 规则：
//   - 授予或回收超管角色，操作者必须是启用态超管；
//   - 超管不能解绑自己的超管角色；
//   - 变更后启用态超管集合不能为空。
func (g *Guard) EnsureCanBindRoles(ctx context.Context, operatorID, targetUserID int64, roles []*rbacentity.Role) error {
	target, err := g.users.GetByID(ctx, targetUserID)
	if err != nil {
		return err
	}

	enabledSupers, err := g.enabledSuperUserIDs(ctx)
	if err != nil {
		return err
	}

	willHaveSuper := false
	for _, role := range roles {
		if role.Code == rbacenums.RoleCodeSuperAdmin && role.Status == rbacenums.StatusEnabled {
			willHaveSuper = true
			break
		}
	}
	_, actorSuper := enabledSupers[operatorID]
	_, targetSuper := enabledSupers[targetUserID]

	if (willHaveSuper || targetSuper) && !actorSuper {
		return errors.Sentinel(ErrGrantSuperDenied)
	}
	if operatorID == targetUserID && targetSuper && !willHaveSuper {
		return errors.Sentinel(ErrUnbindSelfSuper)
	}

	if target.Status == userenums.UserStatusEnabled {
		if willHaveSuper {
			enabledSupers[targetUserID] = struct{}{}
		} else {
			delete(enabledSupers, targetUserID)
		}
	}
	if len(enabledSupers) == 0 {
		return errors.Sentinel(ErrLastSuperAdmin)
	}
	return nil
}

// enabledSuperUserIDs 返回绑定启用态超管角色且账号启用的用户 ID
func (g *Guard) enabledSuperUserIDs(ctx context.Context) (map[int64]struct{}, error) {
	ids, err := g.rbac.ListUserIDsWithEnabledSuperAdmin(ctx)
	if err != nil {
		return nil, err
	}

	userIDs := make(map[int64]struct{}, len(ids))
	for _, id := range ids {
		user, err := g.users.GetByID(ctx, id)
		if err != nil {
			// 用户已删除时关系视为孤儿，不计入启用超管
			if errors.Is(err, usercontract.ErrUserNotFound) {
				continue
			}
			return nil, err
		}
		if user.Status == userenums.UserStatusEnabled {
			userIDs[id] = struct{}{}
		}
	}
	return userIDs, nil
}

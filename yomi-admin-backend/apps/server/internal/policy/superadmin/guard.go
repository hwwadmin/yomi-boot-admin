// Package superadmin 编排层跨域策略：超管角色的授予、回收与最后一名启用超管保护。
//
// 本包只依赖各领域的公开契约与共享类型，不承载协议、存储或单域不变量。
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

// RoleReader 读取角色与超管关系
type RoleReader interface {
	ListUserIDsWithEnabledSuperAdmin(ctx context.Context) ([]int64, error)
	ListRolesByIDs(ctx context.Context, ids []int64) ([]*rbacentity.Role, error)
}

// Guard 超管相关编排策略
type Guard struct {
	users UserReader
	rbac  RoleReader
}

// New 创建超管策略守卫
func New(users UserReader, rbac RoleReader) *Guard {
	return &Guard{users: users, rbac: rbac}
}

// EnsureNotLastSuper 目标用户是唯一启用态超管时拒绝禁用或删除
func (g *Guard) EnsureNotLastSuper(ctx context.Context, targetUserID int64) error {
	userIDs, err := g.enabledSuperUserIDs(ctx)
	if err != nil {
		return err
	}
	if len(userIDs) == 1 {
		if _, ok := userIDs[targetUserID]; ok {
			return errors.Sentinel(ErrLastSuperAdmin)
		}
	}
	return nil
}

// EnsureCanBindRoles 校验本次用户角色变更是否允许。
//
// 规则：
//   - 授予或回收超管角色，操作者必须是启用态超管；
//   - 超管不能解绑自己的超管角色；
//   - 变更后启用态超管集合不能为空。
func (g *Guard) EnsureCanBindRoles(ctx context.Context, operatorID, targetUserID int64, roleIDs []int64) error {
	target, err := g.users.GetByID(ctx, targetUserID)
	if err != nil {
		return err
	}

	enabledSupers, err := g.enabledSuperUserIDs(ctx)
	if err != nil {
		return err
	}
	roles, err := g.rbac.ListRolesByIDs(ctx, roleIDs)
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

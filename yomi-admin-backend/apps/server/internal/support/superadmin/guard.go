// Package superadmin 校验"最后一名启用超管"跨域不变量
package superadmin

import (
	"context"
	userentity "yomi-admin-backend/domains/user/pkg/shared/entity"
	userenums "yomi-admin-backend/domains/user/pkg/shared/enums"
	usercontract "yomi-admin-backend/domains/user/service_contract"

	rbacentity "yomi-admin-backend/domains/rbac/pkg/shared/entity"
	rbacenums "yomi-admin-backend/domains/rbac/pkg/shared/enums"
	rbaccontract "yomi-admin-backend/domains/rbac/service_contract"

	"gitee.com/hq_project/yomi-boot/component/core/errors"
)

// UserGetter 读取用户状态
type UserGetter interface {
	GetByID(ctx context.Context, id int64) (*userentity.User, error)
}

// RbacSuperReader 读取超管角色关系
type RbacSuperReader interface {
	ListUserIDsWithEnabledSuperAdmin(ctx context.Context) ([]int64, error)
	GetRole(ctx context.Context, id int64) (*rbacentity.RoleDetail, error)
}

// Guard 最后一名启用超管保护
type Guard struct {
	users UserGetter
	rbac  RbacSuperReader
}

// New 创建超管保护守卫
func New(users UserGetter, rbac RbacSuperReader) *Guard {
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
			return errors.Sentinel(rbaccontract.ErrLastSuperAdmin)
		}
	}
	return nil
}

// EnsureAfterBind 用户角色变更后启用态超管集合不能为空
func (g *Guard) EnsureAfterBind(ctx context.Context, targetUserID int64, roleIDs []int64) error {
	userIDs, err := g.enabledSuperUserIDs(ctx)
	if err != nil {
		return err
	}

	target, err := g.users.GetByID(ctx, targetUserID)
	if err != nil {
		return err
	}
	if target.Status != userenums.UserStatusEnabled {
		return nil
	}

	willHaveSuper, err := g.hasEnabledSuperRole(ctx, roleIDs)
	if err != nil {
		return err
	}
	if willHaveSuper {
		userIDs[targetUserID] = struct{}{}
	} else {
		delete(userIDs, targetUserID)
	}
	if len(userIDs) == 0 {
		return errors.Sentinel(rbaccontract.ErrLastSuperAdmin)
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

// hasEnabledSuperRole 判断角色 ID 列表中是否包含启用态超管角色
func (g *Guard) hasEnabledSuperRole(ctx context.Context, roleIDs []int64) (bool, error) {
	for _, roleID := range roleIDs {
		detail, err := g.rbac.GetRole(ctx, roleID)
		if err != nil {
			return false, err
		}
		if detail.Role.Code == rbacenums.RoleCodeSuperAdmin && detail.Role.Status == rbacenums.StatusEnabled {
			return true, nil
		}
	}
	return false, nil
}

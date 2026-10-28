package superadmin

import (
	"context"
	"testing"
	userentity "yomi-admin-backend/domains/user/pkg/shared/entity"
	userenums "yomi-admin-backend/domains/user/pkg/shared/enums"
	usercontract "yomi-admin-backend/domains/user/service_contract"

	rbacentity "yomi-admin-backend/domains/rbac/pkg/shared/entity"
	rbacenums "yomi-admin-backend/domains/rbac/pkg/shared/enums"
	rbaccontract "yomi-admin-backend/domains/rbac/service_contract"

	"gitee.com/hq_project/yomi-boot/component/core/errors"
)

// fakeUserGetter 是用户查询测试桩
type fakeUserGetter struct {
	users map[int64]*userentity.User
}

func (f *fakeUserGetter) GetByID(_ context.Context, id int64) (*userentity.User, error) {
	user, ok := f.users[id]
	if !ok {
		return nil, errors.Sentinel(usercontract.ErrUserNotFound)
	}
	return user, nil
}

// fakeSuperReader 是超管关系测试桩
type fakeSuperReader struct {
	userIDs []int64
	roles   map[int64]*rbacentity.RoleDetail
}

func (f *fakeSuperReader) ListUserIDsWithEnabledSuperAdmin(context.Context) ([]int64, error) {
	return f.userIDs, nil
}

func (f *fakeSuperReader) GetRole(_ context.Context, id int64) (*rbacentity.RoleDetail, error) {
	detail, ok := f.roles[id]
	if !ok {
		return nil, errors.Sentinel(rbaccontract.ErrRoleNotFound)
	}
	return detail, nil
}

func superRoleDetail() *rbacentity.RoleDetail {
	return &rbacentity.RoleDetail{Role: &rbacentity.Role{
		ID:     1,
		Code:   rbacenums.RoleCodeSuperAdmin,
		Status: rbacenums.StatusEnabled,
	}}
}

func TestEnsureNotLastSuper(t *testing.T) {
	ctx := context.Background()
	users := &fakeUserGetter{users: map[int64]*userentity.User{
		1: {ID: 1, Status: userenums.UserStatusEnabled},
		2: {ID: 2, Status: userenums.UserStatusDisabled},
	}}

	// 唯一启用超管就是目标：拒绝
	guard := New(users, &fakeSuperReader{userIDs: []int64{1}})
	if err := guard.EnsureNotLastSuper(ctx, 1); !errors.Is(err, rbaccontract.ErrLastSuperAdmin) {
		t.Fatalf("唯一超管期望 ErrLastSuperAdmin，实际 %v", err)
	}

	// 目标不是启用超管：放行
	if err := guard.EnsureNotLastSuper(ctx, 2); err != nil {
		t.Fatalf("非超管不应受限: %v", err)
	}

	// 两名启用超管：放行
	users.users[3] = &userentity.User{ID: 3, Status: userenums.UserStatusEnabled}
	guard = New(users, &fakeSuperReader{userIDs: []int64{1, 3}})
	if err := guard.EnsureNotLastSuper(ctx, 1); err != nil {
		t.Fatalf("存在其他超管时应放行: %v", err)
	}

	// 孤儿关系不计入启用超管
	guard = New(users, &fakeSuperReader{userIDs: []int64{1, 99}})
	if err := guard.EnsureNotLastSuper(ctx, 1); !errors.Is(err, rbaccontract.ErrLastSuperAdmin) {
		t.Fatalf("孤儿关系不应计入: %v", err)
	}
}

func TestEnsureAfterBind(t *testing.T) {
	ctx := context.Background()
	roles := map[int64]*rbacentity.RoleDetail{
		1: superRoleDetail(),
		2: {Role: &rbacentity.Role{ID: 2, Code: "viewer", Status: rbacenums.StatusEnabled}},
	}
	users := &fakeUserGetter{users: map[int64]*userentity.User{
		1: {ID: 1, Status: userenums.UserStatusEnabled},
	}}
	reader := &fakeSuperReader{userIDs: []int64{1}, roles: roles}
	guard := New(users, reader)

	// 唯一超管解绑超管角色：拒绝
	if err := guard.EnsureAfterBind(ctx, 1, []int64{2}); !errors.Is(err, rbaccontract.ErrLastSuperAdmin) {
		t.Fatalf("解绑唯一超管期望 ErrLastSuperAdmin，实际 %v", err)
	}

	// 保留超管角色：放行
	if err := guard.EnsureAfterBind(ctx, 1, []int64{1, 2}); err != nil {
		t.Fatalf("保留超管角色应放行: %v", err)
	}

	// 用户不存在时返回用户域错误
	if err := guard.EnsureAfterBind(ctx, 99, []int64{1}); !errors.Is(err, usercontract.ErrUserNotFound) {
		t.Fatalf("不存在用户期望 ErrUserNotFound，实际 %v", err)
	}

	// 禁用用户不参与启用超管集合，绑定角色不影响结果
	users.users[2] = &userentity.User{ID: 2, Status: userenums.UserStatusDisabled}
	reader.userIDs = []int64{1}
	if err := guard.EnsureAfterBind(ctx, 2, []int64{2}); err != nil {
		t.Fatalf("禁用用户绑定不应影响启用超管: %v", err)
	}
}

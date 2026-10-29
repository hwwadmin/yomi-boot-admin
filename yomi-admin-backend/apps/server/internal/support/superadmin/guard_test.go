package superadmin

import (
	"context"
	"testing"

	userentity "yomi-admin-backend/domains/user/pkg/shared/entity"
	userenums "yomi-admin-backend/domains/user/pkg/shared/enums"
	usercontract "yomi-admin-backend/domains/user/service_contract"

	rbacentity "yomi-admin-backend/domains/rbac/pkg/shared/entity"
	rbacenums "yomi-admin-backend/domains/rbac/pkg/shared/enums"

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

// fakeSuperReader 是角色与超管关系测试桩
type fakeSuperReader struct {
	userIDs []int64
	roles   map[int64]*rbacentity.Role
}

func (f *fakeSuperReader) ListUserIDsWithEnabledSuperAdmin(context.Context) ([]int64, error) {
	return f.userIDs, nil
}

func (f *fakeSuperReader) ListRolesByIDs(_ context.Context, ids []int64) ([]*rbacentity.Role, error) {
	roles := make([]*rbacentity.Role, 0, len(ids))
	for _, id := range ids {
		if role, ok := f.roles[id]; ok {
			roles = append(roles, role)
		}
	}
	return roles, nil
}

func testRoles() map[int64]*rbacentity.Role {
	return map[int64]*rbacentity.Role{
		1: {ID: 1, Code: rbacenums.RoleCodeSuperAdmin, Status: rbacenums.StatusEnabled},
		2: {ID: 2, Code: "viewer", Status: rbacenums.StatusEnabled},
	}
}

// fakeLocker 直接在当前协程执行闭包并记录是否被调用
type fakeLocker struct {
	used bool
}

func (f *fakeLocker) WithLock(ctx context.Context, fn func(context.Context) error) error {
	f.used = true
	return fn(ctx)
}

func TestWithLockDelegates(t *testing.T) {
	locker := &fakeLocker{}
	guard := New(&fakeUserGetter{users: map[int64]*userentity.User{}}, &fakeSuperReader{}, locker)

	called := false
	if err := guard.WithLock(context.Background(), func(context.Context) error {
		called = true
		return nil
	}); err != nil {
		t.Fatalf("WithLock 失败: %v", err)
	}
	if !locker.used || !called {
		t.Fatalf("WithLock 未委托锁或未执行闭包: locker=%v called=%v", locker.used, called)
	}
}

func TestEnsureNotLastSuper(t *testing.T) {
	ctx := context.Background()
	users := &fakeUserGetter{users: map[int64]*userentity.User{
		1: {ID: 1, Status: userenums.UserStatusEnabled},
		2: {ID: 2, Status: userenums.UserStatusDisabled},
	}}

	// 唯一启用超管就是目标：拒绝
	guard := New(users, &fakeSuperReader{userIDs: []int64{1}}, &fakeLocker{})
	if err := guard.EnsureNotLastSuper(ctx, 1); !errors.Is(err, ErrLastSuperAdmin) {
		t.Fatalf("唯一超管期望 ErrLastSuperAdmin，实际 %v", err)
	}

	// 目标不是启用超管：放行
	if err := guard.EnsureNotLastSuper(ctx, 2); err != nil {
		t.Fatalf("非超管不应受限: %v", err)
	}

	// 两名启用超管：放行
	users.users[3] = &userentity.User{ID: 3, Status: userenums.UserStatusEnabled}
	guard = New(users, &fakeSuperReader{userIDs: []int64{1, 3}}, &fakeLocker{})
	if err := guard.EnsureNotLastSuper(ctx, 1); err != nil {
		t.Fatalf("存在其他超管时应放行: %v", err)
	}

	// 孤儿关系不计入启用超管
	guard = New(users, &fakeSuperReader{userIDs: []int64{1, 99}}, &fakeLocker{})
	if err := guard.EnsureNotLastSuper(ctx, 1); !errors.Is(err, ErrLastSuperAdmin) {
		t.Fatalf("孤儿关系不应计入: %v", err)
	}
}

func TestEnsureCanBindRoles(t *testing.T) {
	ctx := context.Background()
	users := &fakeUserGetter{users: map[int64]*userentity.User{
		1: {ID: 1, Status: userenums.UserStatusEnabled},
		2: {ID: 2, Status: userenums.UserStatusEnabled},
		3: {ID: 3, Status: userenums.UserStatusEnabled},
		4: {ID: 4, Status: userenums.UserStatusDisabled},
	}}
	reader := &fakeSuperReader{userIDs: []int64{1}, roles: testRoles()}
	guard := New(users, reader, &fakeLocker{})

	// 非超管授予超管
	if err := guard.EnsureCanBindRoles(ctx, 2, 3, []int64{1}); !errors.Is(err, ErrGrantSuperDenied) {
		t.Fatalf("非超管授予超管期望 ErrGrantSuperDenied，实际 %v", err)
	}
	// 非超管回收他人超管
	if err := guard.EnsureCanBindRoles(ctx, 2, 1, []int64{2}); !errors.Is(err, ErrGrantSuperDenied) {
		t.Fatalf("非超管回收超管期望 ErrGrantSuperDenied，实际 %v", err)
	}
	// 超管解绑自己
	if err := guard.EnsureCanBindRoles(ctx, 1, 1, []int64{2}); !errors.Is(err, ErrUnbindSelfSuper) {
		t.Fatalf("超管自解绑期望 ErrUnbindSelfSuper，实际 %v", err)
	}
	// 超管给他人绑定超管
	if err := guard.EnsureCanBindRoles(ctx, 1, 3, []int64{1, 2}); err != nil {
		t.Fatalf("超管绑定角色失败: %v", err)
	}
	// 禁用用户绑定角色不影响启用超管集合
	if err := guard.EnsureCanBindRoles(ctx, 1, 4, []int64{2}); err != nil {
		t.Fatalf("禁用用户绑定不应影响启用超管: %v", err)
	}
	// 用户不存在时返回用户域错误
	if err := guard.EnsureCanBindRoles(ctx, 1, 99, []int64{2}); !errors.Is(err, usercontract.ErrUserNotFound) {
		t.Fatalf("不存在用户期望 ErrUserNotFound，实际 %v", err)
	}

	// 系统内没有启用超管时，任何角色变更都拒绝
	reader.userIDs = nil
	if err := guard.EnsureCanBindRoles(ctx, 9, 2, []int64{2}); !errors.Is(err, ErrLastSuperAdmin) {
		t.Fatalf("无启用超管期望 ErrLastSuperAdmin，实际 %v", err)
	}
}

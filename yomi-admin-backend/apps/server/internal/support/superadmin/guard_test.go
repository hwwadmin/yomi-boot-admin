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

// fakeSuperReader 是超管关系测试桩
type fakeSuperReader struct {
	userIDs []int64
}

func (f *fakeSuperReader) ListUserIDsWithEnabledSuperAdmin(context.Context) ([]int64, error) {
	return f.userIDs, nil
}

// superRole 返回启用态超管角色
func superRole() *rbacentity.Role {
	return &rbacentity.Role{ID: 1, Code: rbacenums.RoleCodeSuperAdmin, Status: rbacenums.StatusEnabled}
}

// viewerRole 返回普通启用态角色
func viewerRole() *rbacentity.Role {
	return &rbacentity.Role{ID: 2, Code: "viewer", Status: rbacenums.StatusEnabled}
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

func TestEnsureCanManageUser(t *testing.T) {
	ctx := context.Background()
	users := &fakeUserGetter{users: map[int64]*userentity.User{
		1: {ID: 1, Status: userenums.UserStatusEnabled},
		2: {ID: 2, Status: userenums.UserStatusEnabled},
		3: {ID: 3, Status: userenums.UserStatusDisabled},
		4: {ID: 4, Status: userenums.UserStatusEnabled},
	}}

	// 超管处置另一名启用超管：放行
	guard := New(users, &fakeSuperReader{userIDs: []int64{1, 2}}, &fakeLocker{})
	if err := guard.EnsureCanManageUser(ctx, 1, 2); err != nil {
		t.Fatalf("超管处置超管应放行: %v", err)
	}

	// 普通管理员处置启用超管：拒绝
	if err := guard.EnsureCanManageUser(ctx, 4, 1); !errors.Is(err, ErrManageSuperDenied) {
		t.Fatalf("非超管处置超管期望 ErrManageSuperDenied，实际 %v", err)
	}

	// 账号已停用的超管不再受保护，与角色绑定口径一致
	if err := guard.EnsureCanManageUser(ctx, 4, 3); err != nil {
		t.Fatalf("处置已停用超管账号不应受限: %v", err)
	}

	// 超管处置普通用户：放行
	if err := guard.EnsureCanManageUser(ctx, 1, 4); err != nil {
		t.Fatalf("超管处置普通用户应放行: %v", err)
	}

	// 唯一启用超管就是目标：拒绝
	guard = New(users, &fakeSuperReader{userIDs: []int64{1}}, &fakeLocker{})
	if err := guard.EnsureCanManageUser(ctx, 1, 1); !errors.Is(err, ErrLastSuperAdmin) {
		t.Fatalf("唯一超管期望 ErrLastSuperAdmin，实际 %v", err)
	}

	// 孤儿关系不计入启用超管
	guard = New(users, &fakeSuperReader{userIDs: []int64{1, 99}}, &fakeLocker{})
	if err := guard.EnsureCanManageUser(ctx, 1, 1); !errors.Is(err, ErrLastSuperAdmin) {
		t.Fatalf("孤儿关系不应计入: %v", err)
	}
}

func TestEnsureCanKickUser(t *testing.T) {
	ctx := context.Background()
	users := &fakeUserGetter{users: map[int64]*userentity.User{
		1: {ID: 1, Status: userenums.UserStatusEnabled},
		2: {ID: 2, Status: userenums.UserStatusEnabled},
		3: {ID: 3, Status: userenums.UserStatusDisabled},
		4: {ID: 4, Status: userenums.UserStatusEnabled},
	}}
	guard := New(users, &fakeSuperReader{userIDs: []int64{1, 2}}, &fakeLocker{})

	// 超管踢另一名启用超管：放行
	if err := guard.EnsureCanKickUser(ctx, 1, 2); err != nil {
		t.Fatalf("超管踢超管应放行: %v", err)
	}

	// 普通管理员踢启用超管：拒绝
	if err := guard.EnsureCanKickUser(ctx, 4, 1); !errors.Is(err, ErrKickSuperDenied) {
		t.Fatalf("非超管踢超管期望 ErrKickSuperDenied，实际 %v", err)
	}

	// 已停用超管账号不再受保护，与停用/删除口径一致
	if err := guard.EnsureCanKickUser(ctx, 4, 3); err != nil {
		t.Fatalf("踢已停用超管账号不应受限: %v", err)
	}

	// 超管踢普通用户：放行
	if err := guard.EnsureCanKickUser(ctx, 1, 4); err != nil {
		t.Fatalf("超管踢普通用户应放行: %v", err)
	}

	// 唯一启用超管踢自己：放行（踢下线不等于移除超管）
	guard = New(users, &fakeSuperReader{userIDs: []int64{1}}, &fakeLocker{})
	if err := guard.EnsureCanKickUser(ctx, 1, 1); err != nil {
		t.Fatalf("唯一超管踢自己应放行: %v", err)
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
	reader := &fakeSuperReader{userIDs: []int64{1}}
	guard := New(users, reader, &fakeLocker{})

	// 非超管授予超管
	if err := guard.EnsureCanBindRoles(ctx, 2, 3, []*rbacentity.Role{superRole()}); !errors.Is(err, ErrGrantSuperDenied) {
		t.Fatalf("非超管授予超管期望 ErrGrantSuperDenied，实际 %v", err)
	}
	// 非超管回收他人超管
	if err := guard.EnsureCanBindRoles(ctx, 2, 1, []*rbacentity.Role{viewerRole()}); !errors.Is(err, ErrGrantSuperDenied) {
		t.Fatalf("非超管回收超管期望 ErrGrantSuperDenied，实际 %v", err)
	}
	// 超管解绑自己
	if err := guard.EnsureCanBindRoles(ctx, 1, 1, []*rbacentity.Role{viewerRole()}); !errors.Is(err, ErrUnbindSelfSuper) {
		t.Fatalf("超管自解绑期望 ErrUnbindSelfSuper，实际 %v", err)
	}
	// 超管给他人绑定超管
	if err := guard.EnsureCanBindRoles(ctx, 1, 3, []*rbacentity.Role{superRole(), viewerRole()}); err != nil {
		t.Fatalf("超管绑定角色失败: %v", err)
	}
	// 禁用用户绑定角色不影响启用超管集合
	if err := guard.EnsureCanBindRoles(ctx, 1, 4, []*rbacentity.Role{viewerRole()}); err != nil {
		t.Fatalf("禁用用户绑定不应影响启用超管: %v", err)
	}
	// 用户不存在时返回用户域错误
	if err := guard.EnsureCanBindRoles(ctx, 1, 99, []*rbacentity.Role{viewerRole()}); !errors.Is(err, usercontract.ErrUserNotFound) {
		t.Fatalf("不存在用户期望 ErrUserNotFound，实际 %v", err)
	}

	// 系统内没有启用超管时，任何角色变更都拒绝
	reader.userIDs = nil
	if err := guard.EnsureCanBindRoles(ctx, 9, 2, []*rbacentity.Role{viewerRole()}); !errors.Is(err, ErrLastSuperAdmin) {
		t.Fatalf("无启用超管期望 ErrLastSuperAdmin，实际 %v", err)
	}
}

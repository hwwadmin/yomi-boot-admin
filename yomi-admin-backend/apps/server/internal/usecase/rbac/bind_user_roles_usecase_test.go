package rbac

import (
	"context"
	"testing"

	rbacentity "yomi-admin-backend/domains/rbac/pkg/shared/entity"
	rbacinput "yomi-admin-backend/domains/rbac/pkg/shared/input"
	rbaccontract "yomi-admin-backend/domains/rbac/service_contract"

	"gitee.com/hq_project/yomi-boot/component/core/errors"
	"gitee.com/hq_project/yomi-boot/component/core/shared/dto"
)

// fakeBindRolesAdmin 只实现角色解析与用户角色绑定，其余方法由嵌入接口兜底
type fakeBindRolesAdmin struct {
	rbaccontract.RbacAdminService
	binds      int
	got        *rbacinput.BindUserRolesInput
	resolves   int
	resolvedID []int64
	roles      []*rbacentity.Role
	resolveErr error
}

func (f *fakeBindRolesAdmin) ResolveBindingRoles(_ context.Context, ids []int64) ([]*rbacentity.Role, error) {
	f.resolves++
	f.resolvedID = ids
	if f.resolveErr != nil {
		return nil, f.resolveErr
	}
	return f.roles, nil
}

func (f *fakeBindRolesAdmin) BindUserRoles(_ context.Context, in *rbacinput.BindUserRolesInput) error {
	f.binds++
	f.got = in
	return nil
}

// fakeSuperAdminPolicy 记录锁、策略检查与收到的角色
type fakeSuperAdminPolicy struct {
	locks    int
	checks   int
	gotRoles []*rbacentity.Role
}

func (f *fakeSuperAdminPolicy) WithLock(ctx context.Context, fn func(context.Context) error) error {
	f.locks++
	return fn(ctx)
}

func (f *fakeSuperAdminPolicy) EnsureCanBindRoles(_ context.Context, _, _ int64, roles []*rbacentity.Role) error {
	f.checks++
	f.gotRoles = roles
	return nil
}

func TestBindUserRolesRunsUnderLockAndTx(t *testing.T) {
	resolved := []*rbacentity.Role{{ID: 1, Code: "viewer"}}
	admin := &fakeBindRolesAdmin{roles: resolved}
	guard := &fakeSuperAdminPolicy{}
	tx := &fakeTxExecutor{}
	uc := NewBindUserRolesUsecase(admin, guard, tx)

	err := uc.Execute(context.Background(), &BindUserRolesInput{
		UserID:     dto.JSONInt64(5),
		RoleIDs:    dto.NewJSONInt64Slice([]int64{1, 2}),
		OperatorID: 9,
	})
	if err != nil {
		t.Fatalf("绑定角色失败: %v", err)
	}
	if guard.locks != 1 || guard.checks != 1 || tx.runs != 1 || admin.binds != 1 || admin.resolves != 1 {
		t.Fatalf("调用次数不符: locks=%d checks=%d tx=%d binds=%d resolves=%d",
			guard.locks, guard.checks, tx.runs, admin.binds, admin.resolves)
	}
	if len(admin.resolvedID) != 2 || admin.resolvedID[0] != 1 || admin.resolvedID[1] != 2 {
		t.Fatalf("解析入参不符: %v", admin.resolvedID)
	}
	if len(guard.gotRoles) != 1 || guard.gotRoles[0] != resolved[0] {
		t.Fatalf("策略应收到解析后的角色: %+v", guard.gotRoles)
	}
	if admin.got == nil || admin.got.UserID != 5 || admin.got.OperatorID != 9 {
		t.Fatalf("领域入参不符: %+v", admin.got)
	}
	if ids := admin.got.RoleIDs; len(ids) != 2 || ids[0] != 1 || ids[1] != 2 {
		t.Fatalf("角色 ID 不符: %v", ids)
	}
}

func TestBindUserRolesRejectsInvalidUserID(t *testing.T) {
	admin := &fakeBindRolesAdmin{}
	guard := &fakeSuperAdminPolicy{}
	tx := &fakeTxExecutor{}
	uc := NewBindUserRolesUsecase(admin, guard, tx)

	err := uc.Execute(context.Background(), &BindUserRolesInput{})
	if !errors.Is(err, rbaccontract.ErrInvalidInput) {
		t.Fatalf("期望 ErrInvalidInput，实际 %v", err)
	}
	if guard.locks != 0 || tx.runs != 0 || admin.binds != 0 || admin.resolves != 0 {
		t.Fatalf("非法入参不应触达锁/事务/领域: locks=%d tx=%d binds=%d resolves=%d",
			guard.locks, tx.runs, admin.binds, admin.resolves)
	}
}

// TestBindUserRolesStopsWhenRoleResolutionFails 校验角色校验失败时不再进入超管策略与实际绑定
func TestBindUserRolesStopsWhenRoleResolutionFails(t *testing.T) {
	admin := &fakeBindRolesAdmin{resolveErr: errors.Sentinel(rbaccontract.ErrRoleNotFound)}
	guard := &fakeSuperAdminPolicy{}
	tx := &fakeTxExecutor{}
	uc := NewBindUserRolesUsecase(admin, guard, tx)

	err := uc.Execute(context.Background(), &BindUserRolesInput{
		UserID:     dto.JSONInt64(5),
		RoleIDs:    dto.NewJSONInt64Slice([]int64{99}),
		OperatorID: 9,
	})
	if !errors.Is(err, rbaccontract.ErrRoleNotFound) {
		t.Fatalf("期望 ErrRoleNotFound，实际 %v", err)
	}
	if guard.checks != 0 || admin.binds != 0 {
		t.Fatalf("角色校验失败后不应继续: checks=%d binds=%d", guard.checks, admin.binds)
	}
	if tx.runs != 1 {
		t.Fatalf("事务应已开启并由错误回滚: %d", tx.runs)
	}
}

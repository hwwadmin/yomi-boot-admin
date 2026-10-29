package rbac

import (
	"context"
	"testing"

	rbacinput "yomi-admin-backend/domains/rbac/pkg/shared/input"
	rbaccontract "yomi-admin-backend/domains/rbac/service_contract"

	"gitee.com/hq_project/yomi-boot/component/core/errors"
	"gitee.com/hq_project/yomi-boot/component/core/shared/dto"
)

// fakeBindRolesAdmin 只实现用户角色绑定，其余方法由嵌入接口兜底
type fakeBindRolesAdmin struct {
	rbaccontract.RbacAdminService
	calls int
	got   *rbacinput.BindUserRolesInput
}

func (f *fakeBindRolesAdmin) BindUserRoles(_ context.Context, in *rbacinput.BindUserRolesInput) error {
	f.calls++
	f.got = in
	return nil
}

// fakeSuperAdminPolicy 记录锁和策略检查的调用次数
type fakeSuperAdminPolicy struct {
	locks  int
	checks int
}

func (f *fakeSuperAdminPolicy) WithLock(ctx context.Context, fn func(context.Context) error) error {
	f.locks++
	return fn(ctx)
}

func (f *fakeSuperAdminPolicy) EnsureCanBindRoles(context.Context, int64, int64, []int64) error {
	f.checks++
	return nil
}

func TestBindUserRolesRunsUnderLockAndTx(t *testing.T) {
	admin := &fakeBindRolesAdmin{}
	guard := &fakeSuperAdminPolicy{}
	tx := &fakeTxExecutor{}
	uc := NewBindUserRolesUsecase(admin, guard, tx)

	err := uc.Execute(context.Background(), 5, &BindUserRolesInput{
		RoleIDs:    dto.NewJSONInt64Slice([]int64{1, 2}),
		OperatorID: 9,
	})
	if err != nil {
		t.Fatalf("绑定角色失败: %v", err)
	}
	if guard.locks != 1 || guard.checks != 1 || tx.runs != 1 || admin.calls != 1 {
		t.Fatalf("调用次数不符: locks=%d checks=%d tx=%d admin=%d", guard.locks, guard.checks, tx.runs, admin.calls)
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

	err := uc.Execute(context.Background(), 0, &BindUserRolesInput{})
	if !errors.Is(err, rbaccontract.ErrInvalidInput) {
		t.Fatalf("期望 ErrInvalidInput，实际 %v", err)
	}
	if guard.locks != 0 || tx.runs != 0 || admin.calls != 0 {
		t.Fatalf("非法入参不应触达锁/事务/领域: locks=%d tx=%d admin=%d", guard.locks, tx.runs, admin.calls)
	}
}

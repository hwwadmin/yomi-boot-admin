package rbac

import (
	"context"
	"testing"
	"time"

	rbacinput "yomi-admin-backend/domains/rbac/pkg/shared/input"
	rbaccontract "yomi-admin-backend/domains/rbac/service_contract"

	"gitee.com/hq_project/yomi-boot/component/core/errors"
	"gitee.com/hq_project/yomi-boot/component/core/shared/dto"
)

// fakeRbacAdmin 只实现角色权限绑定，其余方法由嵌入接口兜底
type fakeRbacAdmin struct {
	rbaccontract.RbacAdminService
	calls int
	got   *rbacinput.BindRolePermissionsInput
	err   error
}

func (f *fakeRbacAdmin) BindRolePermissions(_ context.Context, in *rbacinput.BindRolePermissionsInput) error {
	f.calls++
	f.got = in
	return f.err
}

// fakeTxExecutor 在当前协程内执行事务闭包并记录调用次数
type fakeTxExecutor struct {
	runs int
}

func (f *fakeTxExecutor) Run(ctx context.Context, fn func(ctx context.Context) error) error {
	f.runs++
	return fn(ctx)
}

func (f *fakeTxExecutor) InTx(context.Context) bool { return false }

func (f *fakeTxExecutor) RunAutonomous(ctx context.Context, _ time.Duration, fn func(ctx context.Context) error) error {
	return fn(ctx)
}

func TestBindRolePermissionsRunsInTransaction(t *testing.T) {
	rbac := &fakeRbacAdmin{}
	tx := &fakeTxExecutor{}
	uc := NewBindRolePermissionsUsecase(rbac, tx)

	err := uc.Execute(context.Background(), 7, &BindRolePermissionsInput{
		PermissionIDs: dto.NewJSONInt64Slice([]int64{10, 11}),
		OperatorID:    9,
	})
	if err != nil {
		t.Fatalf("绑定权限失败: %v", err)
	}
	if tx.runs != 1 {
		t.Fatalf("事务执行次数 = %d, want 1", tx.runs)
	}
	if rbac.calls != 1 || rbac.got == nil {
		t.Fatalf("领域调用次数 = %d", rbac.calls)
	}
	if rbac.got.RoleID != 7 || rbac.got.OperatorID != 9 {
		t.Fatalf("领域入参不符: %+v", rbac.got)
	}
	if ids := rbac.got.PermissionIDs; len(ids) != 2 || ids[0] != 10 || ids[1] != 11 {
		t.Fatalf("权限 ID 不符: %v", ids)
	}
}

func TestBindRolePermissionsPropagatesDomainError(t *testing.T) {
	rbac := &fakeRbacAdmin{err: errors.Sentinel(rbaccontract.ErrRoleNotFound)}
	tx := &fakeTxExecutor{}
	uc := NewBindRolePermissionsUsecase(rbac, tx)

	err := uc.Execute(context.Background(), 7, &BindRolePermissionsInput{})
	if !errors.Is(err, rbaccontract.ErrRoleNotFound) {
		t.Fatalf("期望 ErrRoleNotFound，实际 %v", err)
	}
	if tx.runs != 1 {
		t.Fatalf("事务执行次数 = %d, want 1", tx.runs)
	}
}

func TestBindRolePermissionsRejectsInvalidRoleID(t *testing.T) {
	rbac := &fakeRbacAdmin{}
	tx := &fakeTxExecutor{}
	uc := NewBindRolePermissionsUsecase(rbac, tx)

	err := uc.Execute(context.Background(), 0, &BindRolePermissionsInput{})
	if !errors.Is(err, rbaccontract.ErrInvalidInput) {
		t.Fatalf("期望 ErrInvalidInput，实际 %v", err)
	}
	if tx.runs != 0 || rbac.calls != 0 {
		t.Fatalf("非法入参不应触达事务或领域: tx=%d rbac=%d", tx.runs, rbac.calls)
	}
}

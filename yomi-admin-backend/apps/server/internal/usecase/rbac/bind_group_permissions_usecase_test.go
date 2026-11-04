package rbac

import (
	"context"
	"testing"

	rbacinput "yomi-admin-backend/domains/rbac/pkg/shared/input"
	rbaccontract "yomi-admin-backend/domains/rbac/service_contract"

	"github.com/hwwadmin/yomi-boot/component/core/errors"
	"github.com/hwwadmin/yomi-boot/component/core/shared/dto"
)

// fakeGroupAdmin 只实现权限组权限绑定，其余方法由嵌入接口兜底
type fakeGroupAdmin struct {
	rbaccontract.RbacAdminService
	calls int
	got   *rbacinput.BindGroupPermissionsInput
	err   error
}

func (f *fakeGroupAdmin) BindGroupPermissions(_ context.Context, in *rbacinput.BindGroupPermissionsInput) error {
	f.calls++
	f.got = in
	return f.err
}

func TestBindGroupPermissionsRunsInTransaction(t *testing.T) {
	admin := &fakeGroupAdmin{}
	tx := &fakeTxExecutor{}
	uc := NewBindGroupPermissionsUsecase(admin, tx)

	err := uc.Execute(context.Background(), &BindGroupPermissionsInput{
		GroupID:       dto.JSONInt64(20),
		PermissionIDs: dto.NewJSONInt64Slice([]int64{10, 11}),
		OperatorID:    9,
	})
	if err != nil {
		t.Fatalf("绑定权限失败: %v", err)
	}
	if tx.runs != 1 || admin.calls != 1 || admin.got == nil {
		t.Fatalf("调用次数不符: tx=%d admin=%d", tx.runs, admin.calls)
	}
	if admin.got.GroupID != 20 || admin.got.OperatorID != 9 {
		t.Fatalf("领域入参不符: %+v", admin.got)
	}
	if ids := admin.got.PermissionIDs; len(ids) != 2 || ids[0] != 10 || ids[1] != 11 {
		t.Fatalf("权限 ID 不符: %v", ids)
	}
}

func TestBindGroupPermissionsRejectsInvalidGroupID(t *testing.T) {
	admin := &fakeGroupAdmin{}
	tx := &fakeTxExecutor{}
	uc := NewBindGroupPermissionsUsecase(admin, tx)

	err := uc.Execute(context.Background(), &BindGroupPermissionsInput{})
	if !errors.Is(err, rbaccontract.ErrInvalidInput) {
		t.Fatalf("期望 ErrInvalidInput，实际 %v", err)
	}
	if tx.runs != 0 || admin.calls != 0 {
		t.Fatalf("非法入参不应触达事务或领域: tx=%d admin=%d", tx.runs, admin.calls)
	}
}

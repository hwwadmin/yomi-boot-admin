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

// fakeDeleteAdmin 只实现删除与创建角色，其余方法由嵌入接口兜底
type fakeDeleteAdmin struct {
	rbaccontract.RbacAdminService
	deleteErr error
	createErr error
}

func (f *fakeDeleteAdmin) DeleteRole(context.Context, *rbacinput.DeleteRoleInput) error {
	return f.deleteErr
}

func (f *fakeDeleteAdmin) DeletePermission(context.Context, *rbacinput.DeletePermissionInput) error {
	return f.deleteErr
}

func (f *fakeDeleteAdmin) DeletePermissionGroup(context.Context, *rbacinput.DeletePermissionGroupInput) error {
	return f.deleteErr
}

func (f *fakeDeleteAdmin) CreateRole(context.Context, *rbacinput.CreateRoleInput) (*rbacentity.Role, error) {
	if f.createErr != nil {
		return nil, f.createErr
	}
	return &rbacentity.Role{ID: 9, Code: "ops", Name: "运维"}, nil
}

func TestDeleteUsecasesTreatNotFoundAsSuccess(t *testing.T) {
	cases := []struct {
		name string
		err  error
		run  func(admin *fakeDeleteAdmin) error
	}{
		{
			name: "角色",
			err:  rbaccontract.ErrRoleNotFound,
			run: func(admin *fakeDeleteAdmin) error {
				uc := NewDeleteRoleUsecase(admin, &fakeTxExecutor{})
				return uc.Execute(context.Background(), &DeleteRoleInput{ID: dto.JSONInt64(7), OperatorID: 1})
			},
		},
		{
			name: "权限",
			err:  rbaccontract.ErrPermissionNotFound,
			run: func(admin *fakeDeleteAdmin) error {
				uc := NewDeletePermissionUsecase(admin, &fakeTxExecutor{})
				return uc.Execute(context.Background(), &DeletePermissionInput{ID: dto.JSONInt64(7), OperatorID: 1})
			},
		},
		{
			name: "权限组",
			err:  rbaccontract.ErrPermissionGroupNotFound,
			run: func(admin *fakeDeleteAdmin) error {
				uc := NewDeletePermissionGroupUsecase(admin, &fakeTxExecutor{})
				return uc.Execute(context.Background(), &DeletePermissionGroupInput{ID: dto.JSONInt64(7), OperatorID: 1})
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name+"重复删除按成功", func(t *testing.T) {
			if err := tc.run(&fakeDeleteAdmin{deleteErr: tc.err}); err != nil {
				t.Fatalf("重复删除应返回 nil, got %v", err)
			}
		})
	}
}

func TestDeleteRoleKeepsOtherErrors(t *testing.T) {
	bound := rbaccontract.ErrRoleBoundToUsers
	uc := NewDeleteRoleUsecase(&fakeDeleteAdmin{deleteErr: bound}, &fakeTxExecutor{})
	err := uc.Execute(context.Background(), &DeleteRoleInput{ID: dto.JSONInt64(7), OperatorID: 1})
	if !errors.Is(err, bound) {
		t.Fatalf("角色仍被绑定时应保留原错误, got %v", err)
	}
}

func TestCreateRoleWithoutIdempotencyKeyRunsOnce(t *testing.T) {
	admin := &fakeDeleteAdmin{}
	tx := &fakeTxExecutor{}
	uc := NewCreateRoleUsecase(admin, nil, tx)

	item, err := uc.Execute(context.Background(), &CreateRoleInput{Code: "ops", Name: "运维"})
	if err != nil {
		t.Fatalf("创建角色失败: %v", err)
	}
	if item == nil || item.ID.Int64() != 9 {
		t.Fatalf("创建结果不符合预期: %+v", item)
	}
	if tx.runs != 1 {
		t.Fatalf("事务执行次数 = %d, want 1", tx.runs)
	}
}

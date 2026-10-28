package service

import (
	"context"
	"testing"
	"yomi-admin-backend/domains/rbac/pkg/shared/entity"
	"yomi-admin-backend/domains/rbac/pkg/shared/enums"
	"yomi-admin-backend/domains/rbac/pkg/shared/input"
	"yomi-admin-backend/domains/rbac/service_contract"

	"gitee.com/hq_project/yomi-boot/component/core/errors"
)

// seqIDGenerator 是测试用递增 ID 生成器
type seqIDGenerator struct {
	next int64
}

func (g *seqIDGenerator) NextID() (int64, error) {
	g.next++
	return g.next, nil
}

func newTestAdminService(repo *fakeRepo) *RbacAdminService {
	return NewRbacAdminService(repo, &seqIDGenerator{next: 1000})
}

func TestCreateRoleRules(t *testing.T) {
	ctx := context.Background()
	repo := newFakeRepo()
	repo.roles[1] = &entity.Role{ID: 1, Code: "existing", Name: "已有角色", Status: enums.StatusEnabled}
	svc := newTestAdminService(repo)

	if _, err := svc.CreateRole(ctx, &input.CreateRoleInput{Code: "existing", Name: "重复"}); !errors.Is(err, service_contract.ErrRoleCodeExists) {
		t.Fatalf("重复编码期望 ErrRoleCodeExists，实际 %v", err)
	}
	if _, err := svc.CreateRole(ctx, &input.CreateRoleInput{Code: enums.RoleCodeSuperAdmin, Name: "超管"}); !errors.Is(err, service_contract.ErrBuiltinRoleImmutable) {
		t.Fatalf("创建内置超管期望 ErrBuiltinRoleImmutable，实际 %v", err)
	}

	role, err := svc.CreateRole(ctx, &input.CreateRoleInput{Code: "viewer", Name: "只读", OperatorID: 1})
	if err != nil {
		t.Fatalf("创建角色失败: %v", err)
	}
	if role.ID == 0 || role.Status != enums.StatusEnabled || role.Builtin {
		t.Fatalf("创建结果不符合预期: %+v", role)
	}
	if _, ok := repo.roles[role.ID]; !ok {
		t.Fatal("角色未写入仓储")
	}
}

func TestUpdateBuiltinRoleRejectsDisable(t *testing.T) {
	ctx := context.Background()
	repo := newFakeRepo()
	repo.roles[1] = &entity.Role{ID: 1, Code: enums.RoleCodeSuperAdmin, Name: "超管", Builtin: true, Status: enums.StatusEnabled}
	svc := newTestAdminService(repo)

	err := svc.UpdateRole(ctx, &input.UpdateRoleInput{ID: 1, Name: "超管", Status: enums.StatusDisabled, OperatorID: 1})
	if !errors.Is(err, service_contract.ErrBuiltinRoleImmutable) {
		t.Fatalf("禁用内置角色期望 ErrBuiltinRoleImmutable，实际 %v", err)
	}

	repo.roles[2] = &entity.Role{ID: 2, Code: "viewer", Name: "只读", Status: enums.StatusEnabled}
	if err := svc.UpdateRole(ctx, &input.UpdateRoleInput{ID: 2, Name: "只读角色", Status: enums.StatusDisabled, OperatorID: 1}); err != nil {
		t.Fatalf("更新普通角色失败: %v", err)
	}
	if repo.roles[2].Status != enums.StatusDisabled || repo.roles[2].Name != "只读角色" {
		t.Fatalf("更新未生效: %+v", repo.roles[2])
	}
}

func TestBindRolePermissionsRules(t *testing.T) {
	ctx := context.Background()
	repo := newFakeRepo()
	repo.roles[1] = &entity.Role{ID: 1, Code: enums.RoleCodeSuperAdmin, Builtin: true, Status: enums.StatusEnabled}
	repo.roles[2] = &entity.Role{ID: 2, Code: "viewer", Status: enums.StatusEnabled}
	repo.permissions[10] = &entity.Permission{ID: 10, Code: "system:user:list", Type: enums.PermissionTypeAPI, Status: enums.StatusEnabled}
	repo.permissions[11] = &entity.Permission{ID: 11, Code: "system:user:delete", Type: enums.PermissionTypeAPI, Status: enums.StatusDisabled}
	svc := newTestAdminService(repo)

	if err := svc.BindRolePermissions(ctx, &input.BindRolePermissionsInput{RoleID: 1, PermissionIDs: []int64{10}}); !errors.Is(err, service_contract.ErrBuiltinRoleImmutable) {
		t.Fatalf("内置角色绑定权限期望 ErrBuiltinRoleImmutable，实际 %v", err)
	}
	if err := svc.BindRolePermissions(ctx, &input.BindRolePermissionsInput{RoleID: 2, PermissionIDs: []int64{10, 99}}); !errors.Is(err, service_contract.ErrPermissionNotFound) {
		t.Fatalf("权限不存在期望 ErrPermissionNotFound，实际 %v", err)
	}
	if err := svc.BindRolePermissions(ctx, &input.BindRolePermissionsInput{RoleID: 2, PermissionIDs: []int64{11}}); !errors.Is(err, service_contract.ErrRelatedDisabled) {
		t.Fatalf("绑定禁用权限期望 ErrRelatedDisabled，实际 %v", err)
	}
	if err := svc.BindRolePermissions(ctx, &input.BindRolePermissionsInput{RoleID: 2, PermissionIDs: []int64{10, 10}, OperatorID: 1}); err != nil {
		t.Fatalf("绑定权限失败: %v", err)
	}
	if ids := repo.rolePermissions[2]; len(ids) != 1 || ids[0] != 10 {
		t.Fatalf("权限绑定未去重或未写入: %v", ids)
	}
}

func TestBindUserRolesSuperRules(t *testing.T) {
	ctx := context.Background()
	repo := newFakeRepo()
	repo.roles[1] = &entity.Role{ID: 1, Code: enums.RoleCodeSuperAdmin, Builtin: true, Status: enums.StatusEnabled}
	repo.roles[2] = &entity.Role{ID: 2, Code: "viewer", Status: enums.StatusEnabled}
	repo.userRoles[1] = []int64{1}
	repo.userRoles[2] = []int64{2}
	repo.userRoles[4] = []int64{1}
	svc := newTestAdminService(repo)

	// 非超管授予超管
	err := svc.BindUserRoles(ctx, &input.BindUserRolesInput{UserID: 3, RoleIDs: []int64{1}, OperatorID: 2})
	if !errors.Is(err, service_contract.ErrGrantSuperDenied) {
		t.Fatalf("非超管授予超管期望 ErrGrantSuperDenied，实际 %v", err)
	}
	// 非超管回收他人超管
	err = svc.BindUserRoles(ctx, &input.BindUserRolesInput{UserID: 4, RoleIDs: []int64{2}, OperatorID: 2})
	if !errors.Is(err, service_contract.ErrGrantSuperDenied) {
		t.Fatalf("非超管回收超管期望 ErrGrantSuperDenied，实际 %v", err)
	}
	// 超管解绑自己
	err = svc.BindUserRoles(ctx, &input.BindUserRolesInput{UserID: 1, RoleIDs: []int64{2}, OperatorID: 1})
	if !errors.Is(err, service_contract.ErrUnbindSelfSuper) {
		t.Fatalf("超管自解绑期望 ErrUnbindSelfSuper，实际 %v", err)
	}
	// 超管给他人绑定超管
	if err := svc.BindUserRoles(ctx, &input.BindUserRolesInput{UserID: 3, RoleIDs: []int64{1, 2, 2}, OperatorID: 1}); err != nil {
		t.Fatalf("超管绑定角色失败: %v", err)
	}
	if ids := repo.userRoles[3]; len(ids) != 2 || ids[0] != 1 || ids[1] != 2 {
		t.Fatalf("用户角色未去重或未写入: %v", ids)
	}
}

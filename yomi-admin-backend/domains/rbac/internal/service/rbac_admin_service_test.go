package service

import (
	"context"
	"strings"
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

	invalid := []struct {
		name string
		in   *input.CreateRoleInput
	}{
		{name: "编码含大写", in: &input.CreateRoleInput{Code: "Viewer", Name: "只读"}},
		{name: "编码过短", in: &input.CreateRoleInput{Code: "ab", Name: "只读"}},
		{name: "编码含非法字符", in: &input.CreateRoleInput{Code: "view-er", Name: "只读"}},
		{name: "名称为空", in: &input.CreateRoleInput{Code: "viewer", Name: "  "}},
		{name: "名称超长", in: &input.CreateRoleInput{Code: "viewer", Name: strings.Repeat("名", 33)}},
		{name: "备注超长", in: &input.CreateRoleInput{Code: "viewer", Name: "只读", Remark: strings.Repeat("字", 256)}},
	}
	for _, tc := range invalid {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := svc.CreateRole(ctx, tc.in); !errors.Is(err, service_contract.ErrInvalidInput) {
				t.Fatalf("期望 ErrInvalidInput，实际 %v", err)
			}
		})
	}

	role, err := svc.CreateRole(ctx, &input.CreateRoleInput{Code: " viewer ", Name: " 只读 ", Remark: " 备注 ", OperatorID: 1})
	if err != nil {
		t.Fatalf("创建角色失败: %v", err)
	}
	if role.ID == 0 || role.Status != enums.StatusEnabled || role.Builtin {
		t.Fatalf("创建结果不符合预期: %+v", role)
	}
	if role.Code != "viewer" || role.Name != "只读" || role.Remark != "备注" {
		t.Fatalf("创建结果未去除首尾空白: %+v", role)
	}
	if _, ok := repo.roles[role.ID]; !ok {
		t.Fatal("角色未写入仓储")
	}
}

func TestUpdateRoleRules(t *testing.T) {
	ctx := context.Background()
	repo := newFakeRepo()
	repo.roles[1] = &entity.Role{ID: 1, Code: enums.RoleCodeSuperAdmin, Name: "超管", Builtin: true, Status: enums.StatusEnabled}
	repo.roles[2] = &entity.Role{ID: 2, Code: "viewer", Name: "只读", Status: enums.StatusEnabled}
	svc := newTestAdminService(repo)

	if err := svc.UpdateRole(ctx, &input.UpdateRoleInput{ID: 1, Name: "超管", Status: enums.StatusDisabled, OperatorID: 1}); !errors.Is(err, service_contract.ErrBuiltinRoleImmutable) {
		t.Fatalf("禁用内置角色期望 ErrBuiltinRoleImmutable，实际 %v", err)
	}

	invalid := []struct {
		name string
		in   *input.UpdateRoleInput
	}{
		{name: "名称为空", in: &input.UpdateRoleInput{ID: 2, Name: " ", Status: enums.StatusEnabled}},
		{name: "非法状态", in: &input.UpdateRoleInput{ID: 2, Name: "只读", Status: enums.Status("UNKNOWN")}},
		{name: "备注超长", in: &input.UpdateRoleInput{ID: 2, Name: "只读", Status: enums.StatusEnabled, Remark: strings.Repeat("字", 256)}},
	}
	for _, tc := range invalid {
		t.Run(tc.name, func(t *testing.T) {
			if err := svc.UpdateRole(ctx, tc.in); !errors.Is(err, service_contract.ErrInvalidInput) {
				t.Fatalf("期望 ErrInvalidInput，实际 %v", err)
			}
		})
	}

	if err := svc.UpdateRole(ctx, &input.UpdateRoleInput{ID: 0, Name: "只读", Status: enums.StatusEnabled}); !errors.Is(err, service_contract.ErrRoleNotFound) {
		t.Fatalf("非法 ID 期望 ErrRoleNotFound，实际 %v", err)
	}
	if err := svc.UpdateRole(ctx, &input.UpdateRoleInput{ID: 2, Name: " 只读角色 ", Status: enums.StatusDisabled, OperatorID: 1}); err != nil {
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
	if err := svc.BindRolePermissions(ctx, &input.BindRolePermissionsInput{RoleID: 2, PermissionIDs: []int64{10, 0}}); !errors.Is(err, service_contract.ErrInvalidInput) {
		t.Fatalf("非法权限 ID 期望 ErrInvalidInput，实际 %v", err)
	}
	if err := svc.BindRolePermissions(ctx, &input.BindRolePermissionsInput{RoleID: 2, PermissionIDs: []int64{10, 10}, OperatorID: 1}); err != nil {
		t.Fatalf("绑定权限失败: %v", err)
	}
	if ids := repo.rolePermissions[2]; len(ids) != 1 || ids[0] != 10 {
		t.Fatalf("权限绑定未去重或未写入: %v", ids)
	}
}

func TestBindUserRolesRules(t *testing.T) {
	ctx := context.Background()
	repo := newFakeRepo()
	repo.roles[1] = &entity.Role{ID: 1, Code: enums.RoleCodeSuperAdmin, Builtin: true, Status: enums.StatusEnabled}
	repo.roles[2] = &entity.Role{ID: 2, Code: "viewer", Status: enums.StatusEnabled}
	repo.roles[3] = &entity.Role{ID: 3, Code: "auditor", Status: enums.StatusDisabled}
	svc := newTestAdminService(repo)

	if err := svc.BindUserRoles(ctx, &input.BindUserRolesInput{UserID: 0, RoleIDs: []int64{2}}); !errors.Is(err, service_contract.ErrInvalidInput) {
		t.Fatalf("非法用户 ID 期望 ErrInvalidInput，实际 %v", err)
	}
	if err := svc.BindUserRoles(ctx, &input.BindUserRolesInput{UserID: 1, RoleIDs: []int64{2, 99}}); !errors.Is(err, service_contract.ErrRoleNotFound) {
		t.Fatalf("角色不存在期望 ErrRoleNotFound，实际 %v", err)
	}
	if err := svc.BindUserRoles(ctx, &input.BindUserRolesInput{UserID: 1, RoleIDs: []int64{3}}); !errors.Is(err, service_contract.ErrRelatedDisabled) {
		t.Fatalf("绑定禁用角色期望 ErrRelatedDisabled，实际 %v", err)
	}
	if err := svc.BindUserRoles(ctx, &input.BindUserRolesInput{UserID: 1, RoleIDs: []int64{0}}); !errors.Is(err, service_contract.ErrInvalidInput) {
		t.Fatalf("非法角色 ID 期望 ErrInvalidInput，实际 %v", err)
	}
	if err := svc.BindUserRoles(ctx, &input.BindUserRolesInput{UserID: 1, RoleIDs: []int64{1, 2, 2}, OperatorID: 1}); err != nil {
		t.Fatalf("绑定角色失败: %v", err)
	}
	if ids := repo.userRoles[1]; len(ids) != 2 || ids[0] != 1 || ids[1] != 2 {
		t.Fatalf("用户角色未去重或未写入: %v", ids)
	}
}

func TestQueryAndStatusValidation(t *testing.T) {
	ctx := context.Background()
	repo := newFakeRepo()
	repo.permissions[10] = &entity.Permission{ID: 10, Status: enums.StatusEnabled}
	svc := newTestAdminService(repo)

	if _, err := svc.PageRoles(ctx, &input.PageRolesInput{Status: enums.Status("UNKNOWN")}); !errors.Is(err, service_contract.ErrInvalidInput) {
		t.Fatalf("非法角色状态期望 ErrInvalidInput，实际 %v", err)
	}
	if _, err := svc.PagePermissions(ctx, &input.PagePermissionsInput{Type: enums.PermissionType("UNKNOWN")}); !errors.Is(err, service_contract.ErrInvalidInput) {
		t.Fatalf("非法权限类型期望 ErrInvalidInput，实际 %v", err)
	}
	if _, err := svc.PagePermissions(ctx, &input.PagePermissionsInput{Status: enums.Status("UNKNOWN")}); !errors.Is(err, service_contract.ErrInvalidInput) {
		t.Fatalf("非法权限状态期望 ErrInvalidInput，实际 %v", err)
	}
	if err := svc.SetPermissionStatus(ctx, &input.SetPermissionStatusInput{ID: 10, Status: enums.Status("UNKNOWN")}); !errors.Is(err, service_contract.ErrInvalidInput) {
		t.Fatalf("非法启停状态期望 ErrInvalidInput，实际 %v", err)
	}
	if _, err := svc.ListUserRoles(ctx, 0); !errors.Is(err, service_contract.ErrInvalidInput) {
		t.Fatalf("非法用户 ID 期望 ErrInvalidInput，实际 %v", err)
	}
	if _, err := svc.ListRolesByIDs(ctx, []int64{0}); !errors.Is(err, service_contract.ErrInvalidInput) {
		t.Fatalf("非法角色 ID 期望 ErrInvalidInput，实际 %v", err)
	}
}

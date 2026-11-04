package service

import (
	"context"
	"strings"
	"testing"
	"yomi-admin-backend/domains/rbac/internal/repo_contract"
	"yomi-admin-backend/domains/rbac/pkg/shared/entity"
	"yomi-admin-backend/domains/rbac/pkg/shared/enums"
	"yomi-admin-backend/domains/rbac/pkg/shared/input"
	"yomi-admin-backend/domains/rbac/service_contract"

	"github.com/hwwadmin/yomi-boot/component/core/errors"
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

	invalid := []struct {
		name string
		in   *input.UpdateRoleInput
	}{
		{name: "名称为空", in: &input.UpdateRoleInput{ID: 2, Name: " "}},
		{name: "备注超长", in: &input.UpdateRoleInput{ID: 2, Name: "只读", Remark: strings.Repeat("字", 256)}},
	}
	for _, tc := range invalid {
		t.Run(tc.name, func(t *testing.T) {
			if err := svc.UpdateRole(ctx, tc.in); !errors.Is(err, service_contract.ErrInvalidInput) {
				t.Fatalf("期望 ErrInvalidInput，实际 %v", err)
			}
		})
	}

	if err := svc.UpdateRole(ctx, &input.UpdateRoleInput{ID: 0, Name: "只读"}); !errors.Is(err, service_contract.ErrRoleNotFound) {
		t.Fatalf("非法 ID 期望 ErrRoleNotFound，实际 %v", err)
	}
	if err := svc.UpdateRole(ctx, &input.UpdateRoleInput{ID: 2, Name: " 只读角色 ", Remark: "备注", OperatorID: 1}); err != nil {
		t.Fatalf("更新普通角色失败: %v", err)
	}
	if repo.roles[2].Name != "只读角色" || repo.roles[2].Remark != "备注" {
		t.Fatalf("更新未生效: %+v", repo.roles[2])
	}
	if repo.roles[2].Status != enums.StatusEnabled {
		t.Fatalf("更新资料不应改变启停状态: %+v", repo.roles[2])
	}
	if err := svc.UpdateRole(ctx, &input.UpdateRoleInput{ID: 1, Name: "超级管理员", OperatorID: 1}); err != nil {
		t.Fatalf("内置角色应允许更新资料: %v", err)
	}
}

func TestSetRoleStatusRules(t *testing.T) {
	ctx := context.Background()
	repo := newFakeRepo()
	repo.roles[1] = &entity.Role{ID: 1, Code: enums.RoleCodeSuperAdmin, Name: "超管", Builtin: true, Status: enums.StatusEnabled}
	repo.roles[2] = &entity.Role{ID: 2, Code: "viewer", Name: "只读", Status: enums.StatusEnabled}
	svc := newTestAdminService(repo)

	if err := svc.SetRoleStatus(ctx, &input.SetRoleStatusInput{ID: 0, Status: enums.StatusEnabled}); !errors.Is(err, service_contract.ErrRoleNotFound) {
		t.Fatalf("非法 ID 期望 ErrRoleNotFound，实际 %v", err)
	}
	if err := svc.SetRoleStatus(ctx, &input.SetRoleStatusInput{ID: 2, Status: enums.Status("UNKNOWN")}); !errors.Is(err, service_contract.ErrInvalidInput) {
		t.Fatalf("非法状态期望 ErrInvalidInput，实际 %v", err)
	}
	if err := svc.SetRoleStatus(ctx, &input.SetRoleStatusInput{ID: 99, Status: enums.StatusEnabled}); !errors.Is(err, service_contract.ErrRoleNotFound) {
		t.Fatalf("角色不存在期望 ErrRoleNotFound，实际 %v", err)
	}
	if err := svc.SetRoleStatus(ctx, &input.SetRoleStatusInput{ID: 1, Status: enums.StatusDisabled}); !errors.Is(err, service_contract.ErrBuiltinRoleImmutable) {
		t.Fatalf("禁用内置角色期望 ErrBuiltinRoleImmutable，实际 %v", err)
	}
	if err := svc.SetRoleStatus(ctx, &input.SetRoleStatusInput{ID: 2, Status: enums.StatusDisabled, OperatorID: 1}); err != nil {
		t.Fatalf("设置角色启停失败: %v", err)
	}
	if repo.roles[2].Status != enums.StatusDisabled {
		t.Fatalf("角色启停未生效: %+v", repo.roles[2])
	}
}

func TestDeleteRoleRules(t *testing.T) {
	ctx := context.Background()
	repo := newFakeRepo()
	repo.roles[1] = &entity.Role{ID: 1, Code: enums.RoleCodeSuperAdmin, Name: "超管", Builtin: true, Status: enums.StatusEnabled}
	repo.roles[2] = &entity.Role{ID: 2, Code: "viewer", Name: "只读", Status: enums.StatusEnabled}
	repo.userRoles[9] = []int64{2}
	repo.roleGroups[2] = []int64{20, 21}
	svc := newTestAdminService(repo)

	if err := svc.DeleteRole(ctx, &input.DeleteRoleInput{ID: 0}); !errors.Is(err, service_contract.ErrRoleNotFound) {
		t.Fatalf("非法 ID 期望 ErrRoleNotFound，实际 %v", err)
	}
	if err := svc.DeleteRole(ctx, &input.DeleteRoleInput{ID: 1}); !errors.Is(err, service_contract.ErrBuiltinRoleImmutable) {
		t.Fatalf("删除内置角色期望 ErrBuiltinRoleImmutable，实际 %v", err)
	}
	if err := svc.DeleteRole(ctx, &input.DeleteRoleInput{ID: 2}); !errors.Is(err, service_contract.ErrRoleBoundToUsers) {
		t.Fatalf("角色仍绑定用户期望 ErrRoleBoundToUsers，实际 %v", err)
	}
	if _, ok := repo.roles[2]; !ok {
		t.Fatal("绑定校验失败时不应删除角色")
	}

	delete(repo.userRoles, 9)
	if err := svc.DeleteRole(ctx, &input.DeleteRoleInput{ID: 2, OperatorID: 1}); err != nil {
		t.Fatalf("删除角色失败: %v", err)
	}
	if _, ok := repo.roles[2]; ok {
		t.Fatal("角色未硬删除")
	}
	if _, ok := repo.roleGroups[2]; ok {
		t.Fatal("角色权限组关系未清理")
	}
	if err := svc.DeleteRole(ctx, &input.DeleteRoleInput{ID: 2}); !errors.Is(err, service_contract.ErrRoleNotFound) {
		t.Fatalf("重复删除期望 ErrRoleNotFound，实际 %v", err)
	}
}

func TestBindRoleGroupsRules(t *testing.T) {
	ctx := context.Background()
	repo := newFakeRepo()
	repo.roles[1] = &entity.Role{ID: 1, Code: enums.RoleCodeSuperAdmin, Builtin: true, Status: enums.StatusEnabled}
	repo.roles[2] = &entity.Role{ID: 2, Code: "viewer", Status: enums.StatusEnabled}
	repo.permissionGroups[20] = &entity.PermissionGroup{ID: 20, Code: "system_user", Status: enums.StatusEnabled}
	repo.permissionGroups[21] = &entity.PermissionGroup{ID: 21, Code: "system_role", Status: enums.StatusDisabled}
	svc := newTestAdminService(repo)

	if err := svc.BindRoleGroups(ctx, &input.BindRoleGroupsInput{RoleID: 1, GroupIDs: []int64{20}}); !errors.Is(err, service_contract.ErrBuiltinRoleImmutable) {
		t.Fatalf("内置角色绑定权限组期望 ErrBuiltinRoleImmutable，实际 %v", err)
	}
	if err := svc.BindRoleGroups(ctx, &input.BindRoleGroupsInput{RoleID: 2, GroupIDs: []int64{20, 99}}); !errors.Is(err, service_contract.ErrPermissionGroupNotFound) {
		t.Fatalf("权限组不存在期望 ErrPermissionGroupNotFound，实际 %v", err)
	}
	// 停用的权限组允许绑定：不授予任何权限，由界面标注
	if err := svc.BindRoleGroups(ctx, &input.BindRoleGroupsInput{RoleID: 2, GroupIDs: []int64{21}, OperatorID: 1}); err != nil {
		t.Fatalf("绑定停用权限组应放行: %v", err)
	}
	if ids := repo.roleGroups[2]; len(ids) != 1 || ids[0] != 21 {
		t.Fatalf("停用权限组未写入: %v", ids)
	}
	if err := svc.BindRoleGroups(ctx, &input.BindRoleGroupsInput{RoleID: 2, GroupIDs: []int64{20, 0}}); !errors.Is(err, service_contract.ErrInvalidInput) {
		t.Fatalf("非法权限组 ID 期望 ErrInvalidInput，实际 %v", err)
	}
	if err := svc.BindRoleGroups(ctx, &input.BindRoleGroupsInput{RoleID: 2, GroupIDs: []int64{20, 20}, OperatorID: 1}); err != nil {
		t.Fatalf("绑定权限组失败: %v", err)
	}
	if ids := repo.roleGroups[2]; len(ids) != 1 || ids[0] != 20 {
		t.Fatalf("权限组绑定未去重或未写入: %v", ids)
	}
}

func TestGetRoleAggregatesPermissionGroups(t *testing.T) {
	ctx := context.Background()
	repo := newFakeRepo()
	repo.roles[1] = &entity.Role{ID: 1, Code: "viewer", Status: enums.StatusEnabled}
	repo.permissionGroups[20] = &entity.PermissionGroup{ID: 20, Code: "system_user", Status: enums.StatusEnabled}
	repo.permissionGroups[21] = &entity.PermissionGroup{ID: 21, Code: "system_role", Status: enums.StatusEnabled}
	repo.permissionGroups[22] = &entity.PermissionGroup{ID: 22, Code: "system_frozen", Status: enums.StatusDisabled}
	repo.roleGroups[1] = []int64{20, 21, 22}
	repo.groupPermissions[20] = []int64{10, 11}
	repo.groupPermissions[21] = []int64{11, 12}
	repo.groupPermissions[22] = []int64{13}
	repo.permissions[10] = &entity.Permission{ID: 10, Status: enums.StatusEnabled}
	repo.permissions[11] = &entity.Permission{ID: 11, Status: enums.StatusEnabled}
	repo.permissions[12] = &entity.Permission{ID: 12, Status: enums.StatusDisabled}
	repo.permissions[13] = &entity.Permission{ID: 13, Status: enums.StatusEnabled}
	svc := newTestAdminService(repo)

	detail, err := svc.GetRole(ctx, 1)
	if err != nil {
		t.Fatalf("查询角色详情失败: %v", err)
	}
	if len(detail.GroupIDs) != 3 || detail.GroupIDs[0] != 20 || detail.GroupIDs[1] != 21 || detail.GroupIDs[2] != 22 {
		t.Fatalf("角色权限组不符: %v", detail.GroupIDs)
	}
	// 聚合权限含停用权限组与停用权限，供界面标注
	if len(detail.PermissionIDs) != 4 {
		t.Fatalf("聚合权限应去重为 4 项: %v", detail.PermissionIDs)
	}
	if detail.PermissionIDs[0] != 10 || detail.PermissionIDs[1] != 11 || detail.PermissionIDs[2] != 12 || detail.PermissionIDs[3] != 13 {
		t.Fatalf("聚合权限排序不符: %v", detail.PermissionIDs)
	}
	// 生效权限只含启用权限组下的启用权限：12 本身停用、13 只挂在停用组下
	if len(detail.EffectivePermissionIDs) != 2 || detail.EffectivePermissionIDs[0] != 10 || detail.EffectivePermissionIDs[1] != 11 {
		t.Fatalf("生效权限不符: %v", detail.EffectivePermissionIDs)
	}
}

func TestCreatePermissionGroupRules(t *testing.T) {
	ctx := context.Background()
	repo := newFakeRepo()
	repo.permissionGroups[1] = &entity.PermissionGroup{ID: 1, Code: "system_user", Name: "用户管理", Status: enums.StatusEnabled}
	svc := newTestAdminService(repo)

	if _, err := svc.CreatePermissionGroup(ctx, &input.CreatePermissionGroupInput{Code: "system_user", Name: "重复"}); !errors.Is(err, service_contract.ErrPermissionGroupCodeExists) {
		t.Fatalf("重复编码期望 ErrPermissionGroupCodeExists，实际 %v", err)
	}

	invalid := []struct {
		name string
		in   *input.CreatePermissionGroupInput
	}{
		{name: "编码含大写", in: &input.CreatePermissionGroupInput{Code: "System", Name: "组"}},
		{name: "编码过短", in: &input.CreatePermissionGroupInput{Code: "ab", Name: "组"}},
		{name: "名称为空", in: &input.CreatePermissionGroupInput{Code: "viewer_group", Name: "  "}},
		{name: "备注超长", in: &input.CreatePermissionGroupInput{Code: "viewer_group", Name: "组", Remark: strings.Repeat("字", 256)}},
	}
	for _, tc := range invalid {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := svc.CreatePermissionGroup(ctx, tc.in); !errors.Is(err, service_contract.ErrInvalidInput) {
				t.Fatalf("期望 ErrInvalidInput，实际 %v", err)
			}
		})
	}

	group, err := svc.CreatePermissionGroup(ctx, &input.CreatePermissionGroupInput{Code: " viewer_group ", Name: " 只读组 ", Remark: " 备注 ", OperatorID: 1})
	if err != nil {
		t.Fatalf("创建权限组失败: %v", err)
	}
	if group.ID == 0 || group.Status != enums.StatusEnabled {
		t.Fatalf("创建结果不符合预期: %+v", group)
	}
	if group.Code != "viewer_group" || group.Name != "只读组" || group.Remark != "备注" {
		t.Fatalf("创建结果未去除首尾空白: %+v", group)
	}
	if _, ok := repo.permissionGroups[group.ID]; !ok {
		t.Fatal("权限组未写入仓储")
	}
}

func TestUpdatePermissionGroupRules(t *testing.T) {
	ctx := context.Background()
	repo := newFakeRepo()
	repo.permissionGroups[1] = &entity.PermissionGroup{ID: 1, Code: "system_user", Name: "用户管理", Status: enums.StatusEnabled}
	svc := newTestAdminService(repo)

	if err := svc.UpdatePermissionGroup(ctx, &input.UpdatePermissionGroupInput{ID: 0, Name: "组"}); !errors.Is(err, service_contract.ErrPermissionGroupNotFound) {
		t.Fatalf("非法 ID 期望 ErrPermissionGroupNotFound，实际 %v", err)
	}
	if err := svc.UpdatePermissionGroup(ctx, &input.UpdatePermissionGroupInput{ID: 1, Name: " "}); !errors.Is(err, service_contract.ErrInvalidInput) {
		t.Fatalf("名称为空期望 ErrInvalidInput，实际 %v", err)
	}
	if err := svc.UpdatePermissionGroup(ctx, &input.UpdatePermissionGroupInput{ID: 1, Name: "用户组", Remark: "备注", OperatorID: 1}); err != nil {
		t.Fatalf("更新权限组失败: %v", err)
	}
	if repo.permissionGroups[1].Name != "用户组" || repo.permissionGroups[1].Remark != "备注" {
		t.Fatalf("更新未生效: %+v", repo.permissionGroups[1])
	}
	if repo.permissionGroups[1].Status != enums.StatusEnabled {
		t.Fatalf("更新资料不应改变启停状态: %+v", repo.permissionGroups[1])
	}
}

func TestSetPermissionGroupStatusRules(t *testing.T) {
	ctx := context.Background()
	repo := newFakeRepo()
	repo.permissionGroups[1] = &entity.PermissionGroup{ID: 1, Code: "system_user", Status: enums.StatusEnabled}
	svc := newTestAdminService(repo)

	if err := svc.SetPermissionGroupStatus(ctx, &input.SetPermissionGroupStatusInput{ID: 0, Status: enums.StatusEnabled}); !errors.Is(err, service_contract.ErrPermissionGroupNotFound) {
		t.Fatalf("非法 ID 期望 ErrPermissionGroupNotFound，实际 %v", err)
	}
	if err := svc.SetPermissionGroupStatus(ctx, &input.SetPermissionGroupStatusInput{ID: 1, Status: enums.Status("UNKNOWN")}); !errors.Is(err, service_contract.ErrInvalidInput) {
		t.Fatalf("非法状态期望 ErrInvalidInput，实际 %v", err)
	}
	if err := svc.SetPermissionGroupStatus(ctx, &input.SetPermissionGroupStatusInput{ID: 1, Status: enums.StatusDisabled, OperatorID: 1}); err != nil {
		t.Fatalf("设置权限组启停失败: %v", err)
	}
	if repo.permissionGroups[1].Status != enums.StatusDisabled {
		t.Fatalf("权限组启停未生效: %+v", repo.permissionGroups[1])
	}
}

func TestDeletePermissionGroupRules(t *testing.T) {
	ctx := context.Background()
	repo := newFakeRepo()
	repo.permissionGroups[20] = &entity.PermissionGroup{ID: 20, Code: "system_user", Status: enums.StatusEnabled}
	repo.roleGroups[7] = []int64{20}
	repo.groupPermissions[20] = []int64{10, 11}
	svc := newTestAdminService(repo)

	if err := svc.DeletePermissionGroup(ctx, &input.DeletePermissionGroupInput{ID: 0}); !errors.Is(err, service_contract.ErrPermissionGroupNotFound) {
		t.Fatalf("非法 ID 期望 ErrPermissionGroupNotFound，实际 %v", err)
	}
	if err := svc.DeletePermissionGroup(ctx, &input.DeletePermissionGroupInput{ID: 20}); !errors.Is(err, service_contract.ErrPermissionGroupBoundToRoles) {
		t.Fatalf("权限组仍绑定角色期望 ErrPermissionGroupBoundToRoles，实际 %v", err)
	}
	if _, ok := repo.permissionGroups[20]; !ok {
		t.Fatal("绑定校验失败时不应删除权限组")
	}

	delete(repo.roleGroups, 7)
	if err := svc.DeletePermissionGroup(ctx, &input.DeletePermissionGroupInput{ID: 20, OperatorID: 1}); err != nil {
		t.Fatalf("删除权限组失败: %v", err)
	}
	if _, ok := repo.permissionGroups[20]; ok {
		t.Fatal("权限组未硬删除")
	}
	if _, ok := repo.groupPermissions[20]; ok {
		t.Fatal("权限组权限关系未清理")
	}
	if err := svc.DeletePermissionGroup(ctx, &input.DeletePermissionGroupInput{ID: 20}); !errors.Is(err, service_contract.ErrPermissionGroupNotFound) {
		t.Fatalf("重复删除期望 ErrPermissionGroupNotFound，实际 %v", err)
	}
}

func TestPermissionCRUD(t *testing.T) {
	ctx := context.Background()
	repo := newFakeRepo()
	repo.permissions[10] = &entity.Permission{
		ID: 10, Code: "system:user", Name: "系统用户", Type: enums.PermissionTypeMenu,
		UIPath: "/system/user", Status: enums.StatusEnabled,
	}
	repo.permissions[11] = &entity.Permission{
		ID: 11, Code: "system:user:list", Name: "用户列表", Type: enums.PermissionTypeAPI,
		HTTPMethod: "GET", HTTPPath: "/api/system/user", Status: enums.StatusEnabled,
	}
	svc := newTestAdminService(repo)

	if _, err := svc.CreatePermission(ctx, &input.CreatePermissionInput{
		Code: "system:user", Name: "重复编码", Type: enums.PermissionTypeMenu, UIPath: "/system/dup",
	}); !errors.Is(err, service_contract.ErrPermissionCodeExists) {
		t.Fatalf("重复编码期望 ErrPermissionCodeExists，实际 %v", err)
	}
	if _, err := svc.CreatePermission(ctx, &input.CreatePermissionInput{
		Code: "system:dup", Name: "重复菜单", Type: enums.PermissionTypeMenu, UIPath: "/system/user",
	}); !errors.Is(err, service_contract.ErrPermissionResourceExists) {
		t.Fatalf("重复菜单路径期望 ErrPermissionResourceExists，实际 %v", err)
	}
	if _, err := svc.CreatePermission(ctx, &input.CreatePermissionInput{
		Code: "system:dup", Name: "重复接口", Type: enums.PermissionTypeAPI, HTTPMethod: "GET", HTTPPath: "/api/system/user",
	}); !errors.Is(err, service_contract.ErrPermissionResourceExists) {
		t.Fatalf("重复接口路由期望 ErrPermissionResourceExists，实际 %v", err)
	}

	invalid := []struct {
		name string
		in   *input.CreatePermissionInput
	}{
		{name: "编码非法", in: &input.CreatePermissionInput{Code: "System:Dup", Name: "权限", Type: enums.PermissionTypeMenu, UIPath: "/system/dup"}},
		{name: "类型非法", in: &input.CreatePermissionInput{Code: "system:dup", Name: "权限", Type: enums.PermissionType("OTHER")}},
		{name: "菜单路径非法", in: &input.CreatePermissionInput{Code: "system:dup", Name: "权限", Type: enums.PermissionTypeMenu, UIPath: "system/dup"}},
		{name: "接口方法非法", in: &input.CreatePermissionInput{Code: "system:dup", Name: "权限", Type: enums.PermissionTypeAPI, HTTPMethod: "FETCH", HTTPPath: "/api/system/dup"}},
		{name: "接口路径非法", in: &input.CreatePermissionInput{Code: "system:dup", Name: "权限", Type: enums.PermissionTypeAPI, HTTPMethod: "GET", HTTPPath: "system/dup"}},
		{name: "名称为空", in: &input.CreatePermissionInput{Code: "system:dup", Name: " ", Type: enums.PermissionTypeMenu, UIPath: "/system/dup"}},
	}
	for _, tc := range invalid {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := svc.CreatePermission(ctx, tc.in); !errors.Is(err, service_contract.ErrInvalidInput) {
				t.Fatalf("期望 ErrInvalidInput，实际 %v", err)
			}
		})
	}

	menu, err := svc.CreatePermission(ctx, &input.CreatePermissionInput{
		Code: " system:audit ", Name: " 审计 ", Type: enums.PermissionTypeMenu, UIPath: " /system/audit ", Remark: " 备注 ", OperatorID: 1,
	})
	if err != nil {
		t.Fatalf("创建菜单权限失败: %v", err)
	}
	if menu.Code != "system:audit" || menu.Name != "审计" || menu.UIPath != "/system/audit" ||
		menu.HTTPMethod != "" || menu.HTTPPath != "" || menu.Status != enums.StatusEnabled || menu.Remark != "备注" {
		t.Fatalf("菜单权限归一化不符: %+v", menu)
	}
	api, err := svc.CreatePermission(ctx, &input.CreatePermissionInput{
		Code: "system:audit:list", Name: "审计列表", Type: enums.PermissionTypeAPI, HTTPMethod: "get", HTTPPath: " /api/system/audit ", OperatorID: 1,
	})
	if err != nil {
		t.Fatalf("创建接口权限失败: %v", err)
	}
	if api.HTTPMethod != "GET" || api.HTTPPath != "/api/system/audit" || api.UIPath != "" {
		t.Fatalf("接口权限归一化不符: %+v", api)
	}
	// 接口路径不再强制 /api 前缀，便于对接网关改写或多前缀部署
	custom, err := svc.CreatePermission(ctx, &input.CreatePermissionInput{
		Code: "system:probe", Name: "健康探针", Type: enums.PermissionTypeAPI, HTTPMethod: "get", HTTPPath: " /v1/health ", OperatorID: 1,
	})
	if err != nil {
		t.Fatalf("非 /api 前缀的接口路径应允许: %v", err)
	}
	if custom.HTTPPath != "/v1/health" || custom.HTTPMethod != "GET" {
		t.Fatalf("自定义前缀接口权限归一化不符: %+v", custom)
	}

	if err := svc.UpdatePermission(ctx, &input.UpdatePermissionInput{
		ID: menu.ID, Name: "审计菜单", UIPath: "/system/user",
	}); !errors.Is(err, service_contract.ErrPermissionResourceExists) {
		t.Fatalf("更新为重复菜单路径期望 ErrPermissionResourceExists，实际 %v", err)
	}
	if err := svc.UpdatePermission(ctx, &input.UpdatePermissionInput{
		ID: menu.ID, Name: " 审计菜单 ", UIPath: " /system/audit-log ", OperatorID: 1,
	}); err != nil {
		t.Fatalf("更新菜单权限失败: %v", err)
	}
	if repo.permissions[menu.ID].Name != "审计菜单" || repo.permissions[menu.ID].UIPath != "/system/audit-log" {
		t.Fatalf("菜单权限更新未生效: %+v", repo.permissions[menu.ID])
	}
	if repo.permissions[menu.ID].Type != enums.PermissionTypeMenu || repo.permissions[menu.ID].Status != enums.StatusEnabled {
		t.Fatalf("更新不应改变类型或状态: %+v", repo.permissions[menu.ID])
	}

	repo.groupPermissions[20] = []int64{menu.ID, 10}
	if err := svc.DeletePermission(ctx, &input.DeletePermissionInput{ID: menu.ID, OperatorID: 1}); err != nil {
		t.Fatalf("删除权限失败: %v", err)
	}
	if _, ok := repo.permissions[menu.ID]; ok {
		t.Fatal("权限未硬删除")
	}
	if ids := repo.groupPermissions[20]; len(ids) != 1 || ids[0] != 10 {
		t.Fatalf("权限组引用未清理: %v", ids)
	}
	if err := svc.DeletePermission(ctx, &input.DeletePermissionInput{ID: menu.ID}); !errors.Is(err, service_contract.ErrPermissionNotFound) {
		t.Fatalf("重复删除期望 ErrPermissionNotFound，实际 %v", err)
	}
}

func TestBindGroupPermissionsRules(t *testing.T) {
	ctx := context.Background()
	repo := newFakeRepo()
	repo.permissionGroups[20] = &entity.PermissionGroup{ID: 20, Code: "system_user", Status: enums.StatusEnabled}
	repo.permissions[10] = &entity.Permission{ID: 10, Code: "system:user:list", Type: enums.PermissionTypeAPI, Status: enums.StatusEnabled}
	repo.permissions[11] = &entity.Permission{ID: 11, Code: "system:user:delete", Type: enums.PermissionTypeAPI, Status: enums.StatusDisabled}
	svc := newTestAdminService(repo)

	if err := svc.BindGroupPermissions(ctx, &input.BindGroupPermissionsInput{GroupID: 99, PermissionIDs: []int64{10}}); !errors.Is(err, service_contract.ErrPermissionGroupNotFound) {
		t.Fatalf("权限组不存在期望 ErrPermissionGroupNotFound，实际 %v", err)
	}
	if err := svc.BindGroupPermissions(ctx, &input.BindGroupPermissionsInput{GroupID: 20, PermissionIDs: []int64{10, 99}}); !errors.Is(err, service_contract.ErrPermissionNotFound) {
		t.Fatalf("权限不存在期望 ErrPermissionNotFound，实际 %v", err)
	}
	// 停用的权限允许绑定：实际不生效，由界面标注
	if err := svc.BindGroupPermissions(ctx, &input.BindGroupPermissionsInput{GroupID: 20, PermissionIDs: []int64{11}, OperatorID: 1}); err != nil {
		t.Fatalf("绑定停用权限应放行: %v", err)
	}
	if ids := repo.groupPermissions[20]; len(ids) != 1 || ids[0] != 11 {
		t.Fatalf("停用权限未写入: %v", ids)
	}
	if err := svc.BindGroupPermissions(ctx, &input.BindGroupPermissionsInput{GroupID: 20, PermissionIDs: []int64{10, 0}}); !errors.Is(err, service_contract.ErrInvalidInput) {
		t.Fatalf("非法权限 ID 期望 ErrInvalidInput，实际 %v", err)
	}
	if err := svc.BindGroupPermissions(ctx, &input.BindGroupPermissionsInput{GroupID: 20, PermissionIDs: []int64{10, 10}, OperatorID: 1}); err != nil {
		t.Fatalf("绑定权限失败: %v", err)
	}
	if ids := repo.groupPermissions[20]; len(ids) != 1 || ids[0] != 10 {
		t.Fatalf("权限绑定未去重或未写入: %v", ids)
	}

	if detail, err := svc.GetPermissionGroup(ctx, 20); err != nil || len(detail.PermissionIDs) != 1 || detail.PermissionIDs[0] != 10 {
		t.Fatalf("权限组详情不符: %+v err=%v", detail, err)
	}
}

func TestBindUserRolesRules(t *testing.T) {
	ctx := context.Background()
	repo := newFakeRepo()
	repo.roles[1] = &entity.Role{ID: 1, Code: enums.RoleCodeSuperAdmin, Builtin: true, Status: enums.StatusEnabled}
	repo.roles[2] = &entity.Role{ID: 2, Code: "viewer", Status: enums.StatusEnabled}
	repo.roles[3] = &entity.Role{ID: 3, Code: "auditor", Status: enums.StatusDisabled}
	svc := newTestAdminService(repo)

	// 角色解析与绑定共用同一套规则：存在性、启用态、ID 合法性
	if roles, err := svc.ResolveBindingRoles(ctx, []int64{2}); err != nil || len(roles) != 1 || roles[0].ID != 2 {
		t.Fatalf("解析角色不符: %+v err=%v", roles, err)
	}
	if _, err := svc.ResolveBindingRoles(ctx, []int64{99}); !errors.Is(err, service_contract.ErrRoleNotFound) {
		t.Fatalf("解析不存在角色期望 ErrRoleNotFound，实际 %v", err)
	}
	if _, err := svc.ResolveBindingRoles(ctx, []int64{3}); !errors.Is(err, service_contract.ErrRelatedDisabled) {
		t.Fatalf("解析停用角色期望 ErrRelatedDisabled，实际 %v", err)
	}

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
	scopeGroup := &entity.PermissionGroup{ID: 20, Status: enums.StatusEnabled}
	repo.permissionGroups[20] = scopeGroup
	svc := newTestAdminService(repo)

	if _, err := svc.PageRoles(ctx, &input.PageRolesInput{Status: enums.Status("UNKNOWN")}); !errors.Is(err, service_contract.ErrInvalidInput) {
		t.Fatalf("非法角色状态期望 ErrInvalidInput，实际 %v", err)
	}
	if _, err := svc.PagePermissionGroups(ctx, &input.PagePermissionGroupsInput{Status: enums.Status("UNKNOWN")}); !errors.Is(err, service_contract.ErrInvalidInput) {
		t.Fatalf("非法权限组状态期望 ErrInvalidInput，实际 %v", err)
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
	if _, err := svc.ResolveBindingRoles(ctx, []int64{0}); !errors.Is(err, service_contract.ErrInvalidInput) {
		t.Fatalf("非法角色 ID 期望 ErrInvalidInput，实际 %v", err)
	}
}

// TestCreatePermissionConflictResolution 校验并发插入撞唯一键时按真实冲突字段返回错误，
// 而不是把所有重复键都报成“资源已存在”
func TestCreatePermissionConflictResolution(t *testing.T) {
	ctx := context.Background()

	// 编码冲突：预检查通过后，另一个请求抢先写入了相同编码
	repo := newFakeRepo()
	svc := newTestAdminService(repo)
	repo.createPermission = func(permission *entity.Permission) error {
		repo.permissions[900] = &entity.Permission{
			ID: 900, Code: permission.Code, Name: "并发写入", Type: enums.PermissionTypeMenu,
			UIPath: "/system/other", Status: enums.StatusEnabled,
		}
		return errors.Sentinel(repo_contract.ErrDuplicatedKey)
	}
	if _, err := svc.CreatePermission(ctx, &input.CreatePermissionInput{
		Code: "system:race", Name: "并发编码", Type: enums.PermissionTypeMenu, UIPath: "/system/race",
	}); !errors.Is(err, service_contract.ErrPermissionCodeExists) {
		t.Fatalf("编码冲突期望 ErrPermissionCodeExists，实际 %v", err)
	}

	// 菜单路径冲突：编码没撞，撞的是同一 ui_path
	repo = newFakeRepo()
	svc = newTestAdminService(repo)
	repo.createPermission = func(permission *entity.Permission) error {
		repo.permissions[901] = &entity.Permission{
			ID: 901, Code: "system:other", Name: "并发写入", Type: enums.PermissionTypeMenu,
			UIPath: permission.UIPath, Status: enums.StatusEnabled,
		}
		return errors.Sentinel(repo_contract.ErrDuplicatedKey)
	}
	if _, err := svc.CreatePermission(ctx, &input.CreatePermissionInput{
		Code: "system:race", Name: "并发路径", Type: enums.PermissionTypeMenu, UIPath: "/system/race",
	}); !errors.Is(err, service_contract.ErrPermissionResourceExists) {
		t.Fatalf("菜单路径冲突期望 ErrPermissionResourceExists，实际 %v", err)
	}

	// 接口路由冲突：编码没撞，撞的是同一 method + path
	repo = newFakeRepo()
	svc = newTestAdminService(repo)
	repo.createPermission = func(permission *entity.Permission) error {
		repo.permissions[902] = &entity.Permission{
			ID: 902, Code: "system:other:list", Name: "并发写入", Type: enums.PermissionTypeAPI,
			HTTPMethod: permission.HTTPMethod, HTTPPath: permission.HTTPPath, Status: enums.StatusEnabled,
		}
		return errors.Sentinel(repo_contract.ErrDuplicatedKey)
	}
	if _, err := svc.CreatePermission(ctx, &input.CreatePermissionInput{
		Code: "system:race:list", Name: "并发路由", Type: enums.PermissionTypeAPI,
		HTTPMethod: "GET", HTTPPath: "/api/system/race",
	}); !errors.Is(err, service_contract.ErrPermissionResourceExists) {
		t.Fatalf("接口路由冲突期望 ErrPermissionResourceExists，实际 %v", err)
	}

	// 冲突行在回查前消失：保守返回资源冲突，不泄漏数据库错误
	repo = newFakeRepo()
	svc = newTestAdminService(repo)
	repo.createPermission = func(*entity.Permission) error {
		return errors.Sentinel(repo_contract.ErrDuplicatedKey)
	}
	if _, err := svc.CreatePermission(ctx, &input.CreatePermissionInput{
		Code: "system:ghost", Name: "幽灵冲突", Type: enums.PermissionTypeMenu, UIPath: "/system/ghost",
	}); !errors.Is(err, service_contract.ErrPermissionResourceExists) {
		t.Fatalf("无法定位冲突行时期望 ErrPermissionResourceExists，实际 %v", err)
	}
}

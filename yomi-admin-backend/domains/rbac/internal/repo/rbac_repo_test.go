package repo

import (
	"context"
	"testing"
	"yomi-admin-backend/domains/rbac/internal/repo_contract"
	"yomi-admin-backend/domains/rbac/pkg/shared/entity"
	"yomi-admin-backend/domains/rbac/pkg/shared/enums"

	"gitee.com/hq_project/yomi-boot/component/common/db/dbtest"
	"gitee.com/hq_project/yomi-boot/component/core/errors"
	sharedenums "gitee.com/hq_project/yomi-boot/component/core/shared/enums"
	"gitee.com/hq_project/yomi-boot/component/core/utils/ptrutil"
)

func TestRbacRepoCRUDAndAuthQueries(t *testing.T) {
	gormDB, cleanup := dbtest.New(t,
		dbtest.WithTables(&roleModel{}, &permissionModel{}, &userRoleRelModel{}, &rolePermissionRelModel{}),
	)
	defer cleanup()

	r := NewRbacRepo(gormDB, sharedenums.MySQL)
	ctx := context.Background()

	// 角色 CRUD
	superRole := &entity.Role{ID: 1, Code: enums.RoleCodeSuperAdmin, Name: "超级管理员", Builtin: true, Status: enums.StatusEnabled}
	if err := r.CreateRole(ctx, superRole, 1); err != nil {
		t.Fatalf("创建超管角色失败: %v", err)
	}
	viewerRole := &entity.Role{ID: 2, Code: "viewer", Name: "只读", Status: enums.StatusEnabled}
	if err := r.CreateRole(ctx, viewerRole, 1); err != nil {
		t.Fatalf("创建只读角色失败: %v", err)
	}
	if _, ok, err := r.FindRoleByCode(ctx, "viewer"); err != nil || !ok {
		t.Fatalf("按编码查询角色失败: ok=%v err=%v", ok, err)
	}
	if _, ok, _ := r.FindRoleByID(ctx, 2); !ok {
		t.Fatal("按 ID 查询角色失败")
	}
	dup := &entity.Role{ID: 3, Code: "viewer", Name: "重复", Status: enums.StatusEnabled}
	if err := r.CreateRole(ctx, dup, 1); !errors.Is(err, repo_contract.ErrDuplicatedKey) {
		t.Fatalf("重复编码期望 ErrDuplicatedKey，实际 %v", err)
	}

	// 权限目录直接落库（业务无创建权限接口）
	menu := &permissionModel{
		Code:           "system:user",
		Name:           "系统用户",
		PermissionType: enums.PermissionTypeMenu,
		UIPath:         ptrutil.PtrFromString("/system/user"),
		Status:         enums.StatusEnabled,
	}
	menu.ID = 10
	if err := gormDB.Create(menu).Error; err != nil {
		t.Fatalf("写入菜单权限失败: %v", err)
	}
	api := &permissionModel{
		Code:           "system:user:list",
		Name:           "用户列表",
		PermissionType: enums.PermissionTypeAPI,
		HTTPMethod:     ptrutil.PtrFromString("GET"),
		HTTPPath:       ptrutil.PtrFromString("/api/system/users"),
		Status:         enums.StatusEnabled,
	}
	api.ID = 11
	if err := gormDB.Create(api).Error; err != nil {
		t.Fatalf("写入接口权限失败: %v", err)
	}
	disabled := &permissionModel{
		Code:           "system:user:delete",
		Name:           "删除用户",
		PermissionType: enums.PermissionTypeAPI,
		HTTPMethod:     ptrutil.PtrFromString("DELETE"),
		HTTPPath:       ptrutil.PtrFromString("/api/system/users/:id"),
		Status:         enums.StatusDisabled,
	}
	disabled.ID = 12
	if err := gormDB.Create(disabled).Error; err != nil {
		t.Fatalf("写入禁用权限失败: %v", err)
	}

	// 角色权限绑定
	if err := r.ReplaceRolePermissions(ctx, 2, []*entity.RolePermissionRel{
		{ID: 100, RoleID: 2, PermissionID: 10},
		{ID: 101, RoleID: 2, PermissionID: 11},
	}, 1); err != nil {
		t.Fatalf("绑定角色权限失败: %v", err)
	}
	permissionIDs, err := r.ListPermissionIDsByRoleID(ctx, 2)
	if err != nil || len(permissionIDs) != 2 {
		t.Fatalf("查询角色权限失败: ids=%v err=%v", permissionIDs, err)
	}

	// 用户角色绑定
	if err := r.ReplaceUserRoles(ctx, 7, []*entity.UserRoleRel{{ID: 200, UserID: 7, RoleID: 2}}, 1); err != nil {
		t.Fatalf("绑定用户角色失败: %v", err)
	}
	roleIDs, err := r.ListRoleIDsByUserID(ctx, 7)
	if err != nil || len(roleIDs) != 1 || roleIDs[0] != 2 {
		t.Fatalf("查询用户角色失败: ids=%v err=%v", roleIDs, err)
	}
	rolesByUser, err := r.ListRolesByUserIDs(ctx, []int64{7, 8})
	if err != nil {
		t.Fatalf("批量查询用户角色失败: %v", err)
	}
	if len(rolesByUser[7]) != 1 || rolesByUser[7][0].Code != "viewer" {
		t.Fatalf("批量用户角色结果不符: %+v", rolesByUser)
	}

	// 超管判定
	if ok, err := r.UserHasEnabledSuperAdmin(ctx, 7); err != nil || ok {
		t.Fatalf("普通用户不应是超管: ok=%v err=%v", ok, err)
	}
	if err := r.ReplaceUserRoles(ctx, 8, []*entity.UserRoleRel{{ID: 201, UserID: 8, RoleID: 1}}, 1); err != nil {
		t.Fatalf("绑定超管角色失败: %v", err)
	}
	if ok, err := r.UserHasEnabledSuperAdmin(ctx, 8); err != nil || !ok {
		t.Fatalf("用户 8 应是超管: ok=%v err=%v", ok, err)
	}
	superIDs, err := r.ListUserIDsWithEnabledSuperAdmin(ctx)
	if err != nil || len(superIDs) != 1 || superIDs[0] != 8 {
		t.Fatalf("超管用户列表不符: ids=%v err=%v", superIDs, err)
	}

	// 鉴权链：接口精确匹配、菜单并集、编码并集
	if ok, err := r.ExistsRouteAccess(ctx, 7, "GET", "/api/system/users"); err != nil || !ok {
		t.Fatalf("已授权接口应放行: ok=%v err=%v", ok, err)
	}
	if ok, err := r.ExistsRouteAccess(ctx, 7, "POST", "/api/system/users"); err != nil || ok {
		t.Fatalf("未授权方法不应放行: ok=%v err=%v", ok, err)
	}
	menus, err := r.ListMenusByUserID(ctx, 7)
	if err != nil || len(menus) != 1 || menus[0].UIPath != "/system/user" {
		t.Fatalf("用户菜单不符: menus=%+v err=%v", menus, err)
	}
	codes, err := r.ListPermissionCodesByUserID(ctx, 7)
	if err != nil || len(codes) != 2 {
		t.Fatalf("用户权限编码不符: codes=%v err=%v", codes, err)
	}
	enabledMenus, err := r.ListEnabledMenus(ctx)
	if err != nil || len(enabledMenus) != 1 {
		t.Fatalf("启用菜单列表不符: menus=%+v err=%v", enabledMenus, err)
	}
	enabledCodes, err := r.ListEnabledPermissionCodes(ctx)
	if err != nil || len(enabledCodes) != 2 {
		t.Fatalf("启用权限编码不符: codes=%v err=%v", enabledCodes, err)
	}

	// 分页
	items, total, err := r.PageRoles(ctx, "", enums.StatusEnabled, 1, 10)
	if err != nil || total != 2 || len(items) != 2 {
		t.Fatalf("角色分页不符: total=%d items=%d err=%v", total, len(items), err)
	}
	permissionItems, permissionTotal, err := r.PagePermissions(ctx, "", enums.PermissionTypeAPI, "", 1, 10)
	if err != nil || permissionTotal != 2 || len(permissionItems) != 2 {
		t.Fatalf("权限分页不符: total=%d items=%d err=%v", permissionTotal, len(permissionItems), err)
	}

	// 权限启停
	if err := r.UpdatePermissionStatus(ctx, 12, enums.StatusEnabled, 1); err != nil {
		t.Fatalf("更新权限状态失败: %v", err)
	}
	if permission, ok, _ := r.FindPermissionByID(ctx, 12); !ok || permission.Status != enums.StatusEnabled {
		t.Fatalf("权限状态未更新: %+v", permission)
	}

	// 角色资料更新
	viewerRole.Name = "只读角色"
	if err := r.UpdateRole(ctx, viewerRole, 1); err != nil {
		t.Fatalf("更新角色失败: %v", err)
	}
	if role, ok, _ := r.FindRoleByID(ctx, 2); !ok || role.Name != "只读角色" {
		t.Fatalf("角色更新未生效: %+v", role)
	}
}

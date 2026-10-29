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
		dbtest.WithTables(&roleModel{}, &permissionModel{}, &permissionGroupModel{}, &userRoleRelModel{}, &roleGroupRelModel{}, &groupPermissionRelModel{}),
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
	auditorRole := &entity.Role{ID: 3, Code: "auditor", Name: "审计", Status: enums.StatusEnabled}
	if err := r.CreateRole(ctx, auditorRole, 1); err != nil {
		t.Fatalf("创建审计角色失败: %v", err)
	}
	if _, ok, err := r.FindRoleByCode(ctx, "viewer"); err != nil || !ok {
		t.Fatalf("按编码查询角色失败: ok=%v err=%v", ok, err)
	}
	if _, ok, _ := r.FindRoleByID(ctx, 2); !ok {
		t.Fatal("按 ID 查询角色失败")
	}
	dup := &entity.Role{ID: 4, Code: "viewer", Name: "重复", Status: enums.StatusEnabled}
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
		HTTPPath:       ptrutil.PtrFromString("/api/system/user"),
		Status:         enums.StatusEnabled,
	}
	api.ID = 11
	if err := gormDB.Create(api).Error; err != nil {
		t.Fatalf("写入接口权限失败: %v", err)
	}
	roleMenu := &permissionModel{
		Code:           "system:role",
		Name:           "角色管理",
		PermissionType: enums.PermissionTypeMenu,
		UIPath:         ptrutil.PtrFromString("/system/role"),
		Status:         enums.StatusEnabled,
	}
	roleMenu.ID = 12
	if err := gormDB.Create(roleMenu).Error; err != nil {
		t.Fatalf("写入角色菜单失败: %v", err)
	}
	disabled := &permissionModel{
		Code:           "system:user:delete",
		Name:           "删除用户",
		PermissionType: enums.PermissionTypeAPI,
		HTTPMethod:     ptrutil.PtrFromString("POST"),
		HTTPPath:       ptrutil.PtrFromString("/api/system/user/delete"),
		Status:         enums.StatusDisabled,
	}
	disabled.ID = 13
	if err := gormDB.Create(disabled).Error; err != nil {
		t.Fatalf("写入禁用权限失败: %v", err)
	}

	// 权限组 CRUD
	userGroup := &entity.PermissionGroup{ID: 20, Code: "system_user", Name: "用户管理", Status: enums.StatusEnabled}
	readGroup := &entity.PermissionGroup{ID: 21, Code: "system_user_read", Name: "用户只读", Status: enums.StatusEnabled}
	disabledGroup := &entity.PermissionGroup{ID: 22, Code: "system_role_group", Name: "角色管理组", Status: enums.StatusDisabled}
	for _, group := range []*entity.PermissionGroup{userGroup, readGroup, disabledGroup} {
		if err := r.CreatePermissionGroup(ctx, group, 1); err != nil {
			t.Fatalf("创建权限组失败: %v", err)
		}
	}
	if _, ok, err := r.FindPermissionGroupByCode(ctx, "system_user"); err != nil || !ok {
		t.Fatalf("按编码查询权限组失败: ok=%v err=%v", ok, err)
	}
	if groups, err := r.FindPermissionGroupsByIDs(ctx, []int64{20, 21, 99}); err != nil || len(groups) != 2 {
		t.Fatalf("批量查询权限组不符: groups=%d err=%v", len(groups), err)
	}
	userGroup.Name = "用户权限"
	if err := r.UpdatePermissionGroup(ctx, userGroup, 1); err != nil {
		t.Fatalf("更新权限组失败: %v", err)
	}
	if group, ok, _ := r.FindPermissionGroupByID(ctx, 20); !ok || group.Name != "用户权限" {
		t.Fatalf("权限组更新未生效: %+v", group)
	}

	// 权限组绑定权限
	if err := r.ReplaceGroupPermissions(ctx, 20, []*entity.GroupPermissionRel{
		{ID: 100, GroupID: 20, PermissionID: 10},
		{ID: 101, GroupID: 20, PermissionID: 11},
	}, 1); err != nil {
		t.Fatalf("绑定权限组权限失败: %v", err)
	}
	if err := r.ReplaceGroupPermissions(ctx, 21, []*entity.GroupPermissionRel{
		{ID: 102, GroupID: 21, PermissionID: 11},
	}, 1); err != nil {
		t.Fatalf("绑定权限组权限失败: %v", err)
	}
	if err := r.ReplaceGroupPermissions(ctx, 22, []*entity.GroupPermissionRel{
		{ID: 103, GroupID: 22, PermissionID: 12},
	}, 1); err != nil {
		t.Fatalf("绑定权限组权限失败: %v", err)
	}
	if ids, err := r.ListPermissionIDsByGroupID(ctx, 20); err != nil || len(ids) != 2 {
		t.Fatalf("查询权限组权限失败: ids=%v err=%v", ids, err)
	}

	// 角色绑定权限组
	if err := r.ReplaceRoleGroups(ctx, 2, []*entity.RoleGroupRel{
		{ID: 200, RoleID: 2, GroupID: 20},
		{ID: 201, RoleID: 2, GroupID: 21},
		{ID: 202, RoleID: 2, GroupID: 22},
	}, 1); err != nil {
		t.Fatalf("绑定角色权限组失败: %v", err)
	}
	if ids, err := r.ListGroupIDsByRoleID(ctx, 2); err != nil || len(ids) != 3 {
		t.Fatalf("查询角色权限组失败: ids=%v err=%v", ids, err)
	}
	// 角色详情聚合：跨组去重、不看启停
	aggregated, err := r.ListPermissionIDsByRoleID(ctx, 2)
	if err != nil || len(aggregated) != 3 || aggregated[0] != 10 || aggregated[1] != 11 || aggregated[2] != 12 {
		t.Fatalf("角色聚合权限不符: ids=%v err=%v", aggregated, err)
	}
	// 生效权限：停用权限组（22）下的权限不计入
	effective, err := r.ListEffectivePermissionIDsByRoleID(ctx, 2)
	if err != nil || len(effective) != 2 || effective[0] != 10 || effective[1] != 11 {
		t.Fatalf("角色生效权限不符: ids=%v err=%v", effective, err)
	}
	// 停用权限即使挂在启用组下也不计入
	if err := r.ReplaceGroupPermissions(ctx, 21, []*entity.GroupPermissionRel{
		{ID: 104, GroupID: 21, PermissionID: 11},
		{ID: 105, GroupID: 21, PermissionID: 13},
	}, 1); err != nil {
		t.Fatalf("绑定权限组权限失败: %v", err)
	}
	if effective, err := r.ListEffectivePermissionIDsByRoleID(ctx, 2); err != nil ||
		len(effective) != 2 || effective[0] != 10 || effective[1] != 11 {
		t.Fatalf("停用权限不应计入生效集合: ids=%v err=%v", effective, err)
	}

	// 用户角色绑定
	if err := r.ReplaceUserRoles(ctx, 7, []*entity.UserRoleRel{{ID: 300, UserID: 7, RoleID: 2}}, 1); err != nil {
		t.Fatalf("绑定用户角色失败: %v", err)
	}
	if err := r.ReplaceUserRoles(ctx, 9, []*entity.UserRoleRel{{ID: 301, UserID: 9, RoleID: 3}}, 1); err != nil {
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
	if err := r.ReplaceUserRoles(ctx, 8, []*entity.UserRoleRel{{ID: 302, UserID: 8, RoleID: 1}}, 1); err != nil {
		t.Fatalf("绑定超管角色失败: %v", err)
	}
	if ok, err := r.UserHasEnabledSuperAdmin(ctx, 8); err != nil || !ok {
		t.Fatalf("用户 8 应是超管: ok=%v err=%v", ok, err)
	}
	superIDs, err := r.ListUserIDsWithEnabledSuperAdmin(ctx)
	if err != nil || len(superIDs) != 1 || superIDs[0] != 8 {
		t.Fatalf("超管用户列表不符: ids=%v err=%v", superIDs, err)
	}

	// 鉴权链：接口精确匹配、菜单并集、编码并集；禁用权限组不生效
	if ok, err := r.ExistsRouteAccess(ctx, 7, "GET", "/api/system/user"); err != nil || !ok {
		t.Fatalf("已授权接口应放行: ok=%v err=%v", ok, err)
	}
	if ok, err := r.ExistsRouteAccess(ctx, 7, "POST", "/api/system/user"); err != nil || ok {
		t.Fatalf("未授权方法不应放行: ok=%v err=%v", ok, err)
	}
	menus, err := r.ListMenusByUserID(ctx, 7)
	if err != nil || len(menus) != 1 || menus[0].UIPath != "/system/user" {
		t.Fatalf("用户菜单不符: menus=%+v err=%v", menus, err)
	}
	codes, err := r.ListPermissionCodesByUserID(ctx, 7)
	if err != nil || len(codes) != 2 {
		t.Fatalf("用户权限编码应去重为 2 项: codes=%v err=%v", codes, err)
	}
	if menus, err := r.ListMenusByUserID(ctx, 9); err != nil || len(menus) != 0 {
		t.Fatalf("仅绑禁用权限组的用户不应有菜单: menus=%+v err=%v", menus, err)
	}
	enabledMenus, err := r.ListEnabledMenus(ctx)
	if err != nil || len(enabledMenus) != 2 {
		t.Fatalf("启用菜单列表不符: menus=%+v err=%v", enabledMenus, err)
	}

	// 分页
	items, total, err := r.PageRoles(ctx, "", enums.StatusEnabled, 1, 10)
	if err != nil || total != 3 || len(items) != 3 {
		t.Fatalf("角色分页不符: total=%d items=%d err=%v", total, len(items), err)
	}
	groups, groupTotal, err := r.PagePermissionGroups(ctx, "", "", 1, 10)
	if err != nil || groupTotal != 3 || len(groups) != 3 {
		t.Fatalf("权限组分页不符: total=%d items=%d err=%v", groupTotal, len(groups), err)
	}
	permissionItems, permissionTotal, err := r.PagePermissions(ctx, "", enums.PermissionTypeAPI, "", 1, 10)
	if err != nil || permissionTotal != 2 || len(permissionItems) != 2 {
		t.Fatalf("权限分页不符: total=%d items=%d err=%v", permissionTotal, len(permissionItems), err)
	}

	// 权限启停
	if err := r.UpdatePermissionStatus(ctx, 13, enums.StatusEnabled, 1); err != nil {
		t.Fatalf("更新权限状态失败: %v", err)
	}
	if permission, ok, _ := r.FindPermissionByID(ctx, 13); !ok || permission.Status != enums.StatusEnabled {
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

	// 角色启停独立更新
	if err := r.UpdateRoleStatus(ctx, 2, enums.StatusDisabled, 1); err != nil {
		t.Fatalf("更新角色状态失败: %v", err)
	}
	if role, ok, _ := r.FindRoleByID(ctx, 2); !ok || role.Status != enums.StatusDisabled {
		t.Fatalf("角色状态未更新: %+v", role)
	}

	// 权限创建、查询与更新
	auditMenu := &entity.Permission{
		ID: 30, Code: "system:audit", Name: "审计", Type: enums.PermissionTypeMenu,
		UIPath: "/system/audit", Status: enums.StatusEnabled,
	}
	if err := r.CreatePermission(ctx, auditMenu, 1); err != nil {
		t.Fatalf("创建菜单权限失败: %v", err)
	}
	if permission, ok, err := r.FindPermissionByCode(ctx, "system:audit"); err != nil || !ok || permission.ID != 30 {
		t.Fatalf("按编码查询权限失败: %+v ok=%v err=%v", permission, ok, err)
	}
	if permission, ok, err := r.FindPermissionByUIPath(ctx, "/system/audit"); err != nil || !ok || permission.ID != 30 {
		t.Fatalf("按菜单路径查询权限失败: %+v ok=%v err=%v", permission, ok, err)
	}
	reportAPI := &entity.Permission{
		ID: 31, Code: "system:audit:report", Name: "审计报表", Type: enums.PermissionTypeAPI,
		HTTPMethod: "GET", HTTPPath: "/api/system/audit/report", Status: enums.StatusEnabled,
	}
	if err := r.CreatePermission(ctx, reportAPI, 1); err != nil {
		t.Fatalf("创建接口权限失败: %v", err)
	}
	if permission, ok, err := r.FindPermissionByHTTPRoute(ctx, "GET", "/api/system/audit/report"); err != nil || !ok || permission.ID != 31 {
		t.Fatalf("按 HTTP 路由查询权限失败: %+v ok=%v err=%v", permission, ok, err)
	}
	auditMenu.Name = "审计日志"
	auditMenu.UIPath = "/system/audit-log"
	if err := r.UpdatePermission(ctx, auditMenu, 1); err != nil {
		t.Fatalf("更新权限失败: %v", err)
	}
	if permission, ok, _ := r.FindPermissionByID(ctx, 30); !ok || permission.Name != "审计日志" || permission.UIPath != "/system/audit-log" {
		t.Fatalf("权限更新未生效: %+v", permission)
	}

	// 权限删除同步清理权限组关系
	if err := r.ReplaceGroupPermissions(ctx, 21, []*entity.GroupPermissionRel{
		{ID: 104, GroupID: 21, PermissionID: 30},
	}, 1); err != nil {
		t.Fatalf("权限组绑定待删除权限失败: %v", err)
	}
	if err := r.DeleteGroupPermissionsByPermissionID(ctx, 30); err != nil {
		t.Fatalf("清理权限组关系失败: %v", err)
	}
	if ids, err := r.ListPermissionIDsByGroupID(ctx, 21); err != nil || len(ids) != 0 {
		t.Fatalf("权限组关系未清理: ids=%v err=%v", ids, err)
	}
	if err := r.DeletePermission(ctx, 30); err != nil {
		t.Fatalf("删除权限失败: %v", err)
	}
	if _, ok, _ := r.FindPermissionByID(ctx, 30); ok {
		t.Fatal("权限未硬删除")
	}

	// 用户角色关系清理与绑定计数
	if err := r.DeleteUserRolesByUserID(ctx, 7); err != nil {
		t.Fatalf("清理用户角色失败: %v", err)
	}
	if ids, err := r.ListRoleIDsByUserID(ctx, 7); err != nil || len(ids) != 0 {
		t.Fatalf("用户角色未清理: ids=%v err=%v", ids, err)
	}
	if total, err := r.CountUserRolesByRoleID(ctx, 3); err != nil || total != 1 {
		t.Fatalf("角色用户绑定计数不符: total=%d err=%v", total, err)
	}
	if total, err := r.CountRoleGroupsByGroupID(ctx, 21); err != nil || total != 1 {
		t.Fatalf("权限组角色绑定计数不符: total=%d err=%v", total, err)
	}

	// 角色硬删除同步清理权限组关系
	if err := r.DeleteRoleGroupsByRoleID(ctx, 2); err != nil {
		t.Fatalf("清理角色权限组失败: %v", err)
	}
	if ids, err := r.ListGroupIDsByRoleID(ctx, 2); err != nil || len(ids) != 0 {
		t.Fatalf("角色权限组未清理: ids=%v err=%v", ids, err)
	}
	if err := r.DeleteRole(ctx, 2); err != nil {
		t.Fatalf("删除角色失败: %v", err)
	}
	if _, ok, _ := r.FindRoleByID(ctx, 2); ok {
		t.Fatal("角色未硬删除")
	}
	if total, err := r.CountRoleGroupsByGroupID(ctx, 21); err != nil || total != 0 {
		t.Fatalf("删除角色后权限组绑定计数应为 0: total=%d err=%v", total, err)
	}

	// 权限组硬删除同步清理权限关系
	if err := r.DeleteGroupPermissionsByGroupID(ctx, 21); err != nil {
		t.Fatalf("清理权限组权限失败: %v", err)
	}
	if err := r.DeletePermissionGroup(ctx, 21); err != nil {
		t.Fatalf("删除权限组失败: %v", err)
	}
	if _, ok, _ := r.FindPermissionGroupByID(ctx, 21); ok {
		t.Fatal("权限组未硬删除")
	}
}

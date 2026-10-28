package service

import (
	"context"
	"testing"
	"yomi-admin-backend/domains/rbac/pkg/shared/entity"
	"yomi-admin-backend/domains/rbac/pkg/shared/enums"
)

func TestAuthServiceSuperAdmin(t *testing.T) {
	ctx := context.Background()
	repo := newFakeRepo()
	repo.roles[1] = &entity.Role{ID: 1, Code: enums.RoleCodeSuperAdmin, Status: enums.StatusEnabled}
	repo.permissions[10] = &entity.Permission{ID: 10, Code: "system:user", Type: enums.PermissionTypeMenu, UIPath: "/system/user", Status: enums.StatusEnabled}
	repo.permissions[11] = &entity.Permission{ID: 11, Code: "system:user:list", Type: enums.PermissionTypeAPI, HTTPMethod: "GET", HTTPPath: "/api/system/users", Status: enums.StatusEnabled}
	repo.permissions[12] = &entity.Permission{ID: 12, Code: "system:role", Type: enums.PermissionTypeMenu, UIPath: "/system/role", Status: enums.StatusEnabled}
	repo.userRoles[1] = []int64{1}
	repo.rolePermissions[1] = []int64{10, 11}
	svc := NewRbacAuthService(repo)

	allowed, err := svc.CanAccessAPI(ctx, 1, "get", "/api/system/users")
	if err != nil || !allowed {
		t.Fatalf("超管应放行任意路由: allowed=%v err=%v", allowed, err)
	}
	codes, err := svc.ListPermissionCodes(ctx, 1)
	if err != nil {
		t.Fatalf("查询超管权限编码失败: %v", err)
	}
	if len(codes) != 1 || codes[0] != "*" {
		t.Fatalf("超管权限编码应为 [*]: %v", codes)
	}
	menus, err := svc.ListMenus(ctx, 1)
	if err != nil {
		t.Fatalf("查询超管菜单失败: %v", err)
	}
	if len(menus) != 2 {
		t.Fatalf("超管应看到全部启用菜单: %d", len(menus))
	}
}

func TestAuthServiceNormalUser(t *testing.T) {
	ctx := context.Background()
	repo := newFakeRepo()
	repo.roles[1] = &entity.Role{ID: 1, Code: "viewer", Status: enums.StatusEnabled}
	repo.roles[2] = &entity.Role{ID: 2, Code: "disabled_role", Status: enums.StatusDisabled}
	repo.permissions[10] = &entity.Permission{ID: 10, Code: "system:user", Type: enums.PermissionTypeMenu, UIPath: "/system/user", Status: enums.StatusEnabled}
	repo.permissions[11] = &entity.Permission{ID: 11, Code: "system:user:list", Type: enums.PermissionTypeAPI, HTTPMethod: "GET", HTTPPath: "/api/system/users", Status: enums.StatusEnabled}
	repo.permissions[12] = &entity.Permission{ID: 12, Code: "system:role", Type: enums.PermissionTypeMenu, UIPath: "/system/role", Status: enums.StatusEnabled}
	repo.userRoles[2] = []int64{1}
	repo.userRoles[3] = []int64{2}
	repo.rolePermissions[1] = []int64{10, 11}
	repo.rolePermissions[2] = []int64{12}
	svc := NewRbacAuthService(repo)

	allowed, err := svc.CanAccessAPI(ctx, 2, "GET", "/api/system/users")
	if err != nil || !allowed {
		t.Fatalf("授权接口应放行: allowed=%v err=%v", allowed, err)
	}
	allowed, _ = svc.CanAccessAPI(ctx, 2, "POST", "/api/system/users")
	if allowed {
		t.Fatal("未授权方法不应放行")
	}
	allowed, _ = svc.CanAccessAPI(ctx, 3, "GET", "/api/system/users")
	if allowed {
		t.Fatal("禁用角色不应放行")
	}

	codes, _ := svc.ListPermissionCodes(ctx, 2)
	if len(codes) != 2 {
		t.Fatalf("普通用户权限编码数量不符: %v", codes)
	}
	menus, _ := svc.ListMenus(ctx, 2)
	if len(menus) != 1 || menus[0].UIPath != "/system/user" {
		t.Fatalf("普通用户菜单不符: %+v", menus)
	}
	if menus, _ := svc.ListMenus(ctx, 3); len(menus) != 0 {
		t.Fatalf("禁用角色用户不应有菜单: %+v", menus)
	}

	allowed, _ = svc.CanAccessAPI(ctx, 0, "GET", "/api/system/users")
	if allowed {
		t.Fatal("非法用户不应放行")
	}
}

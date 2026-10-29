// Package test RBAC 接口集成测试（须先启动 yomi-admin-backend 服务）
package test

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"gitee.com/hq_project/yomi-boot/component/common/testclient"
)

// permissionItem 是权限列表项
type permissionItem struct {
	ID   string `json:"id"`
	Code string `json:"code"`
	Type string `json:"type"`
}

// TestRbacRoleFlow 覆盖权限组创建、权限组绑权限、角色绑组、用户绑角色与鉴权拒绝
func TestRbacRoleFlow(t *testing.T) {
	requireServer(t)
	adminToken := login(t, adminUsername, adminPassword)
	suffix := time.Now().UnixNano()
	roleCode := fmt.Sprintf("viewer_%d", suffix)
	groupCode := fmt.Sprintf("viewer_group_%d", suffix)
	roleName := "只读角色"
	groupName := "只读权限组"

	// 创建角色
	body, ok := testclient.Post(API("/system/role/create"), adminToken, map[string]any{
		"code":   roleCode,
		"name":   roleName,
		"remark": "集成测试角色",
	})
	if !ok {
		t.Fatalf("创建角色接口不可达: %s", body)
	}
	created := parseSuccess(t, body)
	var role struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(created.Data, &role); err != nil || role.ID == "" {
		t.Fatalf("解析角色 ID 失败: %v body=%s", err, body)
	}
	t.Cleanup(func() {
		_, _ = testclient.Post(API("/system/role/update"), adminToken, map[string]any{
			"id":     role.ID,
			"name":   roleName,
			"remark": "集成测试角色",
			"status": "DISABLED",
		})
	})

	// 查询权限目录，取出菜单与接口权限 ID
	body, ok = testclient.Get(API("/system/permission?page=1&pageSize=100"), adminToken)
	if !ok {
		t.Fatalf("权限列表接口不可达: %s", body)
	}
	permissionsPage := parseSuccess(t, body)
	var permissions []permissionItem
	if err := json.Unmarshal(permissionsPage.Data, &permissions); err != nil {
		t.Fatalf("解析权限列表失败: %v body=%s", err, body)
	}
	permissionIDs := make([]string, 0, 2)
	for _, permission := range permissions {
		if permission.Code == "system:user" || permission.Code == "system:user:list" {
			permissionIDs = append(permissionIDs, permission.ID)
		}
	}
	if len(permissionIDs) != 2 {
		t.Fatalf("未找到预期权限: %+v", permissions)
	}

	// 创建权限组并绑定权限
	body, ok = testclient.Post(API("/system/permission-group/create"), adminToken, map[string]any{
		"code":   groupCode,
		"name":   groupName,
		"remark": "集成测试权限组",
	})
	if !ok {
		t.Fatalf("创建权限组接口不可达: %s", body)
	}
	groupCreated := parseSuccess(t, body)
	var group struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(groupCreated.Data, &group); err != nil || group.ID == "" {
		t.Fatalf("解析权限组 ID 失败: %v body=%s", err, body)
	}
	t.Cleanup(func() {
		_, _ = testclient.Post(API("/system/permission-group/update"), adminToken, map[string]any{
			"id":     group.ID,
			"name":   groupName,
			"remark": "集成测试权限组",
			"status": "DISABLED",
		})
	})

	body, ok = testclient.Post(API("/system/permission-group/bind-permission"), adminToken, map[string]any{
		"groupId":       group.ID,
		"permissionIds": permissionIDs,
	})
	if !ok {
		t.Fatalf("权限组绑定权限接口不可达: %s", body)
	}
	parseSuccess(t, body)

	// 角色绑定权限组
	body, ok = testclient.Post(API("/system/role/bind-group"), adminToken, map[string]any{
		"roleId":   role.ID,
		"groupIds": []string{group.ID},
	})
	if !ok {
		t.Fatalf("角色绑定权限组接口不可达: %s", body)
	}
	parseSuccess(t, body)

	// 角色详情返回已绑组与聚合权限
	body, ok = testclient.Get(API("/system/role/detail?id="+role.ID), adminToken)
	if !ok {
		t.Fatalf("角色详情接口不可达: %s", body)
	}
	roleDetail := parseSuccess(t, body)
	var detail struct {
		GroupIDs      []string `json:"groupIds"`
		PermissionIDs []string `json:"permissionIds"`
	}
	if err := json.Unmarshal(roleDetail.Data, &detail); err != nil {
		t.Fatalf("解析角色详情失败: %v body=%s", err, body)
	}
	if len(detail.GroupIDs) != 1 || detail.GroupIDs[0] != group.ID {
		t.Fatalf("角色权限组不符: %+v", detail.GroupIDs)
	}
	if len(detail.PermissionIDs) != 2 {
		t.Fatalf("角色聚合权限不符: %+v", detail.PermissionIDs)
	}

	// 创建用户并绑定角色
	username := fmt.Sprintf("rbac_%d", suffix)
	body, ok = testclient.Post(API("/system/user/create"), adminToken, map[string]any{
		"username": username,
		"password": "test1234",
		"nickname": "RBAC 测试用户",
		"email":    "",
		"phone":    "",
		"status":   "ENABLED",
	})
	if !ok {
		t.Fatalf("创建用户接口不可达: %s", body)
	}
	parseSuccess(t, body)

	userID := findUserID(t, adminToken, username)
	t.Cleanup(func() {
		_, _ = testclient.Post(API("/system/user/delete"), adminToken, map[string]any{"id": userID})
	})

	body, ok = testclient.Post(API("/system/user/bind-role"), adminToken, map[string]any{
		"userId":  userID,
		"roleIds": []string{role.ID},
	})
	if !ok {
		t.Fatalf("用户绑定角色接口不可达: %s", body)
	}
	parseSuccess(t, body)

	// 受限用户登录后：有用户菜单与列表权限，无角色管理权限
	userToken := login(t, username, "test1234")
	body, ok = testclient.Get(API("/auth/profile"), userToken)
	if !ok {
		t.Fatalf("受限用户 profile 不可达: %s", body)
	}
	profile := parseSuccess(t, body)
	var profileData struct {
		Permissions []string `json:"permissions"`
		Menus       []struct {
			UIPath string `json:"uiPath"`
		} `json:"menus"`
	}
	if err := json.Unmarshal(profile.Data, &profileData); err != nil {
		t.Fatalf("解析受限用户 profile 失败: %v body=%s", err, body)
	}
	menuPaths := make(map[string]struct{}, len(profileData.Menus))
	for _, menu := range profileData.Menus {
		menuPaths[menu.UIPath] = struct{}{}
	}
	if _, exists := menuPaths["/system/user"]; !exists {
		t.Fatalf("受限用户缺少系统用户菜单: %+v", profileData.Menus)
	}
	if _, exists := menuPaths["/system/role"]; exists {
		t.Fatalf("受限用户不应看到角色菜单: %+v", profileData.Menus)
	}

	body, ok = testclient.Get(API("/system/user?page=1&pageSize=10"), userToken)
	if !ok {
		t.Fatalf("受限用户访问已授权接口不可达: %s", body)
	}
	parseSuccess(t, body)

	body, ok = testclient.Get(API("/system/role?page=1&pageSize=10"), userToken)
	if !ok {
		t.Fatalf("受限用户访问未授权接口不可达: %s", body)
	}
	assertFailure(t, body)
}

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

// TestRbacRoleFlow 覆盖角色创建、绑定权限、用户绑定角色与鉴权拒绝
func TestRbacRoleFlow(t *testing.T) {
	requireServer(t)
	adminToken := login(t, adminUsername, adminPassword)
	suffix := time.Now().UnixNano()
	roleCode := fmt.Sprintf("viewer_%d", suffix)
	roleName := "只读角色"

	// 创建角色
	body, ok := testclient.Post(API("/system/role"), adminToken, map[string]any{
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
		_, _ = testclient.Put(API("/system/role/"+role.ID), adminToken, map[string]any{
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

	// 角色绑定权限
	body, ok = testclient.Put(API("/system/role/"+role.ID+"/permission"), adminToken, map[string]any{
		"permissionIds": permissionIDs,
	})
	if !ok {
		t.Fatalf("角色绑定权限接口不可达: %s", body)
	}
	parseSuccess(t, body)

	// 创建用户并绑定角色
	username := fmt.Sprintf("rbac_%d", suffix)
	body, ok = testclient.Post(API("/system/users"), adminToken, map[string]any{
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
		_, _ = testclient.Delete(API("/system/users/"+userID), adminToken)
	})

	body, ok = testclient.Put(API("/system/user/"+userID+"/role"), adminToken, map[string]any{
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

	body, ok = testclient.Get(API("/system/users?page=1&pageSize=10"), userToken)
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

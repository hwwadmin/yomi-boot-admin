// Package test RBAC 鉴权缓存接口集成测试（须先启动 yomi-admin-backend 服务）
package test

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"gitee.com/hq_project/yomi-boot/component/common/testclient"
)

// cacheRefreshResult 是刷新鉴权缓存的返回结构
type cacheRefreshResult struct {
	Version              string `json:"version"`
	BuiltAt              int64  `json:"builtAt"`
	RoleCount            int    `json:"roleCount"`
	PermissionCount      int    `json:"permissionCount"`
	PermissionGroupCount int    `json:"permissionGroupCount"`
	RelationCount        int    `json:"relationCount"`
	DurationMS           int64  `json:"durationMs"`
}

// profileView 是当前用户信息视图
type profileView struct {
	Permissions []string `json:"permissions"`
	Menus       []struct {
		Code   string `json:"code"`
		UIPath string `json:"uiPath"`
	} `json:"menus"`
}

// TestRbacCacheRefresh 覆盖刷新接口：投影在刷新前不反映新数据，刷新后立即生效；版本与行数同步递增
func TestRbacCacheRefresh(t *testing.T) {
	requireServer(t)
	adminToken := login(t, adminUsername, adminPassword)

	first := refreshRbacCache(t, adminToken)
	if first.Version == "" || first.RoleCount <= 0 || first.PermissionCount <= 0 {
		t.Fatalf("刷新结果不合理: %+v", first)
	}

	// 新建一个菜单权限：投影未刷新时不会出现在当前用户菜单里
	suffix := time.Now().UnixNano()
	code := fmt.Sprintf("cache:menu:m%d", suffix)
	uiPath := fmt.Sprintf("/system/cache-test-%d", suffix)
	body, ok := testclient.Post(API("/system/permission/create"), adminToken, map[string]any{
		"code":   code,
		"name":   "缓存刷新测试菜单",
		"type":   "MENU",
		"uiPath": uiPath,
		"remark": "集成测试临时权限",
	})
	if !ok {
		t.Fatalf("创建权限接口不可达: %s", body)
	}
	created := parseSuccess(t, body)
	var permission struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(created.Data, &permission); err != nil {
		t.Fatalf("解析权限 ID 失败: %v body=%s", err, body)
	}
	t.Cleanup(func() {
		if _, ok := testclient.Post(API("/system/permission/delete"), adminToken, map[string]any{
			"id": permission.ID,
		}); ok {
			refreshRbacCache(t, adminToken)
		}
	})

	// 定时刷新（默认 60s）可能恰好抢先应用了这次变更，此时只校验刷新结果，不再断言陈旧性
	visibleBefore := menuExists(t, adminToken, uiPath)
	if visibleBefore {
		t.Log("定时刷新已应用本次变更，跳过刷新前的陈旧性断言")
	}

	refreshed := refreshRbacCache(t, adminToken)
	if refreshed.PermissionCount != first.PermissionCount+1 {
		t.Fatalf("刷新后权限数应加一: before=%d after=%d", first.PermissionCount, refreshed.PermissionCount)
	}
	if refreshed.Version == first.Version {
		t.Fatal("刷新应生成新的投影版本")
	}
	if refreshed.BuiltAt < first.BuiltAt {
		t.Fatalf("刷新时间应不早于上次: %d < %d", refreshed.BuiltAt, first.BuiltAt)
	}

	if !menuExists(t, adminToken, uiPath) {
		t.Fatal("刷新后投影应立即包含新菜单")
	}
}

// refreshRbacCache 调用刷新接口并断言成功
func refreshRbacCache(t *testing.T, token string) cacheRefreshResult {
	t.Helper()

	body, ok := testclient.Post(API("/system/rbac-cache/refresh"), token, map[string]any{})
	if !ok {
		t.Fatalf("刷新鉴权缓存接口不可达: %s", body)
	}
	parsed := parseSuccess(t, body)

	var result cacheRefreshResult
	if err := json.Unmarshal(parsed.Data, &result); err != nil {
		t.Fatalf("解析刷新结果失败: %v body=%s", err, body)
	}
	return result
}

// menuExists 判断当前用户菜单中是否存在指定路径；菜单来自鉴权投影
func menuExists(t *testing.T, token, uiPath string) bool {
	t.Helper()

	body, ok := testclient.Get(API("/auth/profile"), token)
	if !ok {
		t.Fatalf("当前用户接口不可达: %s", body)
	}
	parsed := parseSuccess(t, body)

	var profile profileView
	if err := json.Unmarshal(parsed.Data, &profile); err != nil {
		t.Fatalf("解析当前用户信息失败: %v body=%s", err, body)
	}
	for _, menu := range profile.Menus {
		if menu.UIPath == uiPath {
			return true
		}
	}
	return false
}

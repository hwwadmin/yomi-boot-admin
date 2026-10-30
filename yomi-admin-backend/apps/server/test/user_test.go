package test

import (
	"encoding/json"
	"fmt"
	"net/url"
	"testing"
	"time"

	"gitee.com/hq_project/yomi-boot/component/common/testclient"
)

type userItem struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Nickname string `json:"nickname"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Status   string `json:"status"`
	Roles    []struct {
		ID   string `json:"id"`
		Code string `json:"code"`
		Name string `json:"name"`
	} `json:"roles"`
	CreatedAt int64 `json:"createdAt"`
}

func TestProfileAndUserCRUD(t *testing.T) {
	requireServer(t)
	token := login(t, adminUsername, adminPassword)

	// profile
	body, ok := testclient.Get(API("/auth/profile"), token)
	if !ok {
		t.Fatalf("profile 不可达: %s", body)
	}
	profile := parseSuccess(t, body)
	var profileData struct {
		Username    string   `json:"username"`
		Nickname    string   `json:"nickname"`
		Permissions []string `json:"permissions"`
		Roles       []struct {
			Code   string `json:"code"`
			Name   string `json:"name"`
			Status string `json:"status"`
		} `json:"roles"`
		Menus []struct {
			Code   string `json:"code"`
			Name   string `json:"name"`
			UIPath string `json:"uiPath"`
		} `json:"menus"`
	}
	if err := json.Unmarshal(profile.Data, &profileData); err != nil {
		t.Fatalf("解析 profile 失败: %v body=%s", err, body)
	}
	if profileData.Username != adminUsername || len(profileData.Permissions) == 0 ||
		len(profileData.Roles) == 0 || len(profileData.Menus) == 0 {
		t.Fatalf("profile 内容不符合预期: %+v", profileData)
	}

	// 初始分页
	body, ok = testclient.Get(API("/system/user?page=1&pageSize=10"), token)
	if !ok {
		t.Fatalf("用户分页不可达: %s", body)
	}
	page := parsePage[userItem](t, body)
	if page.Total == 0 {
		t.Fatalf("内置管理员应出现在用户列表中: %s", body)
	}

	// 创建用户
	username := fmt.Sprintf("test_%d", time.Now().UnixNano())
	createReq := map[string]any{
		"username": username,
		"password": "test1234",
		"nickname": "测试用户",
		"email":    "test@example.com",
		"phone":    "13800138000",
		"status":   "ENABLED",
	}
	body, ok = testclient.Post(API("/system/user/create"), token, createReq)
	if !ok {
		t.Fatalf("创建用户接口不可达: %s", body)
	}
	parseSuccess(t, body)

	// 按关键字分页
	query := url.Values{}
	query.Set("page", "1")
	query.Set("pageSize", "10")
	query.Set("keyword", username)
	body, ok = testclient.Get(API("/system/user?"+query.Encode()), token)
	if !ok {
		t.Fatalf("用户分页不可达: %s", body)
	}
	page = parsePage[userItem](t, body)
	items := page.Items
	if len(items) != 1 || items[0].Username != username {
		t.Fatalf("按用户名搜索未命中: %+v", items)
	}
	userID := items[0].ID
	if items[0].CreatedAt == 0 {
		t.Fatal("createdAt 不应为空")
	}

	// 更新用户
	updateReq := map[string]any{
		"id":       userID,
		"nickname": "更新后的昵称",
		"email":    "updated@example.com",
		"phone":    "13900139000",
		"status":   "ENABLED",
	}
	body, ok = testclient.Post(API("/system/user/update"), token, updateReq)
	if !ok {
		t.Fatalf("更新用户接口不可达: %s", body)
	}
	parseSuccess(t, body)

	body, ok = testclient.Get(API("/system/user?"+query.Encode()), token)
	if !ok {
		t.Fatalf("更新后分页不可达: %s", body)
	}
	page = parsePage[userItem](t, body)
	items = page.Items
	if len(items) != 1 || items[0].Nickname != "更新后的昵称" || items[0].Email != "updated@example.com" {
		t.Fatalf("更新未生效: %+v", items)
	}

	// 删除用户
	body, ok = testclient.Post(API("/system/user/delete"), token, map[string]any{"id": userID})
	if !ok {
		t.Fatalf("删除用户接口不可达: %s", body)
	}
	parseSuccess(t, body)

	body, ok = testclient.Get(API("/system/user?"+query.Encode()), token)
	if !ok {
		t.Fatalf("删除后分页不可达: %s", body)
	}
	page = parsePage[userItem](t, body)
	if page.Total != 0 {
		t.Fatalf("软删除后仍能查到用户: %s", body)
	}
}

func TestUpdateOwnProfile(t *testing.T) {
	requireServer(t)
	token := login(t, adminUsername, adminPassword)

	body, ok := testclient.Get(API("/auth/profile"), token)
	if !ok {
		t.Fatalf("profile 不可达: %s", body)
	}
	profile := parseSuccess(t, body)
	var before struct {
		Nickname string `json:"nickname"`
		Email    string `json:"email"`
		Phone    string `json:"phone"`
	}
	if err := json.Unmarshal(profile.Data, &before); err != nil {
		t.Fatalf("解析 profile 失败: %v body=%s", err, body)
	}
	defer func() {
		body, ok := testclient.Post(API("/auth/profile"), token, map[string]any{
			"nickname": before.Nickname,
			"email":    before.Email,
			"phone":    before.Phone,
		})
		if !ok {
			t.Fatalf("恢复管理员资料不可达: %s", body)
		}
	}()

	body, ok = testclient.Post(API("/auth/profile"), token, map[string]any{
		"nickname": "资料更新测试",
		"email":    "profile@example.com",
		"phone":    "13900139001",
	})
	if !ok {
		t.Fatalf("更新个人资料接口不可达: %s", body)
	}
	updated := parseSuccess(t, body)
	var after struct {
		Username string `json:"username"`
		Nickname string `json:"nickname"`
		Email    string `json:"email"`
		Phone    string `json:"phone"`
	}
	if err := json.Unmarshal(updated.Data, &after); err != nil {
		t.Fatalf("解析更新结果失败: %v body=%s", err, body)
	}
	if after.Username != adminUsername || after.Nickname != "资料更新测试" ||
		after.Email != "profile@example.com" || after.Phone != "13900139001" {
		t.Fatalf("个人资料更新结果不符合预期: %+v", after)
	}

	body, ok = testclient.Get(API("/auth/profile"), token)
	if !ok {
		t.Fatalf("更新后 profile 不可达: %s", body)
	}
	profile = parseSuccess(t, body)
	if err := json.Unmarshal(profile.Data, &after); err != nil {
		t.Fatalf("解析更新后 profile 失败: %v body=%s", err, body)
	}
	if after.Nickname != "资料更新测试" || after.Email != "profile@example.com" {
		t.Fatalf("个人资料未持久化: %+v", after)
	}
}

func TestDisableUserRevokesToken(t *testing.T) {
	requireServer(t)
	adminToken := login(t, adminUsername, adminPassword)

	username := fmt.Sprintf("revoke_%d", time.Now().UnixNano())
	createReq := map[string]any{
		"username": username,
		"password": "test1234",
		"nickname": "待禁用用户",
		"email":    "",
		"phone":    "",
		"status":   "ENABLED",
	}
	body, ok := testclient.Post(API("/system/user/create"), adminToken, createReq)
	if !ok {
		t.Fatalf("创建用户接口不可达: %s", body)
	}
	parseSuccess(t, body)

	userToken := login(t, username, "test1234")
	body, ok = testclient.Get(API("/auth/profile"), userToken)
	if !ok {
		t.Fatalf("用户 profile 不可达: %s", body)
	}
	parseSuccess(t, body)

	userID := findUserID(t, adminToken, username)

	body, ok = testclient.Post(API("/system/user/update"), adminToken, map[string]any{
		"id":       userID,
		"nickname": "待禁用用户",
		"email":    "",
		"phone":    "",
		"status":   "DISABLED",
	})
	if !ok {
		t.Fatalf("禁用用户接口不可达: %s", body)
	}
	parseSuccess(t, body)

	body, ok = testclient.Get(API("/auth/profile"), userToken)
	if !ok {
		t.Fatalf("禁用后 profile 不可达: %s", body)
	}
	assertFailure(t, body)

	body, ok = testclient.Post(API("/system/user/delete"), adminToken, map[string]any{"id": userID})
	if !ok {
		t.Fatalf("删除用户接口不可达: %s", body)
	}
	parseSuccess(t, body)
}

func TestLogoutInvalidatesToken(t *testing.T) {
	requireServer(t)
	token := login(t, adminUsername, adminPassword)

	body, ok := testclient.Post(API("/auth/logout"), token, nil)
	if !ok {
		t.Fatalf("登出接口不可达: %s", body)
	}
	parseSuccess(t, body)

	body, ok = testclient.Get(API("/auth/profile"), token)
	if !ok {
		t.Fatalf("登出后 profile 不可达: %s", body)
	}
	assertFailure(t, body)
}

// findUserID 通过用户名搜索用户 ID
func findUserID(t *testing.T, token, username string) string {
	t.Helper()

	query := url.Values{}
	query.Set("page", "1")
	query.Set("pageSize", "10")
	query.Set("keyword", username)

	body, ok := testclient.Get(API("/system/user?"+query.Encode()), token)
	if !ok {
		t.Fatalf("用户分页不可达: %s", body)
	}
	page := parsePage[userItem](t, body)
	if len(page.Items) != 1 {
		t.Fatalf("按用户名未唯一命中: %+v", page.Items)
	}
	return page.Items[0].ID
}

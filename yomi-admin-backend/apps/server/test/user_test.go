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
	ID        string `json:"id"`
	Username  string `json:"username"`
	Nickname  string `json:"nickname"`
	Email     string `json:"email"`
	Phone     string `json:"phone"`
	Status    int8   `json:"status"`
	CreatedAt int64  `json:"createdAt"`
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
	}
	if err := json.Unmarshal(profile.Data, &profileData); err != nil {
		t.Fatalf("解析 profile 失败: %v body=%s", err, body)
	}
	if profileData.Username != adminUsername || len(profileData.Permissions) == 0 {
		t.Fatalf("profile 内容不符合预期: %+v", profileData)
	}

	// 初始分页
	body, ok = testclient.Get(API("/system/users?page=1&pageSize=10"), token)
	if !ok {
		t.Fatalf("用户分页不可达: %s", body)
	}
	page := parseSuccess(t, body)
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
		"status":   1,
	}
	body, ok = testclient.Post(API("/system/users"), token, createReq)
	if !ok {
		t.Fatalf("创建用户接口不可达: %s", body)
	}
	parseSuccess(t, body)

	// 按关键字分页
	query := url.Values{}
	query.Set("page", "1")
	query.Set("pageSize", "10")
	query.Set("keyword", username)
	body, ok = testclient.Get(API("/system/users?"+query.Encode()), token)
	if !ok {
		t.Fatalf("用户分页不可达: %s", body)
	}
	page = parseSuccess(t, body)
	var items []userItem
	if err := json.Unmarshal(page.Data, &items); err != nil {
		t.Fatalf("解析用户列表失败: %v body=%s", err, body)
	}
	if len(items) != 1 || items[0].Username != username {
		t.Fatalf("按用户名搜索未命中: %+v", items)
	}
	userID := items[0].ID
	if items[0].CreatedAt == 0 {
		t.Fatal("createdAt 不应为空")
	}

	// 更新用户
	updateReq := map[string]any{
		"nickname": "更新后的昵称",
		"email":    "updated@example.com",
		"phone":    "13900139000",
		"status":   1,
	}
	body, ok = testclient.Put(API(fmt.Sprintf("/system/users/%s", userID)), token, updateReq)
	if !ok {
		t.Fatalf("更新用户接口不可达: %s", body)
	}
	parseSuccess(t, body)

	body, ok = testclient.Get(API("/system/users?"+query.Encode()), token)
	if !ok {
		t.Fatalf("更新后分页不可达: %s", body)
	}
	page = parseSuccess(t, body)
	if err := json.Unmarshal(page.Data, &items); err != nil {
		t.Fatalf("解析更新后用户列表失败: %v body=%s", err, body)
	}
	if len(items) != 1 || items[0].Nickname != "更新后的昵称" || items[0].Email != "updated@example.com" {
		t.Fatalf("更新未生效: %+v", items)
	}

	// 删除用户
	body, ok = testclient.Delete(API(fmt.Sprintf("/system/users/%s", userID)), token)
	if !ok {
		t.Fatalf("删除用户接口不可达: %s", body)
	}
	parseSuccess(t, body)

	body, ok = testclient.Get(API("/system/users?"+query.Encode()), token)
	if !ok {
		t.Fatalf("删除后分页不可达: %s", body)
	}
	page = parseSuccess(t, body)
	if page.Total != 0 {
		t.Fatalf("软删除后仍能查到用户: %s", body)
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
		"status":   1,
	}
	body, ok := testclient.Post(API("/system/users"), adminToken, createReq)
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

	body, ok = testclient.Put(API(fmt.Sprintf("/system/users/%s", userID)), adminToken, map[string]any{
		"nickname": "待禁用用户",
		"email":    "",
		"phone":    "",
		"status":   0,
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

	body, ok = testclient.Delete(API(fmt.Sprintf("/system/users/%s", userID)), adminToken)
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

	body, ok := testclient.Get(API("/system/users?"+query.Encode()), token)
	if !ok {
		t.Fatalf("用户分页不可达: %s", body)
	}
	page := parseSuccess(t, body)
	var items []userItem
	if err := json.Unmarshal(page.Data, &items); err != nil {
		t.Fatalf("解析用户列表失败: %v body=%s", err, body)
	}
	if len(items) != 1 {
		t.Fatalf("按用户名未唯一命中: %+v", items)
	}
	return items[0].ID
}

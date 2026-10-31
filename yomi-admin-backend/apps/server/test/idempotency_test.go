package test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"testing"
	"time"
)

// postWithIdempotencyKey 发送带幂等键的 POST，返回响应体
func postWithIdempotencyKey(t *testing.T, apiPath, token, key string, body any) string {
	t.Helper()

	payload, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("序列化请求体失败: %v", err)
	}
	req, err := http.NewRequest(http.MethodPost, API(apiPath), bytes.NewReader(payload))
	if err != nil {
		t.Fatalf("创建请求失败: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("X-Idempotency-Key", key)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("请求失败: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("读取响应失败: %v", err)
	}
	return string(respBody)
}

// TestCreateIdempotencyReplaysSameResult 验证同一幂等键重复创建只落一条数据，且回放首次结果
func TestCreateIdempotencyReplaysSameResult(t *testing.T) {
	requireServer(t)
	token := login(t, adminUsername, adminPassword)

	key := fmt.Sprintf("it-user-%d", time.Now().UnixNano())
	username := fmt.Sprintf("idem_%d", time.Now().UnixNano())
	createBody := map[string]any{
		"username": username,
		"password": "test1234",
		"nickname": "幂等用户",
		"email":    "idem@example.com",
		"phone":    "13800138000",
		"status":   "ENABLED",
	}

	// 同键提交两次：第二次应回放首次成功结果，不再创建
	for i := 0; i < 2; i++ {
		parseSuccess(t, postWithIdempotencyKey(t, "/system/user/create", token, key, createBody))
	}

	body, ok := httpGet(t, API("/system/user?page=1&pageSize=100&keyword="+url.QueryEscape(username)), token)
	if !ok {
		t.Fatalf("用户分页不可达: %s", body)
	}
	page := parsePage[userItem](t, body)
	if page.Total != 1 {
		t.Fatalf("同一幂等键应只创建一条用户, total = %d, body = %s", page.Total, body)
	}

	// 重复删除：两次都按成功处理
	userID := page.Items[0].ID
	for i := 0; i < 2; i++ {
		parseSuccess(t, postJSON(t, "/system/user/delete", token, map[string]any{"id": userID}))
	}
}

// TestCreateRoleIdempotencyReplaysItem 验证返回实体的创建接口重放时返回同一角色
func TestCreateRoleIdempotencyReplaysItem(t *testing.T) {
	requireServer(t)
	token := login(t, adminUsername, adminPassword)

	key := fmt.Sprintf("it-role-%d", time.Now().UnixNano())
	code := fmt.Sprintf("idem_role_%d", time.Now().UnixNano())
	createBody := map[string]any{"code": code, "name": "幂等角色", "remark": ""}

	var firstID, secondID string
	for i := 0; i < 2; i++ {
		parsed := parseSuccess(t, postWithIdempotencyKey(t, "/system/role/create", token, key, createBody))
		var item struct {
			ID string `json:"id"`
		}
		if err := json.Unmarshal(parsed.Data, &item); err != nil {
			t.Fatalf("解析创建角色结果失败: %v data=%s", err, parsed.Data)
		}
		if i == 0 {
			firstID = item.ID
		} else {
			secondID = item.ID
		}
	}
	if firstID == "" || firstID != secondID {
		t.Fatalf("重放应返回同一角色, first=%q second=%q", firstID, secondID)
	}

	// 重复删除：两次都按成功处理
	for i := 0; i < 2; i++ {
		parseSuccess(t, postJSON(t, "/system/role/delete", token, map[string]any{"id": firstID}))
	}
}

// TestChangePasswordIdempotencySkipsSecondChange 验证同键改密重放时不再次校验原密码、不重复写审计
func TestChangePasswordIdempotencySkipsSecondChange(t *testing.T) {
	requireServer(t)
	adminToken := login(t, adminUsername, adminPassword)

	username := fmt.Sprintf("pwd_%d", time.Now().UnixNano())
	oldPassword, newPassword := "oldpass123", "newpass123"
	createBody := map[string]any{
		"username": username,
		"password": oldPassword,
		"nickname": "幂等改密用户",
		"email":    "idem-pwd@example.com",
		"phone":    "13800138000",
		"status":   "ENABLED",
	}
	parseSuccess(t, postWithIdempotencyKey(t, "/system/user/create", adminToken, fmt.Sprintf("it-pwd-user-%d", time.Now().UnixNano()), createBody))

	// 清理临时用户（重复删除按成功处理）
	defer func() {
		body, ok := httpGet(t, API("/system/user?page=1&pageSize=10&keyword="+url.QueryEscape(username)), adminToken)
		if !ok {
			return
		}
		page := parsePage[userItem](t, body)
		for _, item := range page.Items {
			if item.Username == username {
				_ = postJSON(t, "/system/user/delete", adminToken, map[string]any{"id": item.ID})
			}
		}
	}()

	userToken := login(t, username, oldPassword)
	key := fmt.Sprintf("it-pwd-%d", time.Now().UnixNano())
	changeBody := map[string]any{"oldPassword": oldPassword, "newPassword": newPassword}
	parseSuccess(t, postWithIdempotencyKey(t, "/auth/password", userToken, key, changeBody))

	// 首次改密已撤销全部令牌，用新密码重新登录后再用同键重放
	userToken = login(t, username, newPassword)
	parseSuccess(t, postWithIdempotencyKey(t, "/auth/password", userToken, key, changeBody))

	// 审计是异步写入，等待消费后断言只有一对 STARTED/FINISHED
	time.Sleep(500 * time.Millisecond)
	now := time.Now().UnixMilli()
	auditURL := fmt.Sprintf(
		"/system/audit-log?page=1&pageSize=10&eventCode=auth.change_password&keyword=%s&startTime=%d&endTime=%d",
		url.QueryEscape(username), now-3600_000, now+60_000,
	)
	body, ok := httpGet(t, API(auditURL), adminToken)
	if !ok {
		t.Fatalf("审计日志不可达: %s", body)
	}
	page := parsePage[struct{}](t, body)
	if page.Total != 2 {
		t.Fatalf("同键重放不应重复写审计, total = %d, body = %s", page.Total, body)
	}
}

// httpGet 发送带 token 的 GET，返回响应体
func httpGet(t *testing.T, rawURL, token string) (string, bool) {
	t.Helper()

	req, err := http.NewRequest(http.MethodGet, rawURL, nil)
	if err != nil {
		t.Fatalf("创建请求失败: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "请求失败: " + err.Error(), false
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "读取响应失败: " + err.Error(), false
	}
	return string(respBody), resp.StatusCode == http.StatusOK
}

// postJSON 发送带 token 的普通 POST（不带幂等键），返回响应体
func postJSON(t *testing.T, apiPath, token string, body any) string {
	t.Helper()

	payload, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("序列化请求体失败: %v", err)
	}
	req, err := http.NewRequest(http.MethodPost, API(apiPath), bytes.NewReader(payload))
	if err != nil {
		t.Fatalf("创建请求失败: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("请求失败: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("读取响应失败: %v", err)
	}
	return string(respBody)
}

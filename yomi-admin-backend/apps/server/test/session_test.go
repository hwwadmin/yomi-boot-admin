package test

import (
	"encoding/json"
	"fmt"
	"net/url"
	"testing"
	"time"

	"github.com/hwwadmin/yomi-boot/component/core/response"
)

type userSessionItem struct {
	SessionID  string `json:"sessionId"`
	UserID     string `json:"userId"`
	Username   string `json:"username"`
	Nickname   string `json:"nickname"`
	ClientType string `json:"clientType"`
	LoginIP    string `json:"loginIp"`
	UserAgent  string `json:"userAgent"`
	LoginAt    int64  `json:"loginAt"`
	ExpireAt   int64  `json:"expireAt"`
	Current    bool   `json:"current"`
}

// TestUserSessionsListAndRevoke 覆盖按用户查看会话、单个下线、全部下线与自我保护
func TestUserSessionsListAndRevoke(t *testing.T) {
	requireServer(t)
	adminToken := login(t, adminUsername, adminPassword)

	username := fmt.Sprintf("session_%d", time.Now().UnixNano())
	body, ok := Post(API("/system/user/create"), adminToken, map[string]any{
		"username": username,
		"password": "test1234",
		"nickname": "会话测试用户",
		"email":    "",
		"phone":    "",
		"status":   "ENABLED",
	})
	if !ok {
		t.Fatalf("创建用户接口不可达: %s", body)
	}
	parseSuccess(t, body)

	userID := findUserID(t, adminToken, username)
	tokenA := login(t, username, "test1234")
	tokenB := login(t, username, "test1234")

	t.Cleanup(func() {
		Post(API("/system/user/delete"), adminToken, map[string]any{"id": userID})
	})

	sessions := listUserSessions(t, adminToken, userID)
	if len(sessions) != 2 {
		t.Fatalf("期望 2 条在线会话，实际 %d: %+v", len(sessions), sessions)
	}
	for _, session := range sessions {
		if session.SessionID == "" || session.UserID != userID || session.Username != username {
			t.Fatalf("会话字段不符: %+v", session)
		}
		if session.LoginAt <= 0 || session.ExpireAt <= session.LoginAt {
			t.Fatalf("会话时间字段不符: %+v", session)
		}
		if session.Current {
			t.Fatalf("查看他人会话不应标记 current: %+v", session)
		}
	}

	// 单个下线：只应让其中一个令牌失效
	body, ok = Post(API("/system/user/session/revoke"), adminToken, map[string]any{
		"userId":    userID,
		"sessionId": sessions[0].SessionID,
	})
	if !ok {
		t.Fatalf("下线会话接口不可达: %s", body)
	}
	parseSuccess(t, body)

	okA := profileOK(tokenA)
	okB := profileOK(tokenB)
	if okA == okB {
		t.Fatalf("单个下线后应恰好一个令牌失效: A=%v B=%v", okA, okB)
	}
	if remaining := listUserSessions(t, adminToken, userID); len(remaining) != 1 {
		t.Fatalf("单个下线后应剩 1 条会话，实际 %d", len(remaining))
	}

	// 不能踢自己的当前会话
	adminID := findUserID(t, adminToken, adminUsername)
	var currentSessionID string
	for _, session := range listUserSessions(t, adminToken, adminID) {
		if session.Current {
			currentSessionID = session.SessionID
			break
		}
	}
	if currentSessionID == "" {
		t.Fatal("管理员会话列表中应包含 current 会话")
	}
	body, _ = Post(API("/system/user/session/revoke"), adminToken, map[string]any{
		"userId":    adminID,
		"sessionId": currentSessionID,
	})
	if code := errorCode(t, body); code != "ADMIN_CANNOT_OPERATE_SELF" {
		t.Fatalf("踢自己当前会话期望 ADMIN_CANNOT_OPERATE_SELF，实际 %s body=%s", code, body)
	}

	// 全部下线：两个令牌都应失效
	body, ok = Post(API("/system/user/session/revoke-all"), adminToken, map[string]any{
		"userId": userID,
	})
	if !ok {
		t.Fatalf("全部下线接口不可达: %s", body)
	}
	parseSuccess(t, body)
	if profileOK(tokenA) || profileOK(tokenB) {
		t.Fatal("全部下线后令牌仍可用")
	}
	if remaining := listUserSessions(t, adminToken, userID); len(remaining) != 0 {
		t.Fatalf("全部下线后仍有 %d 条会话", len(remaining))
	}

	// 不能按用户踢自己
	body, _ = Post(API("/system/user/session/revoke-all"), adminToken, map[string]any{
		"userId": adminID,
	})
	if code := errorCode(t, body); code != "ADMIN_CANNOT_OPERATE_SELF" {
		t.Fatalf("按用户踢自己期望 ADMIN_CANNOT_OPERATE_SELF，实际 %s body=%s", code, body)
	}
}

// listUserSessions 查询指定用户的在线会话
func listUserSessions(t *testing.T, token, userID string) []userSessionItem {
	t.Helper()

	query := url.Values{}
	query.Set("userId", userID)
	body, ok := Get(API("/system/user/session?"+query.Encode()), token)
	if !ok {
		t.Fatalf("会话列表接口不可达: %s", body)
	}
	parsed := parseSuccess(t, body)

	var sessions []userSessionItem
	if err := json.Unmarshal(parsed.Data, &sessions); err != nil {
		t.Fatalf("解析会话列表失败: %v body=%s", err, body)
	}
	return sessions
}

// profileOK 判断令牌是否仍可访问当前用户接口
func profileOK(token string) bool {
	body, ok := Get(API("/auth/profile"), token)
	if !ok {
		return false
	}
	parsed, ok := response.Enc().Parse([]byte(body))
	return ok && response.Enc().IsSuccess(parsed.Code)
}

// errorCode 解析失败响应的业务码
func errorCode(t *testing.T, body string) string {
	t.Helper()

	parsed, ok := response.Enc().Parse([]byte(body))
	if !ok {
		t.Fatalf("响应不符合默认协议: %s", body)
	}
	if response.Enc().IsSuccess(parsed.Code) {
		t.Fatalf("期望失败响应，实际成功: %s", body)
	}
	return parsed.Code
}

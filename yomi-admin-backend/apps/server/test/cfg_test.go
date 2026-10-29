// Package test 接口集成测试（须先启动 yomi-admin-backend 服务）
package test

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"gitee.com/hq_project/yomi-boot/component/common/testclient"
	"gitee.com/hq_project/yomi-boot/component/core/response"
)

func TestMain(m *testing.M) {
	response.MustRegister(response.DefaultEncoder{})
	os.Exit(m.Run())
}

var (
	baseURL       = envOrDefault("YOMI_TEST_BASE_URL", "http://127.0.0.1:8080")
	adminUsername = envOrDefault("YOMI_TEST_ADMIN_USERNAME", "admin")
	adminPassword = envOrDefault("YOMI_TEST_ADMIN_PASSWORD", "admin123")
)

// API 拼接服务地址与路由前缀
func API(path string) string {
	return strings.TrimRight(baseURL, "/") + "/api" + path
}

// requireServer 服务不可达时跳过集成测试
func requireServer(t *testing.T) {
	t.Helper()
	if _, ok := testclient.Get(API("/v1/health"), ""); !ok {
		t.Skip("服务未启动或不可达，跳过集成测试")
	}
}

// captchaFailureCodes 表示登录需要图像验证码的业务码
var captchaFailureCodes = map[string]struct{}{
	"ADMIN_CAPTCHA_REQUIRED": {},
	"ADMIN_CAPTCHA_INVALID":  {},
	"ADMIN_CAPTCHA_MISMATCH": {},
}

// login 登录并返回 Token
func login(t *testing.T, username, password string) string {
	t.Helper()

	body, ok := testclient.Post(API("/auth/login"), "", map[string]string{
		"username": username,
		"password": password,
	})
	if !ok {
		t.Fatalf("登录接口不可达: %s", body)
	}

	parsed, ok := response.Enc().Parse([]byte(body))
	if !ok {
		t.Fatalf("响应不符合默认协议: %s", body)
	}
	if _, required := captchaFailureCodes[parsed.Code]; required {
		t.Skip("服务启用了登录图像验证码：请以 biz.captcha.enabled=false（或 YOMI_ADMIN_BIZ_CAPTCHA_ENABLED=false）启动服务后运行集成测试")
	}
	if !response.Enc().IsSuccess(parsed.Code) {
		t.Fatalf("登录失败 code=%s body=%s", parsed.Code, body)
	}

	var data struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(parsed.Data, &data); err != nil {
		t.Fatalf("解析登录结果失败: %v body=%s", err, body)
	}
	if data.Token == "" {
		t.Fatalf("登录未返回 token: %s", body)
	}
	return data.Token
}

// parseSuccess 断言响应为默认协议的成功响应
func parseSuccess(t *testing.T, body string) response.Response {
	t.Helper()

	parsed, ok := response.Enc().Parse([]byte(body))
	if !ok {
		t.Fatalf("响应不符合默认协议: %s", body)
	}
	if !response.Enc().IsSuccess(parsed.Code) {
		t.Fatalf("业务失败 code=%s body=%s", parsed.Code, body)
	}
	return parsed
}

// pageData 分页响应 data 载荷：{items,total}
type pageData[T any] struct {
	Items []T   `json:"items"`
	Total int64 `json:"total"`
}

// parsePage 断言响应为默认协议的成功分页响应并解析 data 载荷
func parsePage[T any](t *testing.T, body string) pageData[T] {
	t.Helper()

	parsed := parseSuccess(t, body)
	var data pageData[T]
	if err := json.Unmarshal(parsed.Data, &data); err != nil {
		t.Fatalf("解析分页 data 失败: %v body=%s", err, body)
	}
	return data
}

// assertFailure 断言响应为默认协议的失败响应
func assertFailure(t *testing.T, body string) {
	t.Helper()

	parsed, ok := response.Enc().Parse([]byte(body))
	if !ok {
		t.Fatalf("响应不符合默认协议: %s", body)
	}
	if response.Enc().IsSuccess(parsed.Code) {
		t.Fatalf("期望失败响应，实际成功: %s", body)
	}
}

func envOrDefault(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

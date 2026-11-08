package test

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/hwwadmin/yomi-boot/component/core/response"
)

// TestLoginGuardLocksRepeatedFailures 验证连续凭证失败后触发临时锁定。
//
// 该用例会消耗当前 IP 的失败额度，默认跳过；设置 YOMI_TEST_LOGIN_GUARD=1 运行。
// 运行前需以 biz.captcha.enabled=false、biz.login_guard.enabled=true 启动服务。
func TestLoginGuardLocksRepeatedFailures(t *testing.T) {
	if os.Getenv("YOMI_TEST_LOGIN_GUARD") != "1" {
		t.Skip("设置 YOMI_TEST_LOGIN_GUARD=1 运行登录保护集成测试")
	}
	requireServer(t)

	username := fmt.Sprintf("guard_probe_%d", time.Now().UnixNano())
	lastCode := ""
	for i := 0; i < 5; i++ {
		body, ok := Post(API("/auth/login"), "", map[string]string{
			"username": username,
			"password": "wrong-password",
		})
		if !ok {
			t.Fatalf("登录接口不可达: %s", body)
		}
		parsed, ok := response.Enc().Parse([]byte(body))
		if !ok {
			t.Fatalf("响应不符合默认协议: %s", body)
		}
		if _, required := captchaFailureCodes[parsed.Code]; required {
			t.Skip("服务启用了登录图像验证码：请以 biz.captcha.enabled=false 启动服务")
		}
		lastCode = parsed.Code
	}

	if lastCode != "ADMIN_LOGIN_LOCKED" {
		t.Fatalf("连续 5 次失败后期望 ADMIN_LOGIN_LOCKED，实际 %s", lastCode)
	}
}

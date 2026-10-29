package auth

import "testing"

func TestLoginConfigReflectsCaptchaSwitch(t *testing.T) {
	if got := NewLoginConfigUsecase(true).Execute(); !got.CaptchaRequired {
		t.Fatalf("验证码开启时应返回 captchaRequired=true: %+v", got)
	}
	if got := NewLoginConfigUsecase(false).Execute(); got.CaptchaRequired {
		t.Fatalf("验证码关闭时应返回 captchaRequired=false: %+v", got)
	}
}

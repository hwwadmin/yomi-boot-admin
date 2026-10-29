package config

import (
	"testing"
	"time"

	"gitee.com/hq_project/yomi-boot/component/common/captcha"
	"gitee.com/hq_project/yomi-boot/component/core/errors"
)

func TestCaptchaConfigDefaults(t *testing.T) {
	var cfg CaptchaConfig
	if !cfg.EnabledOrDefault() {
		t.Fatal("未配置 enabled 时应默认启用")
	}
	if got := cfg.KeyPrefixOrDefault(); got != defaultCaptchaKeyPrefix {
		t.Fatalf("默认前缀不符: %s", got)
	}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("零值配置应合法: %v", err)
	}
}

func TestCaptchaConfigToConfig(t *testing.T) {
	cfg := CaptchaConfig{Driver: "string", Length: 5, Width: 160, Height: 60, TTLSeconds: 180}

	runtime := cfg.ToConfig()
	if runtime.Driver != captcha.DriverString || runtime.Length != 5 ||
		runtime.Width != 160 || runtime.Height != 60 || runtime.TTL != 180*time.Second {
		t.Fatalf("运行时配置不符: %+v", runtime)
	}
}

func TestCaptchaConfigRejectsInvalidDriver(t *testing.T) {
	cfg := CaptchaConfig{Driver: "audio"}

	if err := cfg.Validate(); !errors.Is(err, captcha.ErrInvalidArgument) {
		t.Fatalf("期望 ErrInvalidArgument，实际 %v", err)
	}
}

func TestCaptchaConfigRejectsInvalidTTL(t *testing.T) {
	cfg := CaptchaConfig{TTLSeconds: 10}

	if err := cfg.Validate(); !errors.Is(err, captcha.ErrInvalidArgument) {
		t.Fatalf("期望 ErrInvalidArgument，实际 %v", err)
	}
}

func TestCaptchaConfigEnabledFalse(t *testing.T) {
	disabled := false
	cfg := CaptchaConfig{Enabled: &disabled}

	if cfg.EnabledOrDefault() {
		t.Fatal("enabled=false 时应关闭")
	}
}

func TestCaptchaConfigRefreshInterval(t *testing.T) {
	var cfg CaptchaConfig
	if got := cfg.RefreshInterval(); got != 10*time.Second {
		t.Fatalf("默认刷新间隔应为 10s，实际 %s", got)
	}

	zero := 0
	cfg.RefreshIntervalSeconds = &zero
	if got := cfg.RefreshInterval(); got != 0 {
		t.Fatalf("显式 0 应关闭限流，实际 %s", got)
	}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("0 应合法: %v", err)
	}
}

func TestCaptchaConfigRejectsInvalidRefreshInterval(t *testing.T) {
	negative := -1
	cfg := CaptchaConfig{RefreshIntervalSeconds: &negative}
	if err := cfg.Validate(); err == nil {
		t.Fatal("负间隔应校验失败")
	}

	tooLarge := maxCaptchaRefreshIntervalSeconds + 1
	cfg = CaptchaConfig{RefreshIntervalSeconds: &tooLarge}
	if err := cfg.Validate(); err == nil {
		t.Fatal("超过上限的间隔应校验失败")
	}
}

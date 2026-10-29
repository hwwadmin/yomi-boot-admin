package config

import (
	"testing"
	"time"
	"yomi-admin-backend/apps/server/internal/support/loginguard"
)

func TestLoginGuardConfigDefaults(t *testing.T) {
	policy := LoginGuardConfig{}.ToPolicy()
	want := loginguard.DefaultPolicy()

	if !policy.Enabled {
		t.Fatal("未配置 enabled 时应默认启用")
	}
	if policy.KeyPrefix != want.KeyPrefix {
		t.Fatalf("默认前缀不符: %s", policy.KeyPrefix)
	}
	if policy.IPAttempts != want.IPAttempts || policy.IPWindow != want.IPWindow {
		t.Fatalf("IP 尝试配额默认值不符: %+v", policy)
	}
	if policy.PairFailures != want.PairFailures || policy.PairLock != want.PairLock {
		t.Fatalf("组合失败预算默认值不符: %+v", policy)
	}
	if policy.AccountFailures != want.AccountFailures || policy.AccountLock != want.AccountLock {
		t.Fatalf("账号失败预算默认值不符: %+v", policy)
	}
	if policy.AccountWindow >= policy.PairLock {
		t.Fatalf("账号窗口必须小于组合锁: window=%s lock=%s", policy.AccountWindow, policy.PairLock)
	}
	if err := (LoginGuardConfig{}).Validate(); err != nil {
		t.Fatalf("零值配置应合法: %v", err)
	}
}

func TestLoginGuardConfigOverrides(t *testing.T) {
	cfg := LoginGuardConfig{
		KeyPrefix:             "yomi:custom:",
		IPMaxAttempts:         30,
		IPWindowSeconds:       600,
		IPMaxFailures:         8,
		IPFailWindowSeconds:   1200,
		IPLockSeconds:         300,
		PairMaxFailures:       4,
		PairWindowSeconds:     900,
		PairLockSeconds:       1200,
		AccountMaxFailures:    12,
		AccountWindowSeconds:  600,
		AccountLockSeconds:    1800,
		AccountMaxLockSeconds: 7200,
		LevelTTLSeconds:       43200,
	}

	policy := cfg.ToPolicy()
	if policy.KeyPrefix != "yomi:custom" {
		t.Fatalf("尾部冒号应被去掉: %s", policy.KeyPrefix)
	}
	if policy.IPAttempts != 30 || policy.IPWindow != 10*time.Minute {
		t.Fatalf("IP 配置覆盖不符: %+v", policy)
	}
	if policy.PairFailures != 4 || policy.AccountFailures != 12 || policy.AccountWindow != 10*time.Minute {
		t.Fatalf("失败预算覆盖不符: %+v", policy)
	}
	if policy.AccountLock != 30*time.Minute || policy.AccountMaxLock != 2*time.Hour {
		t.Fatalf("账号锁覆盖不符: %+v", policy)
	}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("覆盖配置应合法: %v", err)
	}
}

func TestLoginGuardConfigDisabledSkipsValidation(t *testing.T) {
	disabled := false
	cfg := LoginGuardConfig{Enabled: &disabled, IPMaxAttempts: -1}

	if cfg.EnabledOrDefault() {
		t.Fatal("enabled=false 时应关闭")
	}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("关闭时不应校验阈值: %v", err)
	}
}

func TestLoginGuardConfigRejectsInvalidValues(t *testing.T) {
	cases := map[string]LoginGuardConfig{
		"账号窗口不小于组合锁": {AccountWindowSeconds: 600},
		"组合阈值不小于账号":  {PairMaxFailures: 10},
		"账号锁超过上限":    {AccountMaxLockSeconds: 600},
		"键前缀含大写":     {KeyPrefix: "yomi:Admin"},
		"IP 尝试次数为负":  {IPMaxAttempts: -1},
		"窗口超过上限":     {IPWindowSeconds: 86401},
	}

	for name, cfg := range cases {
		t.Run(name, func(t *testing.T) {
			if err := cfg.Validate(); err == nil {
				t.Fatal("期望校验失败")
			}
		})
	}
}

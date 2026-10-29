package config

import (
	"testing"
	"time"
)

func TestRbacCacheConfigDefaults(t *testing.T) {
	var cfg RbacCacheConfig

	if !cfg.EnabledOrDefault() {
		t.Fatal("未配置 enabled 时应默认启用")
	}
	if !cfg.BootstrapWarmupOrDefault() {
		t.Fatal("未配置 bootstrap_warmup 时应默认预热")
	}
	if got := cfg.KeyOrDefault(); got != defaultRbacProjectionKey {
		t.Fatalf("默认投影键不符: %s", got)
	}
	if got := cfg.LockKeyOrDefault(); got != defaultRbacProjectionKey+":lock" {
		t.Fatalf("默认刷新锁键不符: %s", got)
	}
	if got := cfg.MaxBytesOrDefault(); got != defaultRbacProjectionMaxBytes {
		t.Fatalf("默认投影体积上限不符: %d", got)
	}
	if got := cfg.LockWait(); got != 2*time.Second {
		t.Fatalf("默认刷新锁等待不符: %s", got)
	}
	if got := cfg.RefreshInterval(); got != 60*time.Second {
		t.Fatalf("默认定时刷新间隔应为 60s，实际 %s", got)
	}
	if got := cfg.MemorySyncOrDefault(); got != 5*time.Second {
		t.Fatalf("默认内存同步周期应为 5s，实际 %s", got)
	}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("零值配置应合法: %v", err)
	}
}

func TestRbacCacheConfigExplicitValues(t *testing.T) {
	disabled := false
	warmup := false
	zero := 0
	cfg := RbacCacheConfig{
		Enabled:                &disabled,
		BootstrapWarmup:        &warmup,
		RefreshIntervalSeconds: &zero,
		ProjectionMaxBytes:     1024,
		LockWaitSeconds:        5,
	}

	if cfg.EnabledOrDefault() || cfg.BootstrapWarmupOrDefault() {
		t.Fatal("显式 false 应生效")
	}
	if got := cfg.RefreshInterval(); got != 0 {
		t.Fatalf("显式 0 应关闭定时刷新，实际 %s", got)
	}
	if got := cfg.MaxBytesOrDefault(); got != 1024 {
		t.Fatalf("显式体积上限应生效: %d", got)
	}
	if got := cfg.LockWait(); got != 5*time.Second {
		t.Fatalf("显式锁等待应生效: %s", got)
	}
	// 缓存关闭时跳过其余校验
	if err := cfg.Validate(); err != nil {
		t.Fatalf("关闭缓存时配置应合法: %v", err)
	}
}

func TestRbacCacheConfigRejectsInvalidValues(t *testing.T) {
	negative := -1
	tooLarge := maxRbacProjectionRefreshIntervalSeconds + 1

	cases := map[string]RbacCacheConfig{
		"负刷新间隔":   {RefreshIntervalSeconds: &negative},
		"超限刷新间隔":  {RefreshIntervalSeconds: &tooLarge},
		"投影键含空格":  {Key: " yomi:rbac "},
		"锁键含尾随空格": {LockKey: "yomi:rbac:lock "},
		"负体积上限":   {ProjectionMaxBytes: -1},
		"负锁等待":    {LockWaitSeconds: -1},
	}
	for name, cfg := range cases {
		t.Run(name, func(t *testing.T) {
			if err := cfg.Validate(); err == nil {
				t.Fatal("非法配置应校验失败")
			}
		})
	}
}

func TestRbacCacheConfigMemorySync(t *testing.T) {
	// 显式 0 关闭内存层
	zero := 0
	disabled := RbacCacheConfig{MemorySyncSeconds: &zero}
	if got := disabled.MemorySyncOrDefault(); got != 0 {
		t.Fatalf("显式 0 应关闭内存层，实际 %s", got)
	}
	if err := disabled.Validate(); err != nil {
		t.Fatalf("关闭内存层应合法: %v", err)
	}

	// 显式周期生效
	ten := 10
	enabled := RbacCacheConfig{MemorySyncSeconds: &ten}
	if got := enabled.MemorySyncOrDefault(); got != 10*time.Second {
		t.Fatalf("显式内存同步周期应生效: %s", got)
	}

	// 越界拒绝
	negative := -1
	tooLarge := maxRbacMemorySyncSeconds + 1
	for name, cfg := range map[string]RbacCacheConfig{
		"负内存同步周期":  {MemorySyncSeconds: &negative},
		"超限内存同步周期": {MemorySyncSeconds: &tooLarge},
	} {
		t.Run(name, func(t *testing.T) {
			if err := cfg.Validate(); err == nil {
				t.Fatal("非法配置应校验失败")
			}
		})
	}
}

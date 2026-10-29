package config

import (
	"strings"
	"time"

	"gitee.com/hq_project/yomi-boot/component/core/errors"
)

const (
	// 默认投影键与刷新锁键
	defaultRbacProjectionKey     = "yomi:admin:rbac:projection"
	defaultRbacProjectionLockKey = "yomi:admin:rbac:projection:lock"
	// defaultRbacProjectionMaxBytes 是投影载荷默认体积上限（8 MiB）
	defaultRbacProjectionMaxBytes = 8 << 20
	// defaultRbacProjectionLockWaitSeconds 是刷新锁默认最长等待（秒）
	defaultRbacProjectionLockWaitSeconds = 2
	// defaultRbacProjectionRefreshIntervalSeconds 是定时刷新默认间隔（秒）
	defaultRbacProjectionRefreshIntervalSeconds = 60
	// maxRbacProjectionRefreshIntervalSeconds 是定时刷新间隔上限（秒）
	maxRbacProjectionRefreshIntervalSeconds = 86400
	// defaultRbacMemorySyncSeconds 是内存快照同步周期默认值（秒）
	defaultRbacMemorySyncSeconds = 5
	// maxRbacMemorySyncSeconds 是内存快照同步周期上限（秒）
	maxRbacMemorySyncSeconds = 600
)

// RbacConfig 是 biz.rbac 配置
type RbacConfig struct {
	Cache RbacCacheConfig `mapstructure:"cache"`
}

// RbacCacheConfig 是鉴权投影缓存配置；零值字段按默认值处理
type RbacCacheConfig struct {
	Enabled                *bool  `mapstructure:"enabled"`                  // Enabled 是是否启用鉴权缓存，未配置默认 true
	Key                    string `mapstructure:"key"`                      // Key 是投影键，空则 yomi:admin:rbac:projection
	LockKey                string `mapstructure:"lock_key"`                 // LockKey 是刷新锁键，空则投影键加 :lock
	BootstrapWarmup        *bool  `mapstructure:"bootstrap_warmup"`         // BootstrapWarmup 是启动时投影不可用才构建，未配置默认 true
	RefreshIntervalSeconds *int   `mapstructure:"refresh_interval_seconds"` // RefreshIntervalSeconds 是定时刷新间隔（秒），未配置默认 60，0 表示关闭
	ProjectionMaxBytes     int    `mapstructure:"projection_max_bytes"`     // ProjectionMaxBytes 是投影体积上限，0 则 8 MiB
	LockWaitSeconds        int    `mapstructure:"lock_wait_seconds"`        // LockWaitSeconds 是刷新锁最长等待（秒），0 则 2
	MemorySyncSeconds      *int   `mapstructure:"memory_sync_seconds"`      // MemorySyncSeconds 是内存快照同步周期（秒），未配置默认 5，0 表示关闭内存层
}

// EnabledOrDefault 返回是否启用鉴权缓存，未配置默认 true
func (c RbacCacheConfig) EnabledOrDefault() bool {
	return c.Enabled == nil || *c.Enabled
}

// BootstrapWarmupOrDefault 返回启动时是否预热投影，未配置默认 true
func (c RbacCacheConfig) BootstrapWarmupOrDefault() bool {
	return c.BootstrapWarmup == nil || *c.BootstrapWarmup
}

// KeyOrDefault 返回投影键
func (c RbacCacheConfig) KeyOrDefault() string {
	if key := strings.TrimSpace(c.Key); key != "" {
		return key
	}
	return defaultRbacProjectionKey
}

// LockKeyOrDefault 返回刷新锁键，空则取投影键加 :lock
func (c RbacCacheConfig) LockKeyOrDefault() string {
	if key := strings.TrimSpace(c.LockKey); key != "" {
		return key
	}
	return c.KeyOrDefault() + ":lock"
}

// MaxBytesOrDefault 返回投影体积上限
func (c RbacCacheConfig) MaxBytesOrDefault() int {
	if c.ProjectionMaxBytes > 0 {
		return c.ProjectionMaxBytes
	}
	return defaultRbacProjectionMaxBytes
}

// LockWait 返回刷新锁最长等待
func (c RbacCacheConfig) LockWait() time.Duration {
	if c.LockWaitSeconds > 0 {
		return time.Duration(c.LockWaitSeconds) * time.Second
	}
	return defaultRbacProjectionLockWaitSeconds * time.Second
}

// RefreshInterval 返回定时刷新间隔，未配置默认 60s；显式 0 表示关闭
func (c RbacCacheConfig) RefreshInterval() time.Duration {
	if c.RefreshIntervalSeconds == nil {
		return defaultRbacProjectionRefreshIntervalSeconds * time.Second
	}
	return time.Duration(*c.RefreshIntervalSeconds) * time.Second
}

// MemorySyncOrDefault 返回内存快照同步周期，未配置默认 5s；显式 0 表示关闭内存层
func (c RbacCacheConfig) MemorySyncOrDefault() time.Duration {
	if c.MemorySyncSeconds == nil {
		return defaultRbacMemorySyncSeconds * time.Second
	}
	return time.Duration(*c.MemorySyncSeconds) * time.Second
}

// Validate 校验鉴权缓存配置语义
func (c RbacCacheConfig) Validate() error {
	if !c.EnabledOrDefault() {
		return nil
	}
	if key := c.Key; key != strings.TrimSpace(key) {
		return errors.Newf("%s.rbac.cache: key 不能含首尾空格", SectionKey)
	}
	if lockKey := c.LockKey; lockKey != strings.TrimSpace(lockKey) {
		return errors.Newf("%s.rbac.cache: lock_key 不能含首尾空格", SectionKey)
	}
	if c.ProjectionMaxBytes < 0 {
		return errors.Newf("%s.rbac.cache: projection_max_bytes 不能为负数", SectionKey)
	}
	if c.LockWaitSeconds < 0 {
		return errors.Newf("%s.rbac.cache: lock_wait_seconds 不能为负数", SectionKey)
	}
	if c.RefreshIntervalSeconds != nil {
		seconds := *c.RefreshIntervalSeconds
		if seconds < 0 || seconds > maxRbacProjectionRefreshIntervalSeconds {
			return errors.Newf("%s.rbac.cache: refresh_interval_seconds 必须在 0..%d 之间",
				SectionKey, maxRbacProjectionRefreshIntervalSeconds)
		}
	}
	if c.MemorySyncSeconds != nil {
		seconds := *c.MemorySyncSeconds
		if seconds < 0 || seconds > maxRbacMemorySyncSeconds {
			return errors.Newf("%s.rbac.cache: memory_sync_seconds 必须在 0..%d 之间",
				SectionKey, maxRbacMemorySyncSeconds)
		}
	}
	return nil
}

// Validate 校验 RBAC 配置语义
func (c RbacConfig) Validate() error {
	return c.Cache.Validate()
}

package config

import (
	"strings"
	"time"

	"yomi-admin-backend/apps/server/internal/support/loginguard"

	"gitee.com/hq_project/yomi-boot/component/core/errors"
)

// LoginGuardConfig 是 biz.login_guard 配置；零值字段按默认值处理
type LoginGuardConfig struct {
	Enabled               *bool  `mapstructure:"enabled"`                  // Enabled 是是否启用登录失败锁定，未配置默认 true
	KeyPrefix             string `mapstructure:"key_prefix"`               // KeyPrefix 是登录保护状态 Redis 键前缀，空则 yomi:admin:loginguard
	IPMaxAttempts         int64  `mapstructure:"ip_max_attempts"`          // IPMaxAttempts 是单 IP 窗口内允许的登录尝试次数，0 则 20
	IPWindowSeconds       int    `mapstructure:"ip_window_seconds"`        // IPWindowSeconds 是尝试次数窗口（秒），0 则 300
	IPMaxFailures         int64  `mapstructure:"ip_max_failures"`          // IPMaxFailures 是单 IP 跨账号失败次数上限，0 则 10
	IPFailWindowSeconds   int    `mapstructure:"ip_fail_window_seconds"`   // IPFailWindowSeconds 是 IP 失败计数窗口（秒），0 则 900
	IPLockSeconds         int    `mapstructure:"ip_lock_seconds"`          // IPLockSeconds 是 IP 锁定时长（秒），0 则 600
	PairMaxFailures       int64  `mapstructure:"pair_max_failures"`        // PairMaxFailures 是 IP+账号失败次数上限，0 则 5
	PairWindowSeconds     int    `mapstructure:"pair_window_seconds"`      // PairWindowSeconds 是 IP+账号失败计数窗口（秒），0 则 600
	PairLockSeconds       int    `mapstructure:"pair_lock_seconds"`        // PairLockSeconds 是 IP+账号锁定时长（秒），0 则 600
	AccountMaxFailures    int64  `mapstructure:"account_max_failures"`     // AccountMaxFailures 是账号跨 IP 失败次数上限，0 则 10
	AccountWindowSeconds  int    `mapstructure:"account_window_seconds"`   // AccountWindowSeconds 是账号失败计数窗口（秒），0 则 300
	AccountLockSeconds    int    `mapstructure:"account_lock_seconds"`     // AccountLockSeconds 是账号基础锁定时长（秒），0 则 900
	AccountMaxLockSeconds int    `mapstructure:"account_max_lock_seconds"` // AccountMaxLockSeconds 是账号锁定时长上限（秒），0 则 3600
	LevelTTLSeconds       int    `mapstructure:"level_ttl_seconds"`        // LevelTTLSeconds 是账号锁定升级等级的记忆时长（秒），0 则 86400
}

// EnabledOrDefault 返回是否启用登录失败锁定，未配置默认 true
func (c LoginGuardConfig) EnabledOrDefault() bool {
	return c.Enabled == nil || *c.Enabled
}

// ToPolicy 转换为登录保护运行时策略；零值字段使用默认值。
//
// 默认值与语义校验以 support/loginguard 为准，避免两处维护。
func (c LoginGuardConfig) ToPolicy() loginguard.Policy {
	policy := loginguard.DefaultPolicy()
	policy.Enabled = c.EnabledOrDefault()
	if prefix := strings.TrimSpace(c.KeyPrefix); prefix != "" {
		policy.KeyPrefix = strings.TrimRight(prefix, ":")
	}
	if c.IPMaxAttempts != 0 {
		policy.IPAttempts = c.IPMaxAttempts
	}
	if c.IPWindowSeconds != 0 {
		policy.IPWindow = seconds(c.IPWindowSeconds)
	}
	if c.IPMaxFailures != 0 {
		policy.IPFailures = c.IPMaxFailures
	}
	if c.IPFailWindowSeconds != 0 {
		policy.IPFailWindow = seconds(c.IPFailWindowSeconds)
	}
	if c.IPLockSeconds != 0 {
		policy.IPLock = seconds(c.IPLockSeconds)
	}
	if c.PairMaxFailures != 0 {
		policy.PairFailures = c.PairMaxFailures
	}
	if c.PairWindowSeconds != 0 {
		policy.PairWindow = seconds(c.PairWindowSeconds)
	}
	if c.PairLockSeconds != 0 {
		policy.PairLock = seconds(c.PairLockSeconds)
	}
	if c.AccountMaxFailures != 0 {
		policy.AccountFailures = c.AccountMaxFailures
	}
	if c.AccountWindowSeconds != 0 {
		policy.AccountWindow = seconds(c.AccountWindowSeconds)
	}
	if c.AccountLockSeconds != 0 {
		policy.AccountLock = seconds(c.AccountLockSeconds)
	}
	if c.AccountMaxLockSeconds != 0 {
		policy.AccountMaxLock = seconds(c.AccountMaxLockSeconds)
	}
	if c.LevelTTLSeconds != 0 {
		policy.LevelTTL = seconds(c.LevelTTLSeconds)
	}
	return policy
}

// Validate 校验登录保护配置语义
func (c LoginGuardConfig) Validate() error {
	if err := c.ToPolicy().Validate(); err != nil {
		return errors.Wrap(err, "%s.login_guard", SectionKey)
	}
	return nil
}

func seconds(value int) time.Duration {
	return time.Duration(value) * time.Second
}

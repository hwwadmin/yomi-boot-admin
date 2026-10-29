// Package loginguard 提供登录失败预算与临时锁定策略。
//
// 状态只存 Redis：所有 key 都带 TTL，多实例共享，进程重启不保留。
package loginguard

import (
	"strings"
	"time"

	"gitee.com/hq_project/yomi-boot/component/core/errors"
)

// DefaultKeyPrefix 是登录保护状态默认 Redis 键前缀
const DefaultKeyPrefix = "yomi:admin:loginguard"

const (
	maxKeyPrefixLen = 64
	maxWindow       = 24 * time.Hour
	maxLock         = 24 * time.Hour
	maxLevelTTL     = 7 * 24 * time.Hour
)

// Policy 描述登录失败预算与锁定时长。
//
// 计数维度分四类：IP 尝试配额、IP 失败预算、IP+账号失败预算、账号失败预算。
// AccountWindow 必须小于 PairLock，保证单个 IP 在组合锁定期间无法凑满账号阈值，
// 避免攻击者用锁定机制反过来锁死管理员。
type Policy struct {
	Enabled   bool
	KeyPrefix string

	IPAttempts   int64
	IPWindow     time.Duration
	IPFailures   int64
	IPFailWindow time.Duration
	IPLock       time.Duration

	PairFailures int64
	PairWindow   time.Duration
	PairLock     time.Duration

	AccountFailures int64
	AccountWindow   time.Duration
	AccountLock     time.Duration
	AccountMaxLock  time.Duration
	LevelTTL        time.Duration
}

// DefaultPolicy 返回默认策略
func DefaultPolicy() Policy {
	return Policy{
		Enabled:         true,
		KeyPrefix:       DefaultKeyPrefix,
		IPAttempts:      20,
		IPWindow:        5 * time.Minute,
		IPFailures:      10,
		IPFailWindow:    15 * time.Minute,
		IPLock:          10 * time.Minute,
		PairFailures:    5,
		PairWindow:      10 * time.Minute,
		PairLock:        10 * time.Minute,
		AccountFailures: 10,
		AccountWindow:   5 * time.Minute,
		AccountLock:     15 * time.Minute,
		AccountMaxLock:  60 * time.Minute,
		LevelTTL:        24 * time.Hour,
	}
}

// Validate 校验策略语义
func (p Policy) Validate() error {
	if !p.Enabled {
		return nil
	}
	if err := validateKeyPrefix(p.normalizedKeyPrefix()); err != nil {
		return err
	}
	if p.IPAttempts <= 0 {
		return errors.New("loginguard: ip_max_attempts 必须大于 0")
	}
	if p.IPFailures <= 0 {
		return errors.New("loginguard: ip_max_failures 必须大于 0")
	}
	if p.PairFailures <= 0 {
		return errors.New("loginguard: pair_max_failures 必须大于 0")
	}
	if p.AccountFailures <= 0 {
		return errors.New("loginguard: account_max_failures 必须大于 0")
	}
	if p.PairFailures >= p.AccountFailures {
		return errors.New("loginguard: pair_max_failures 必须小于 account_max_failures，否则单个 IP 可以触发账号锁")
	}

	durations := []struct {
		name  string
		value time.Duration
		max   time.Duration
	}{
		{"ip_window", p.IPWindow, maxWindow},
		{"ip_fail_window", p.IPFailWindow, maxWindow},
		{"ip_lock", p.IPLock, maxLock},
		{"pair_window", p.PairWindow, maxWindow},
		{"pair_lock", p.PairLock, maxLock},
		{"account_window", p.AccountWindow, maxWindow},
		{"account_lock", p.AccountLock, maxLock},
		{"account_max_lock", p.AccountMaxLock, maxLock},
		{"level_ttl", p.LevelTTL, maxLevelTTL},
	}
	for _, item := range durations {
		if item.value <= 0 {
			return errors.Newf("loginguard: %s 必须大于 0", item.name)
		}
		if item.value > item.max {
			return errors.Newf("loginguard: %s 不能超过 %s", item.name, item.max)
		}
	}
	if p.AccountLock > p.AccountMaxLock {
		return errors.New("loginguard: account_lock_seconds 不能大于 account_max_lock_seconds")
	}
	if p.AccountWindow >= p.PairLock {
		return errors.New("loginguard: account_window_seconds 必须小于 pair_lock_seconds，否则单个 IP 可以触发账号锁")
	}
	return nil
}

// normalizedKeyPrefix 返回去掉首尾空白与尾部冒号的键前缀
func (p Policy) normalizedKeyPrefix() string {
	return strings.TrimRight(strings.TrimSpace(p.KeyPrefix), ":")
}

func validateKeyPrefix(prefix string) error {
	if prefix == "" {
		return errors.New("loginguard: key_prefix 不能为空")
	}
	if len(prefix) > maxKeyPrefixLen {
		return errors.Newf("loginguard: key_prefix 长度不能超过 %d", maxKeyPrefixLen)
	}
	for _, r := range prefix {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == ':' || r == '_' {
			continue
		}
		return errors.New("loginguard: key_prefix 仅允许小写字母、数字、冒号和下划线")
	}
	return nil
}

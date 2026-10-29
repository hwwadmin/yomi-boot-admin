package loginguard

import (
	"context"
	"strings"
	"time"
	"unicode/utf8"

	"gitee.com/hq_project/yomi-boot/component/common/redisratelimit"
	"gitee.com/hq_project/yomi-boot/component/core/errors"
	"gitee.com/hq_project/yomi-boot/component/core/logger"
)

const (
	// attemptKeyPrefix 是 IP 尝试配额的逻辑 key 前缀，由 redisratelimit 统一加配置前缀
	attemptKeyPrefix = "loginguard:ip:"
	// maxLogRunes 限制日志中的账号长度，避免超长输入放大日志
	maxLogRunes = 64
)

// ScriptExecutor 是登录保护需要的 Redis 原子脚本能力
type ScriptExecutor interface {
	// OpsScript4Eval 执行 Lua 脚本，keys 与 args 分别映射为 KEYS 与 ARGV
	OpsScript4Eval(ctx context.Context, script string, keys []string, args ...any) (any, error)
	// OpsKey4Delete 删除一个或多个 key
	OpsKey4Delete(ctx context.Context, keys ...string) error
}

// Limiter 是 IP 尝试配额需要的限流能力
type Limiter interface {
	// Allow 尝试消耗 1 次配额；超限返回 redisratelimit.ErrRateLimited
	Allow(ctx context.Context, key string, rule redisratelimit.Rule) error
}

// Guard 是登录失败预算与临时锁定策略
type Guard struct {
	redis   ScriptExecutor
	limiter Limiter
	policy  Policy
	prefix  string
}

// New 创建登录保护守卫；启用状态下 redis 与 limiter 不能为 nil
func New(redis ScriptExecutor, limiter Limiter, policy Policy) *Guard {
	if err := policy.Validate(); err != nil {
		panic("loginguard: " + err.Error())
	}
	if policy.Enabled && (redis == nil || limiter == nil) {
		panic("loginguard: 启用时必须提供 redis 与 limiter")
	}
	return &Guard{redis: redis, limiter: limiter, policy: policy, prefix: policy.normalizedKeyPrefix()}
}

// BeforeAttempt 在验证码与验密之前执行：先消耗 IP 尝试配额，再检查三层锁定。
//
// 校验码错误也会消耗尝试配额；锁定检查放在验证码之前，被锁用户不必再解验证码。
func (g *Guard) BeforeAttempt(ctx context.Context, account, clientIP string) (Decision, error) {
	if !g.policy.Enabled {
		return Decision{Outcome: OutcomeAllowed}, nil
	}

	ip := ipID(clientIP)
	rule := redisratelimit.Rule{Limit: g.policy.IPAttempts, Window: g.policy.IPWindow}
	if err := g.limiter.Allow(ctx, attemptKeyPrefix+ip, rule); err != nil {
		if errors.Is(err, redisratelimit.ErrRateLimited) {
			return Decision{Outcome: OutcomeRateLimited, RetryAfter: g.policy.IPWindow}, nil
		}
		return Decision{}, err
	}

	keys := []string{g.ipKey(ip, "lock")}
	if id := accountID(account); id != "" {
		keys = append(keys, g.pairKey(id, ip, "lock"), g.accountKey(id, "lock"))
	}
	ttlMs, err := g.evalInt64(ctx, checkLocksScript, keys)
	if err != nil {
		return Decision{}, err
	}
	if ttlMs > 0 {
		return Decision{Outcome: OutcomeLocked, RetryAfter: time.Duration(ttlMs) * time.Millisecond}, nil
	}
	return Decision{Outcome: OutcomeAllowed}, nil
}

// RecordFailure 记录一次凭证失败，返回本次是否触发锁定。
//
// 三个维度都会记账，命中任意一个维度即视为锁定；账号为空时只记 IP 维度。
func (g *Guard) RecordFailure(ctx context.Context, account, clientIP string) (Decision, error) {
	if !g.policy.Enabled {
		return Decision{Outcome: OutcomeAllowed}, nil
	}

	ip := ipID(clientIP)
	id := accountID(account)

	ipLock, err := g.recordFailure(ctx, g.ipKey(ip, "fail"), g.ipKey(ip, "lock"),
		g.policy.IPFailWindow, g.policy.IPFailures, g.policy.IPLock)
	if err != nil {
		return Decision{}, err
	}

	var pairLock, accountLock time.Duration
	if id != "" {
		pairLock, err = g.recordFailure(ctx, g.pairKey(id, ip, "fail"), g.pairKey(id, ip, "lock"),
			g.policy.PairWindow, g.policy.PairFailures, g.policy.PairLock)
		if err != nil {
			return Decision{}, err
		}
		accountLock, err = g.recordAccountFailure(ctx, id)
		if err != nil {
			return Decision{}, err
		}
	}

	decision := Decision{Outcome: OutcomeAllowed}
	applyLock := func(dimension string, lock time.Duration) {
		if lock <= 0 {
			return
		}
		g.logLock(ctx, dimension, account, clientIP, lock)
		if lock > decision.RetryAfter {
			decision = Decision{Outcome: OutcomeLocked, RetryAfter: lock}
		}
	}
	applyLock("ip", ipLock)
	applyLock("pair", pairLock)
	applyLock("account", accountLock)
	return decision, nil
}

// RecordSuccess 清空账号维度与 IP+账号维度的状态。
//
// 不清 IP 维度：IP 失败计数跨账号，成功登录不能把 IP 的喷洒记录洗白。
func (g *Guard) RecordSuccess(ctx context.Context, account, clientIP string) error {
	if !g.policy.Enabled {
		return nil
	}
	id := accountID(account)
	if id == "" {
		return nil
	}
	ip := ipID(clientIP)
	return g.redis.OpsKey4Delete(ctx,
		g.pairKey(id, ip, "fail"), g.pairKey(id, ip, "lock"),
		g.accountKey(id, "fail"), g.accountKey(id, "lock"), g.accountKey(id, "level"),
	)
}

// recordFailure 累加一个固定锁时长的失败维度
func (g *Guard) recordFailure(ctx context.Context, failKey, lockKey string, window time.Duration, threshold int64, lock time.Duration) (time.Duration, error) {
	lockMs, err := g.evalInt64(ctx, recordFailureScript, []string{failKey, lockKey},
		window.Milliseconds(), threshold, lock.Milliseconds())
	if err != nil {
		return 0, err
	}
	return time.Duration(lockMs) * time.Millisecond, nil
}

// recordAccountFailure 累加账号维度失败次数，并按等级升级锁定时长
func (g *Guard) recordAccountFailure(ctx context.Context, account string) (time.Duration, error) {
	lockMs, err := g.evalInt64(ctx, recordAccountFailureScript,
		[]string{g.accountKey(account, "fail"), g.accountKey(account, "lock"), g.accountKey(account, "level")},
		g.policy.AccountWindow.Milliseconds(), g.policy.AccountFailures,
		g.policy.AccountLock.Milliseconds(), g.policy.AccountMaxLock.Milliseconds(),
		g.policy.LevelTTL.Milliseconds())
	if err != nil {
		return 0, err
	}
	return time.Duration(lockMs) * time.Millisecond, nil
}

// evalInt64 执行返回整数的 Lua 脚本
func (g *Guard) evalInt64(ctx context.Context, script string, keys []string, args ...any) (int64, error) {
	result, err := g.redis.OpsScript4Eval(ctx, script, keys, args...)
	if err != nil {
		return 0, err
	}
	switch value := result.(type) {
	case int64:
		return value, nil
	case int:
		return int64(value), nil
	default:
		return 0, errors.Newf("loginguard: 脚本返回类型异常 %T", result)
	}
}

// logLock 记录锁定事件，账号字段做长度截断
func (g *Guard) logLock(ctx context.Context, dimension, account, clientIP string, lock time.Duration) {
	logger.Warn(ctx, "登录失败触发临时锁定",
		"event", "auth.login.locked",
		"dimension", dimension,
		"account", truncate(account, maxLogRunes),
		"ip", clientIP,
		"lock_seconds", int64(lock/time.Second),
	)
}

func truncate(value string, maxRunes int) string {
	value = strings.TrimSpace(value)
	if utf8.RuneCountInString(value) <= maxRunes {
		return value
	}
	return string([]rune(value)[:maxRunes])
}

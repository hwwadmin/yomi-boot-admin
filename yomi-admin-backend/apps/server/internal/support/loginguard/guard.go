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

// dimensionLocks 是一次失败记账后各维度的锁时长；0 表示该维度未触发锁定
type dimensionLocks struct {
	ip      time.Duration
	pair    time.Duration
	account time.Duration
}

// RecordFailure 记录一次凭证失败，返回本次是否触发锁定。
//
// 涉及的维度在单次 Redis 往返内原子记账：三个维度都会记账，命中任意一个维度即视为锁定；
// 账号为空时只记 IP 维度。不存在"部分维度记上、部分丢失"的中间态。
func (g *Guard) RecordFailure(ctx context.Context, account, clientIP string) (Decision, error) {
	if !g.policy.Enabled {
		return Decision{Outcome: OutcomeAllowed}, nil
	}

	locks, err := g.recordFailure(ctx, account, clientIP)
	if err != nil {
		return Decision{}, err
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
	applyLock("ip", locks.ip)
	applyLock("pair", locks.pair)
	applyLock("account", locks.account)
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

// recordFailure 在单次 Redis 往返内累加一次凭证失败涉及的全部维度。
//
// 账号为空只记 IP 维度；否则 IP、IP+账号、账号三个维度在同一个 Lua 脚本内原子生效，
// 脚本失败时三个维度都不会落账，由调用方按"保护不可用"统一处理。
func (g *Guard) recordFailure(ctx context.Context, account, clientIP string) (dimensionLocks, error) {
	ip := ipID(clientIP)
	id := accountID(account)

	var keys []string
	var args []any
	if id == "" {
		keys = []string{g.ipKey(ip, "fail"), g.ipKey(ip, "lock")}
		args = []any{
			g.policy.IPFailWindow.Milliseconds(), g.policy.IPFailures, g.policy.IPLock.Milliseconds(),
		}
	} else {
		keys = []string{
			g.ipKey(ip, "fail"), g.ipKey(ip, "lock"),
			g.pairKey(id, ip, "fail"), g.pairKey(id, ip, "lock"),
			g.accountKey(id, "fail"), g.accountKey(id, "lock"), g.accountKey(id, "level"),
		}
		args = []any{
			g.policy.IPFailWindow.Milliseconds(), g.policy.IPFailures, g.policy.IPLock.Milliseconds(),
			g.policy.PairWindow.Milliseconds(), g.policy.PairFailures, g.policy.PairLock.Milliseconds(),
			g.policy.AccountWindow.Milliseconds(), g.policy.AccountFailures, g.policy.AccountLock.Milliseconds(),
			g.policy.AccountMaxLock.Milliseconds(), g.policy.LevelTTL.Milliseconds(),
		}
	}

	locks, err := g.evalLocks(ctx, recordFailureScript, keys, args...)
	if err != nil {
		return dimensionLocks{}, err
	}
	if len(locks) == 0 {
		return dimensionLocks{}, errors.New("loginguard: 失败记账脚本未返回结果")
	}
	if len(locks) >= 3 {
		return dimensionLocks{ip: locks[0], pair: locks[1], account: locks[2]}, nil
	}
	return dimensionLocks{ip: locks[0]}, nil
}

// evalInt64 执行返回整数的 Lua 脚本
func (g *Guard) evalInt64(ctx context.Context, script string, keys []string, args ...any) (int64, error) {
	result, err := g.redis.OpsScript4Eval(ctx, script, keys, args...)
	if err != nil {
		return 0, err
	}
	return scriptInt64(result)
}

// evalLocks 执行返回"各维度锁时长（毫秒）数组"的 Lua 脚本
func (g *Guard) evalLocks(ctx context.Context, script string, keys []string, args ...any) ([]time.Duration, error) {
	result, err := g.redis.OpsScript4Eval(ctx, script, keys, args...)
	if err != nil {
		return nil, err
	}
	values, ok := result.([]any)
	if !ok {
		return nil, errors.Newf("loginguard: 脚本返回类型异常 %T", result)
	}

	locks := make([]time.Duration, 0, len(values))
	for _, value := range values {
		ms, err := scriptInt64(value)
		if err != nil {
			return nil, err
		}
		locks = append(locks, time.Duration(ms)*time.Millisecond)
	}
	return locks, nil
}

// scriptInt64 把 Redis 脚本返回的整数值转换为 int64
func scriptInt64(value any) (int64, error) {
	switch v := value.(type) {
	case int64:
		return v, nil
	case int:
		return int64(v), nil
	default:
		return 0, errors.Newf("loginguard: 脚本返回类型异常 %T", value)
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

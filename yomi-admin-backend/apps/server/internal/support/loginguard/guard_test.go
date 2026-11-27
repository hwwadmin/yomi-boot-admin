package loginguard

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/hwwadmin/yomi-boot/component/common/redisratelimit"
	"github.com/hwwadmin/yomi-boot/component/core/errors"
)

// scriptCall 记录一次 Lua 脚本调用的入参
type scriptCall struct {
	script string
	keys   []string
	args   []any
}

// fakeScriptExecutor 是 Redis 脚本执行测试桩
type fakeScriptExecutor struct {
	results []any
	err     error
	calls   []scriptCall
}

func (f *fakeScriptExecutor) OpsScript4Eval(_ context.Context, script string, keys []string, args ...any) (any, error) {
	f.calls = append(f.calls, scriptCall{
		script: script,
		keys:   append([]string(nil), keys...),
		args:   append([]any(nil), args...),
	})
	if f.err != nil {
		return nil, f.err
	}
	if len(f.results) == 0 {
		return int64(0), nil
	}
	result := f.results[0]
	f.results = f.results[1:]
	return result, nil
}

// fakeLimiter 是 IP 尝试配额测试桩
type fakeLimiter struct {
	err   error
	calls int
	rule  redisratelimit.Rule
}

func (f *fakeLimiter) Allow(_ context.Context, _ string, rule redisratelimit.Rule) error {
	f.calls++
	f.rule = rule
	return f.err
}

func TestDefaultPolicyValid(t *testing.T) {
	if err := DefaultPolicy().Validate(); err != nil {
		t.Fatalf("默认策略应合法: %v", err)
	}
}

func TestPolicyValidation(t *testing.T) {
	cases := map[string]func(*Policy){
		"pair 阈值不能等于账号阈值": func(p *Policy) { p.PairFailures = p.AccountFailures },
		"账号窗口不能等于组合锁":     func(p *Policy) { p.AccountWindow = p.PairLock },
		"账号锁不能超过上限":       func(p *Policy) { p.AccountLock = p.AccountMaxLock + time.Second },
		"键前缀不能含大写":        func(p *Policy) { p.KeyPrefix = "yomi:Admin" },
		"IP 尝试次数必须为正":     func(p *Policy) { p.IPAttempts = 0 },
		"等级 TTL 必须为正":     func(p *Policy) { p.LevelTTL = 0 },
	}

	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			policy := DefaultPolicy()
			mutate(&policy)
			if err := policy.Validate(); err == nil {
				t.Fatal("期望校验失败")
			}
		})
	}
}

func TestPolicyDisabledSkipsValidation(t *testing.T) {
	policy := Policy{Enabled: false}
	if err := policy.Validate(); err != nil {
		t.Fatalf("关闭时不应校验: %v", err)
	}

	guard := New(nil, nil, policy)
	decision, err := guard.BeforeAttempt(context.Background(), "admin", "10.0.0.1")
	if err != nil || decision.Outcome != OutcomeAllowed {
		t.Fatalf("关闭时应直接放行: decision=%+v err=%v", decision, err)
	}
	if err := guard.ReleaseAttempt(context.Background(), "admin", "10.0.0.1"); err != nil {
		t.Fatalf("关闭时释放预占不应报错: %v", err)
	}
	decision, err = guard.ReserveAttempt(context.Background(), "admin", "10.0.0.1")
	if err != nil || decision.Outcome != OutcomeAllowed {
		t.Fatalf("关闭时预占应直接放行: decision=%+v err=%v", decision, err)
	}
}

func TestNewPanicsOnInvalidPolicy(t *testing.T) {
	policy := DefaultPolicy()
	policy.PairFailures = policy.AccountFailures

	defer func() {
		if recover() == nil {
			t.Fatal("非法策略应 panic")
		}
	}()
	New(&fakeScriptExecutor{}, &fakeLimiter{}, policy)
}

func TestNewPanicsOnMissingDependencies(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("启用状态缺少依赖应 panic")
		}
	}()
	New(nil, nil, DefaultPolicy())
}

func TestBeforeAttemptAllowed(t *testing.T) {
	redis := &fakeScriptExecutor{}
	limiter := &fakeLimiter{}
	guard := New(redis, limiter, DefaultPolicy())

	decision, err := guard.BeforeAttempt(context.Background(), "admin", "10.0.0.1")
	if err != nil {
		t.Fatalf("前置校验失败: %v", err)
	}
	if decision.Outcome != OutcomeAllowed {
		t.Fatalf("期望放行: %+v", decision)
	}
	if limiter.calls != 1 || limiter.rule.Limit != DefaultPolicy().IPAttempts || limiter.rule.Window != DefaultPolicy().IPWindow {
		t.Fatalf("尝试配额入参不符: %+v", limiter.rule)
	}
	if len(redis.calls) != 1 || len(redis.calls[0].keys) != 3 {
		t.Fatalf("应检查 IP、组合、账号三个锁定 key: %+v", redis.calls)
	}
	for _, key := range redis.calls[0].keys {
		if !strings.HasSuffix(key, ":lock") {
			t.Fatalf("锁定 key 后缀不符: %s", key)
		}
	}
}

func TestBeforeAttemptWithoutAccountChecksIPOnly(t *testing.T) {
	redis := &fakeScriptExecutor{}
	guard := New(redis, &fakeLimiter{}, DefaultPolicy())

	if _, err := guard.BeforeAttempt(context.Background(), "  ", "10.0.0.1"); err != nil {
		t.Fatalf("前置校验失败: %v", err)
	}
	if len(redis.calls) != 1 || len(redis.calls[0].keys) != 1 {
		t.Fatalf("空账号只应检查 IP 锁定: %+v", redis.calls)
	}
}

func TestBeforeAttemptLocked(t *testing.T) {
	redis := &fakeScriptExecutor{results: []any{int64(1500)}}
	guard := New(redis, &fakeLimiter{}, DefaultPolicy())

	decision, err := guard.BeforeAttempt(context.Background(), "admin", "10.0.0.1")
	if err != nil {
		t.Fatalf("前置校验失败: %v", err)
	}
	if decision.Outcome != OutcomeLocked || decision.RetryAfter != 1500*time.Millisecond {
		t.Fatalf("期望锁定 1.5s: %+v", decision)
	}
}

func TestBeforeAttemptRateLimited(t *testing.T) {
	limiter := &fakeLimiter{err: errors.Sentinel(redisratelimit.ErrRateLimited)}
	guard := New(&fakeScriptExecutor{}, limiter, DefaultPolicy())

	decision, err := guard.BeforeAttempt(context.Background(), "admin", "10.0.0.1")
	if err != nil {
		t.Fatalf("限流不应作为错误返回: %v", err)
	}
	if decision.Outcome != OutcomeRateLimited {
		t.Fatalf("期望限流: %+v", decision)
	}
}

func TestBeforeAttemptLimiterFailure(t *testing.T) {
	limiter := &fakeLimiter{err: errors.New("redis down")}
	guard := New(&fakeScriptExecutor{}, limiter, DefaultPolicy())

	if _, err := guard.BeforeAttempt(context.Background(), "admin", "10.0.0.1"); err == nil {
		t.Fatal("限流器故障应向上传播")
	}
}

func TestReserveAttemptAllowed(t *testing.T) {
	redis := &fakeScriptExecutor{results: []any{[]any{int64(0), int64(0)}}}
	guard := New(redis, &fakeLimiter{}, DefaultPolicy())

	decision, err := guard.ReserveAttempt(context.Background(), "admin", "10.0.0.1")
	if err != nil {
		t.Fatalf("预占失败: %v", err)
	}
	if decision.Outcome != OutcomeAllowed {
		t.Fatalf("期望放行: %+v", decision)
	}
	if len(redis.calls) != 1 {
		t.Fatalf("预占应一次脚本调用完成: %d", len(redis.calls))
	}

	call := redis.calls[0]
	if call.script != reserveAttemptScript {
		t.Fatal("应使用预占脚本")
	}

	ip := ipID("10.0.0.1")
	id := accountID("admin")
	wantKeys := []string{
		guard.ipKey(ip, "lock"), guard.ipKey(ip, "fail"), guard.ipKey(ip, "pending"),
		guard.pairKey(id, ip, "lock"), guard.pairKey(id, ip, "fail"), guard.pairKey(id, ip, "pending"),
		guard.accountKey(id, "lock"), guard.accountKey(id, "fail"), guard.accountKey(id, "pending"),
	}
	assertKeys(t, call.keys, wantKeys)

	policy := DefaultPolicy()
	wantArgs := []any{
		attemptReservationTTL.Milliseconds(), policy.IPFailures, policy.PairFailures, policy.AccountFailures,
	}
	assertArgs(t, call.args, wantArgs)
}

func TestReserveAttemptWithoutAccountUsesIPOnly(t *testing.T) {
	redis := &fakeScriptExecutor{results: []any{[]any{int64(0), int64(0)}}}
	guard := New(redis, &fakeLimiter{}, DefaultPolicy())

	if _, err := guard.ReserveAttempt(context.Background(), " ", "10.0.0.1"); err != nil {
		t.Fatalf("预占失败: %v", err)
	}
	if len(redis.calls[0].keys) != 3 {
		t.Fatalf("空账号只应预占 IP 维度: %+v", redis.calls[0].keys)
	}
	assertArgs(t, redis.calls[0].args, []any{attemptReservationTTL.Milliseconds(), DefaultPolicy().IPFailures})
}

func TestReserveAttemptLocked(t *testing.T) {
	redis := &fakeScriptExecutor{results: []any{[]any{int64(1), int64(1500)}}}
	guard := New(redis, &fakeLimiter{}, DefaultPolicy())

	decision, err := guard.ReserveAttempt(context.Background(), "admin", "10.0.0.1")
	if err != nil {
		t.Fatalf("预占失败: %v", err)
	}
	if decision.Outcome != OutcomeLocked || decision.RetryAfter != 1500*time.Millisecond {
		t.Fatalf("期望锁定 1.5s: %+v", decision)
	}
}

func TestReserveAttemptBudgetExhausted(t *testing.T) {
	redis := &fakeScriptExecutor{results: []any{[]any{int64(2), int64(30000)}}}
	guard := New(redis, &fakeLimiter{}, DefaultPolicy())

	decision, err := guard.ReserveAttempt(context.Background(), "admin", "10.0.0.1")
	if err != nil {
		t.Fatalf("预占失败: %v", err)
	}
	if decision.Outcome != OutcomeRateLimited || decision.RetryAfter != 30*time.Second {
		t.Fatalf("预算占满应返回限流: %+v", decision)
	}
}

func TestReserveAttemptUnexpectedResult(t *testing.T) {
	cases := []any{int64(0), []any{int64(0)}, []any{int64(9), int64(0)}}
	for _, result := range cases {
		redis := &fakeScriptExecutor{results: []any{result}}
		guard := New(redis, &fakeLimiter{}, DefaultPolicy())
		if _, err := guard.ReserveAttempt(context.Background(), "admin", "10.0.0.1"); err == nil {
			t.Fatalf("异常返回应报错: %#v", result)
		}
	}
}

func TestReserveAttemptRedisError(t *testing.T) {
	redis := &fakeScriptExecutor{err: errors.New("redis down")}
	guard := New(redis, &fakeLimiter{}, DefaultPolicy())

	if _, err := guard.ReserveAttempt(context.Background(), "admin", "10.0.0.1"); err == nil {
		t.Fatal("Redis 故障应向上传播")
	}
}

func TestReleaseAttemptReleasesPendingKeys(t *testing.T) {
	redis := &fakeScriptExecutor{}
	guard := New(redis, &fakeLimiter{}, DefaultPolicy())

	if err := guard.ReleaseAttempt(context.Background(), "admin", "10.0.0.1"); err != nil {
		t.Fatalf("释放预占失败: %v", err)
	}
	if len(redis.calls) != 1 || redis.calls[0].script != releaseAttemptScript {
		t.Fatalf("应使用释放脚本: %+v", redis.calls)
	}
	ip := ipID("10.0.0.1")
	id := accountID("admin")
	assertKeys(t, redis.calls[0].keys, []string{
		guard.ipKey(ip, "pending"),
		guard.pairKey(id, ip, "pending"),
		guard.accountKey(id, "pending"),
	})
}

func TestReleaseAttemptRedisError(t *testing.T) {
	redis := &fakeScriptExecutor{err: errors.New("redis down")}
	guard := New(redis, &fakeLimiter{}, DefaultPolicy())

	if err := guard.ReleaseAttempt(context.Background(), "admin", "10.0.0.1"); err == nil {
		t.Fatal("Redis 故障应向上传播")
	}
}

func TestRecordFailureReturnsMaxLock(t *testing.T) {
	redis := &fakeScriptExecutor{results: []any{[]any{int64(0), int64(600000), int64(900000)}}}
	guard := New(redis, &fakeLimiter{}, DefaultPolicy())

	decision, err := guard.RecordFailure(context.Background(), "admin", "10.0.0.1")
	if err != nil {
		t.Fatalf("记账失败: %v", err)
	}
	if decision.Outcome != OutcomeLocked || decision.RetryAfter != 15*time.Minute {
		t.Fatalf("期望取最大锁定时长 15m: %+v", decision)
	}
}

// TestRecordFailureSingleRoundTrip 校验三个维度在单次脚本调用内原子记账并释放预占
func TestRecordFailureSingleRoundTrip(t *testing.T) {
	redis := &fakeScriptExecutor{results: []any{[]any{int64(0), int64(600000), int64(900000)}}}
	guard := New(redis, &fakeLimiter{}, DefaultPolicy())

	if _, err := guard.RecordFailure(context.Background(), "admin", "10.0.0.1"); err != nil {
		t.Fatalf("记账失败: %v", err)
	}
	if len(redis.calls) != 1 {
		t.Fatalf("三个维度应在一次脚本调用内记账，实际 %d 次", len(redis.calls))
	}

	call := redis.calls[0]
	if call.script != recordFailureScript {
		t.Fatal("应使用合并后的失败记账脚本")
	}

	ip := ipID("10.0.0.1")
	id := accountID("admin")
	wantKeys := []string{
		guard.ipKey(ip, "lock"), guard.ipKey(ip, "fail"), guard.ipKey(ip, "pending"),
		guard.pairKey(id, ip, "lock"), guard.pairKey(id, ip, "fail"), guard.pairKey(id, ip, "pending"),
		guard.accountKey(id, "lock"), guard.accountKey(id, "fail"), guard.accountKey(id, "pending"),
		guard.accountKey(id, "level"),
	}
	assertKeys(t, call.keys, wantKeys)

	policy := DefaultPolicy()
	wantArgs := []any{
		policy.IPFailWindow.Milliseconds(), policy.IPFailures, policy.IPLock.Milliseconds(),
		policy.PairWindow.Milliseconds(), policy.PairFailures, policy.PairLock.Milliseconds(),
		policy.AccountWindow.Milliseconds(), policy.AccountFailures, policy.AccountLock.Milliseconds(),
		policy.AccountMaxLock.Milliseconds(), policy.LevelTTL.Milliseconds(),
	}
	assertArgs(t, call.args, wantArgs)
}

func TestRecordFailureWithoutAccountRecordsIPOnly(t *testing.T) {
	redis := &fakeScriptExecutor{results: []any{[]any{int64(0)}}}
	guard := New(redis, &fakeLimiter{}, DefaultPolicy())

	decision, err := guard.RecordFailure(context.Background(), "", "10.0.0.1")
	if err != nil {
		t.Fatalf("记账失败: %v", err)
	}
	if decision.Outcome != OutcomeAllowed {
		t.Fatalf("未达阈值应放行: %+v", decision)
	}
	if len(redis.calls) != 1 {
		t.Fatalf("空账号只应记录 IP 维度: %d", len(redis.calls))
	}
	if len(redis.calls[0].keys) != 3 {
		t.Fatalf("空账号应有 IP 的 lock/fail/pending 三个 key: %+v", redis.calls[0].keys)
	}
	if redis.calls[0].keys[1] != guard.ipKey(ipID("10.0.0.1"), "fail") {
		t.Fatalf("key 不是 IP 失败计数: %s", redis.calls[0].keys[1])
	}
}

// TestRecordFailurePairLockOnly 校验只有组合维度触发锁定时按组合锁时长返回
func TestRecordFailurePairLockOnly(t *testing.T) {
	redis := &fakeScriptExecutor{results: []any{[]any{int64(0), int64(600000), int64(0)}}}
	guard := New(redis, &fakeLimiter{}, DefaultPolicy())

	decision, err := guard.RecordFailure(context.Background(), "admin", "10.0.0.1")
	if err != nil {
		t.Fatalf("记账失败: %v", err)
	}
	if decision.Outcome != OutcomeLocked || decision.RetryAfter != 10*time.Minute {
		t.Fatalf("期望锁定 10m: %+v", decision)
	}
}

func TestRecordFailureRedisError(t *testing.T) {
	redis := &fakeScriptExecutor{err: errors.New("redis down")}
	guard := New(redis, &fakeLimiter{}, DefaultPolicy())

	if _, err := guard.RecordFailure(context.Background(), "admin", "10.0.0.1"); err == nil {
		t.Fatal("Redis 故障应向上传播")
	}
}

func TestRecordSuccessReleasesPendingAndClearsAccountKeys(t *testing.T) {
	redis := &fakeScriptExecutor{}
	guard := New(redis, &fakeLimiter{}, DefaultPolicy())

	if err := guard.RecordSuccess(context.Background(), " Admin ", "10.0.0.1"); err != nil {
		t.Fatalf("清零失败: %v", err)
	}
	if len(redis.calls) != 1 || redis.calls[0].script != recordSuccessScript {
		t.Fatalf("应使用成功清零脚本: %+v", redis.calls)
	}

	ip := ipID("10.0.0.1")
	id := accountID("admin")
	keys := redis.calls[0].keys
	assertKeys(t, keys, []string{
		guard.ipKey(ip, "pending"),
		guard.pairKey(id, ip, "pending"), guard.pairKey(id, ip, "fail"), guard.pairKey(id, ip, "lock"),
		guard.accountKey(id, "pending"), guard.accountKey(id, "fail"), guard.accountKey(id, "lock"), guard.accountKey(id, "level"),
	})
	for _, key := range keys {
		if strings.HasSuffix(key, ":fail") && strings.Contains(key, ":ip:") {
			t.Fatalf("成功登录不能清 IP 失败计数: %s", key)
		}
	}
}

func TestRecordSuccessSkipsEmptyAccount(t *testing.T) {
	redis := &fakeScriptExecutor{}
	guard := New(redis, &fakeLimiter{}, DefaultPolicy())

	if err := guard.RecordSuccess(context.Background(), "", "10.0.0.1"); err != nil {
		t.Fatalf("空账号不应报错: %v", err)
	}
	if len(redis.calls) != 0 {
		t.Fatalf("空账号不应调用脚本: %+v", redis.calls)
	}
}

func TestRecordSuccessRedisError(t *testing.T) {
	redis := &fakeScriptExecutor{err: errors.New("redis down")}
	guard := New(redis, &fakeLimiter{}, DefaultPolicy())

	if err := guard.RecordSuccess(context.Background(), "admin", "10.0.0.1"); err == nil {
		t.Fatal("Redis 故障应向上传播")
	}
}

func TestAccountIDNormalizes(t *testing.T) {
	if accountID(" Admin ") != accountID("admin") {
		t.Fatal("大小写与空白应归一化")
	}
	if accountID("") != "" {
		t.Fatal("空账号应返回空串")
	}
	if len(accountID("admin")) != idLength {
		t.Fatalf("摘要长度应为 %d", idLength)
	}
}

func TestIPIDGroupsIPv6By64(t *testing.T) {
	sameA := ipID("2001:db8:1:2::1")
	sameB := ipID("2001:db8:1:2::ffff")
	if sameA != sameB {
		t.Fatal("同一 /64 的 IPv6 地址应归为同一标识")
	}
	if sameA == ipID("2001:db8:1:3::1") {
		t.Fatal("不同 /64 的 IPv6 地址不应归为同一标识")
	}
}

func TestIPIDUnmapsIPv4(t *testing.T) {
	if ipID("::ffff:127.0.0.1") != ipID("127.0.0.1") {
		t.Fatal("IPv4-mapped IPv6 应与 IPv4 归为同一标识")
	}
}

func TestIPIDInvalidIsStable(t *testing.T) {
	if ipID("not-an-ip") != ipID("") {
		t.Fatal("非法 IP 应归入统一标识")
	}
}

func TestTruncate(t *testing.T) {
	value := strings.Repeat("中", maxLogRunes+5)
	got := truncate(value, maxLogRunes)
	if len([]rune(got)) != maxLogRunes {
		t.Fatalf("应截断到 %d 个字符: %d", maxLogRunes, len([]rune(got)))
	}
	if truncate(" admin ", maxLogRunes) != "admin" {
		t.Fatal("应去掉首尾空白")
	}
}

func assertKeys(t *testing.T, got, want []string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("key 数量不符: got=%d want=%d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("keys[%d] = %s, want %s", i, got[i], want[i])
		}
	}
}

func assertArgs(t *testing.T, got, want []any) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("argv 数量不符: got=%d want=%d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("args[%d] = %v, want %v", i, got[i], want[i])
		}
	}
}

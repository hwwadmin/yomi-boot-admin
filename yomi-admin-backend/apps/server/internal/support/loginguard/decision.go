package loginguard

import "time"

// Outcome 表示一次登录保护判定的结果
type Outcome int

const (
	// OutcomeAllowed 允许继续登录流程
	OutcomeAllowed Outcome = iota
	// OutcomeLocked 命中临时锁定
	OutcomeLocked
	// OutcomeRateLimited 命中 IP 级尝试配额
	OutcomeRateLimited
)

// Decision 是登录保护判定结果
type Decision struct {
	Outcome    Outcome
	RetryAfter time.Duration
}

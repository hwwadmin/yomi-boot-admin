package auth

import (
	"math"
	"time"

	"gitee.com/hq_project/yomi-boot/component/core/errors"
)

var (
	// ErrLoginRateLimited 登录请求触发 IP 级尝试配额
	ErrLoginRateLimited = errors.NewBiz("ADMIN_LOGIN_RATE_LIMITED", "操作过于频繁，请稍后重试")
	// ErrLoginLocked 登录失败次数过多，进入临时锁定
	ErrLoginLocked = errors.NewBiz("ADMIN_LOGIN_LOCKED", "登录尝试次数过多，请稍后重试")
	// ErrLoginGuardUnavailable 登录保护依赖的存储不可用
	ErrLoginGuardUnavailable = errors.NewBiz("ADMIN_LOGIN_GUARD_UNAVAILABLE", "登录保护暂不可用，请稍后重试")
)

// loginLockedError 生成带剩余等待秒数的锁定异常
func loginLockedError(retryAfter time.Duration) error {
	seconds := int(math.Ceil(retryAfter.Seconds()))
	if seconds < 1 {
		seconds = 1
	}
	return errors.Sentinel(errors.NewBizf(ErrLoginLocked.Code, "登录尝试次数过多，请 %d 秒后重试", seconds))
}

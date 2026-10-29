// Package auth 认证相关编排
package auth

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"time"

	"gitee.com/hq_project/yomi-boot/component/common/captcha"
	"gitee.com/hq_project/yomi-boot/component/common/redisratelimit"
	"gitee.com/hq_project/yomi-boot/component/core/errors"
)

// 登录图像验证码业务异常
var (
	// ErrCaptchaRequired 未填写验证码
	ErrCaptchaRequired = errors.NewBiz("ADMIN_CAPTCHA_REQUIRED", "请输入验证码")
	// ErrCaptchaInvalid 验证码不存在、已过期或已被使用
	ErrCaptchaInvalid = errors.NewBiz("ADMIN_CAPTCHA_INVALID", "验证码已失效，请刷新后重试")
	// ErrCaptchaMismatch 验证码答案错误
	ErrCaptchaMismatch = errors.NewBiz("ADMIN_CAPTCHA_MISMATCH", "验证码错误")
	// ErrCaptchaTooFrequent 同一 IP 刷新验证码过于频繁
	ErrCaptchaTooFrequent = errors.NewBiz("ADMIN_CAPTCHA_TOO_FREQUENT", "验证码刷新过于频繁，请稍后重试")
	// ErrCaptchaGenerateFailed 生成验证码失败
	ErrCaptchaGenerateFailed = errors.NewBiz("ADMIN_CAPTCHA_GENERATE_FAILED", "验证码生成失败，请稍后重试")
	// ErrCaptchaUnavailable 验证码存储、限流或校验链路不可用
	ErrCaptchaUnavailable = errors.NewBiz("ADMIN_CAPTCHA_UNAVAILABLE", "验证码服务暂不可用，请稍后重试")
)

// captchaRateLimitPrefix 是验证码生成限流的逻辑 key 前缀
const captchaRateLimitPrefix = "captcha:login:"

// CaptchaUsecase 图像验证码生成与校验
type CaptchaUsecase struct {
	service         captcha.Service
	limiter         redisratelimit.LimitService
	refreshInterval time.Duration
	enforce         bool
}

// NewCaptchaUsecase 创建验证码用例；refreshInterval<=0 表示不限制生成频率，enforce=false 表示登录跳过校验
func NewCaptchaUsecase(service captcha.Service, limiter redisratelimit.LimitService, refreshInterval time.Duration, enforce bool) *CaptchaUsecase {
	return &CaptchaUsecase{service: service, limiter: limiter, refreshInterval: refreshInterval, enforce: enforce}
}

// CaptchaOutput 验证码出参
type CaptchaOutput struct {
	ID        string `json:"id"`
	Image     string `json:"image"`
	ExpiresIn int    `json:"expiresIn"`
}

// Generate 生成验证码；答案只写 Redis，不返回给前端
func (uc *CaptchaUsecase) Generate(ctx context.Context, clientIP string) (*CaptchaOutput, error) {
	if err := uc.checkRefreshInterval(ctx, clientIP); err != nil {
		return nil, err
	}

	challenge, err := uc.service.Generate(ctx)
	if err != nil {
		return nil, errors.WrapBiz(err, ErrCaptchaGenerateFailed.Code, ErrCaptchaGenerateFailed.Message)
	}
	return &CaptchaOutput{ID: challenge.ID, Image: challenge.Image, ExpiresIn: challenge.ExpiresIn}, nil
}

// Verify 校验并消费验证码；未启用时直接放行
func (uc *CaptchaUsecase) Verify(ctx context.Context, id, code string) error {
	if !uc.enforce {
		return nil
	}
	if strings.TrimSpace(id) == "" || strings.TrimSpace(code) == "" {
		return errors.Sentinel(ErrCaptchaRequired)
	}
	if err := uc.service.Verify(ctx, id, code); err != nil {
		return wrapCaptchaError(err)
	}
	return nil
}

// checkRefreshInterval 按 IP 限制生成频率：窗口内只允许生成一次；未配置间隔时直接放行
func (uc *CaptchaUsecase) checkRefreshInterval(ctx context.Context, clientIP string) error {
	if uc.refreshInterval <= 0 {
		return nil
	}

	rule := redisratelimit.Rule{Limit: 1, Window: uc.refreshInterval}
	err := uc.limiter.Allow(ctx, captchaRateLimitKey(clientIP), rule)
	if err == nil {
		return nil
	}
	if errors.Is(err, redisratelimit.ErrRateLimited) {
		return errors.Sentinel(errors.NewBizf(
			ErrCaptchaTooFrequent.Code,
			"验证码刷新过于频繁，请 %d 秒后重试",
			int(uc.refreshInterval/time.Second),
		))
	}
	return errors.WrapBiz(err, ErrCaptchaUnavailable.Code, ErrCaptchaUnavailable.Message)
}

// wrapCaptchaError 把 captcha 库哨兵映射为登录业务异常
func wrapCaptchaError(err error) error {
	switch {
	case errors.Is(err, captcha.ErrMismatch):
		return errors.Sentinel(ErrCaptchaMismatch)
	case errors.Is(err, captcha.ErrExpired), errors.Is(err, captcha.ErrInvalidArgument):
		return errors.Sentinel(ErrCaptchaInvalid)
	default:
		return errors.WrapBiz(err, ErrCaptchaUnavailable.Code, ErrCaptchaUnavailable.Message)
	}
}

// captchaRateLimitKey 用 IP 摘要作为限流 key，避免 IP 中的特殊字符影响 Redis key 校验
func captchaRateLimitKey(clientIP string) string {
	ip := strings.TrimSpace(clientIP)
	if ip == "" {
		ip = "unknown"
	}
	sum := sha256.Sum256([]byte(ip))
	return captchaRateLimitPrefix + hex.EncodeToString(sum[:8])
}

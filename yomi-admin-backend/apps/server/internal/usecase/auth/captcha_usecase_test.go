package auth

import (
	"context"
	"testing"
	"time"

	"github.com/hwwadmin/yomi-boot/component/common/captcha"
	"github.com/hwwadmin/yomi-boot/component/common/redisratelimit"
	"github.com/hwwadmin/yomi-boot/component/core/errors"
)

// fakeCaptchaService 是验证码服务测试桩
type fakeCaptchaService struct {
	challenge     captcha.Challenge
	generateErr   error
	verifyErr     error
	generateCalls int
	verifyCalls   int
	verifyID      string
	verifyCode    string
}

func (f *fakeCaptchaService) Generate(context.Context) (captcha.Challenge, error) {
	f.generateCalls++
	return f.challenge, f.generateErr
}

func (f *fakeCaptchaService) Verify(_ context.Context, id, code string) error {
	f.verifyCalls++
	f.verifyID = id
	f.verifyCode = code
	return f.verifyErr
}

// fakeLimiter 是限流服务测试桩
type fakeLimiter struct {
	err   error
	key   string
	rule  redisratelimit.Rule
	calls int
}

func (f *fakeLimiter) Allow(_ context.Context, key string, rule redisratelimit.Rule) error {
	f.calls++
	f.key = key
	f.rule = rule
	return f.err
}

func (f *fakeLimiter) AllowN(context.Context, string, int64, redisratelimit.Rule) error {
	return errors.New("AllowN 不应被调用")
}

func TestCaptchaGenerateReturnsChallenge(t *testing.T) {
	service := &fakeCaptchaService{challenge: captcha.Challenge{
		ID:        "cid",
		Image:     "data:image/png;base64,xx",
		ExpiresIn: 180,
	}}
	limiter := &fakeLimiter{}
	uc := NewCaptchaUsecase(service, limiter, 10*time.Second, true)

	out, err := uc.Generate(context.Background(), "192.168.1.10")
	if err != nil {
		t.Fatalf("生成验证码失败: %v", err)
	}
	if out.ID != "cid" || out.Image != "data:image/png;base64,xx" || out.ExpiresIn != 180 {
		t.Fatalf("验证码出参不符: %+v", out)
	}
	if limiter.calls != 1 {
		t.Fatalf("应按 IP 限流一次: %d", limiter.calls)
	}
	if limiter.rule != (redisratelimit.Rule{Limit: 1, Window: 10 * time.Second}) {
		t.Fatalf("限流规则不符: %+v", limiter.rule)
	}
	if limiter.key != captchaRateLimitKey("192.168.1.10") {
		t.Fatalf("限流 key 不符: %s", limiter.key)
	}
}

func TestCaptchaGenerateThrottled(t *testing.T) {
	service := &fakeCaptchaService{}
	limiter := &fakeLimiter{err: errors.Sentinel(redisratelimit.ErrRateLimited)}
	uc := NewCaptchaUsecase(service, limiter, 10*time.Second, true)

	_, err := uc.Generate(context.Background(), "192.168.1.10")
	if !errors.Is(err, ErrCaptchaTooFrequent) {
		t.Fatalf("期望 ErrCaptchaTooFrequent，实际 %v", err)
	}
	if service.generateCalls != 0 {
		t.Fatalf("限流后不应生成验证码: %d", service.generateCalls)
	}
}

func TestCaptchaGenerateLimiterFailure(t *testing.T) {
	service := &fakeCaptchaService{}
	limiter := &fakeLimiter{err: errors.New("redis 不可用")}
	uc := NewCaptchaUsecase(service, limiter, 10*time.Second, true)

	_, err := uc.Generate(context.Background(), "192.168.1.10")
	if !errors.Is(err, ErrCaptchaUnavailable) {
		t.Fatalf("期望 ErrCaptchaUnavailable，实际 %v", err)
	}
	if service.generateCalls != 0 {
		t.Fatalf("限流故障时不应生成验证码: %d", service.generateCalls)
	}
}

func TestCaptchaGenerateWithoutRefreshInterval(t *testing.T) {
	service := &fakeCaptchaService{challenge: captcha.Challenge{ID: "cid"}}
	limiter := &fakeLimiter{err: errors.Sentinel(redisratelimit.ErrRateLimited)}
	uc := NewCaptchaUsecase(service, limiter, 0, true)

	if _, err := uc.Generate(context.Background(), "192.168.1.10"); err != nil {
		t.Fatalf("未配置间隔时不应限流: %v", err)
	}
	if limiter.calls != 0 {
		t.Fatalf("未配置间隔时不应触达限流: %d", limiter.calls)
	}
}

func TestCaptchaRateLimitKeyIsolatesIP(t *testing.T) {
	if captchaRateLimitKey("10.0.0.1") == captchaRateLimitKey("10.0.0.2") {
		t.Fatal("不同 IP 应使用不同限流 key")
	}
	if captchaRateLimitKey("10.0.0.1") != captchaRateLimitKey(" 10.0.0.1 ") {
		t.Fatal("同一 IP 的限流 key 应稳定")
	}
}

func TestCaptchaGenerateFailure(t *testing.T) {
	service := &fakeCaptchaService{generateErr: errors.Wrap(captcha.ErrGenerateFailed, "渲染失败")}
	uc := NewCaptchaUsecase(service, &fakeLimiter{}, 10*time.Second, true)

	_, err := uc.Generate(context.Background(), "192.168.1.10")
	if !errors.Is(err, ErrCaptchaGenerateFailed) {
		t.Fatalf("期望 ErrCaptchaGenerateFailed，实际 %v", err)
	}
}

func TestCaptchaVerifyRequiresCode(t *testing.T) {
	service := &fakeCaptchaService{}
	uc := NewCaptchaUsecase(service, &fakeLimiter{}, 10*time.Second, true)

	err := uc.Verify(context.Background(), "", "1234")
	if !errors.Is(err, ErrCaptchaRequired) {
		t.Fatalf("期望 ErrCaptchaRequired，实际 %v", err)
	}
	if service.verifyCalls != 0 {
		t.Fatalf("参数缺失不应触达验证码服务: %d", service.verifyCalls)
	}
}

func TestCaptchaVerifyPassesIDAndCode(t *testing.T) {
	service := &fakeCaptchaService{}
	uc := NewCaptchaUsecase(service, &fakeLimiter{}, 10*time.Second, true)

	if err := uc.Verify(context.Background(), "cid", "1234"); err != nil {
		t.Fatalf("校验验证码失败: %v", err)
	}
	if service.verifyCalls != 1 || service.verifyID != "cid" || service.verifyCode != "1234" {
		t.Fatalf("验证码入参不符: calls=%d id=%q code=%q", service.verifyCalls, service.verifyID, service.verifyCode)
	}
}

func TestCaptchaVerifyMapsServiceErrors(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want error
	}{
		{name: "答案不符", err: errors.Sentinel(captcha.ErrMismatch), want: ErrCaptchaMismatch},
		{name: "已过期", err: errors.Sentinel(captcha.ErrExpired), want: ErrCaptchaInvalid},
		{name: "参数非法", err: errors.Sentinel(captcha.ErrInvalidArgument), want: ErrCaptchaInvalid},
		{name: "存储故障", err: errors.New("redis 不可用"), want: ErrCaptchaUnavailable},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			service := &fakeCaptchaService{verifyErr: tc.err}
			uc := NewCaptchaUsecase(service, &fakeLimiter{}, 10*time.Second, true)

			err := uc.Verify(context.Background(), "cid", "1234")
			if !errors.Is(err, tc.want) {
				t.Fatalf("期望 %v，实际 %v", tc.want, err)
			}
		})
	}
}

func TestCaptchaVerifyDisabledSkipsService(t *testing.T) {
	service := &fakeCaptchaService{verifyErr: errors.Sentinel(captcha.ErrMismatch)}
	uc := NewCaptchaUsecase(service, &fakeLimiter{}, 10*time.Second, false)

	if err := uc.Verify(context.Background(), "", ""); err != nil {
		t.Fatalf("未启用时不应校验: %v", err)
	}
	if service.verifyCalls != 0 {
		t.Fatalf("未启用时不应触达验证码服务: %d", service.verifyCalls)
	}
}

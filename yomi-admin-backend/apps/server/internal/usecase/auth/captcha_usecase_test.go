package auth

import (
	"context"
	"testing"

	"gitee.com/hq_project/yomi-boot/component/common/captcha"
	"gitee.com/hq_project/yomi-boot/component/core/errors"
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

func TestCaptchaGenerateReturnsChallenge(t *testing.T) {
	service := &fakeCaptchaService{challenge: captcha.Challenge{
		ID:        "cid",
		Image:     "data:image/png;base64,xx",
		ExpiresIn: 180,
	}}
	uc := NewCaptchaUsecase(service, true)

	out, err := uc.Generate(context.Background())
	if err != nil {
		t.Fatalf("生成验证码失败: %v", err)
	}
	if out.ID != "cid" || out.Image != "data:image/png;base64,xx" || out.ExpiresIn != 180 {
		t.Fatalf("验证码出参不符: %+v", out)
	}
}

func TestCaptchaGenerateFailure(t *testing.T) {
	service := &fakeCaptchaService{generateErr: errors.Wrap(captcha.ErrGenerateFailed, "渲染失败")}
	uc := NewCaptchaUsecase(service, true)

	_, err := uc.Generate(context.Background())
	if !errors.Is(err, ErrCaptchaGenerateFailed) {
		t.Fatalf("期望 ErrCaptchaGenerateFailed，实际 %v", err)
	}
}

func TestCaptchaVerifyRequiresCode(t *testing.T) {
	service := &fakeCaptchaService{}
	uc := NewCaptchaUsecase(service, true)

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
	uc := NewCaptchaUsecase(service, true)

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
			uc := NewCaptchaUsecase(service, true)

			err := uc.Verify(context.Background(), "cid", "1234")
			if !errors.Is(err, tc.want) {
				t.Fatalf("期望 %v，实际 %v", tc.want, err)
			}
		})
	}
}

func TestCaptchaVerifyDisabledSkipsService(t *testing.T) {
	service := &fakeCaptchaService{verifyErr: errors.Sentinel(captcha.ErrMismatch)}
	uc := NewCaptchaUsecase(service, false)

	if err := uc.Verify(context.Background(), "", ""); err != nil {
		t.Fatalf("未启用时不应校验: %v", err)
	}
	if service.verifyCalls != 0 {
		t.Fatalf("未启用时不应触达验证码服务: %d", service.verifyCalls)
	}
}

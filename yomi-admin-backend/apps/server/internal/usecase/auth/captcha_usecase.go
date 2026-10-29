// Package auth 认证相关编排
package auth

import (
	"context"
	"strings"

	"gitee.com/hq_project/yomi-boot/component/common/captcha"
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
	// ErrCaptchaGenerateFailed 生成验证码失败
	ErrCaptchaGenerateFailed = errors.NewBiz("ADMIN_CAPTCHA_GENERATE_FAILED", "验证码生成失败，请稍后重试")
	// ErrCaptchaUnavailable 验证码存储或校验链路不可用
	ErrCaptchaUnavailable = errors.NewBiz("ADMIN_CAPTCHA_UNAVAILABLE", "验证码服务暂不可用，请稍后重试")
)

// CaptchaUsecase 图像验证码生成与校验
type CaptchaUsecase struct {
	service captcha.Service
	enforce bool
}

// NewCaptchaUsecase 创建验证码用例；enforce=false 时登录跳过校验，仅用于本地或测试
func NewCaptchaUsecase(service captcha.Service, enforce bool) *CaptchaUsecase {
	return &CaptchaUsecase{service: service, enforce: enforce}
}

// CaptchaOutput 验证码出参
type CaptchaOutput struct {
	ID        string `json:"id"`
	Image     string `json:"image"`
	ExpiresIn int    `json:"expiresIn"`
}

// Generate 生成验证码；答案只写 Redis，不返回给前端
func (uc *CaptchaUsecase) Generate(ctx context.Context) (*CaptchaOutput, error) {
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

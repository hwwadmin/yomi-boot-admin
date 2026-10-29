// Package auth 认证相关编排
package auth

// LoginConfigOutput 登录页公开配置
type LoginConfigOutput struct {
	CaptchaRequired bool `json:"captchaRequired"` // CaptchaRequired 表示登录是否需要图形验证码
}

// LoginConfigUsecase 登录页公开配置编排
type LoginConfigUsecase struct {
	captchaRequired bool
}

// NewLoginConfigUsecase 创建登录页配置用例
func NewLoginConfigUsecase(captchaRequired bool) *LoginConfigUsecase {
	return &LoginConfigUsecase{captchaRequired: captchaRequired}
}

// Execute 返回登录页所需的公开配置
func (uc *LoginConfigUsecase) Execute() *LoginConfigOutput {
	return &LoginConfigOutput{CaptchaRequired: uc.captchaRequired}
}

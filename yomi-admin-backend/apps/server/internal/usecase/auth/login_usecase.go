// Package auth 认证相关编排
package auth

import (
	"context"
	rbaccontract "yomi-admin-backend/domains/rbac/service_contract"
	"yomi-admin-backend/domains/user/service_contract"

	userinput "yomi-admin-backend/domains/user/pkg/shared/input"

	"gitee.com/hq_project/yomi-boot/component/common/token"
	"gitee.com/hq_project/yomi-boot/component/core/errors"
)

// captchaVerifier 校验登录图像验证码
type captchaVerifier interface {
	Verify(ctx context.Context, id, code string) error
}

// LoginUsecase 登录编排：验证码 + 领域鉴权 + 聚合权限 + 签发 Token
type LoginUsecase struct {
	users   service_contract.UserService
	auth    rbaccontract.RbacAuthService
	token   token.TokenService
	captcha captchaVerifier
}

// NewLoginUsecase 创建登录用例
func NewLoginUsecase(users service_contract.UserService, auth rbaccontract.RbacAuthService, tokenService token.TokenService, captcha captchaVerifier) *LoginUsecase {
	return &LoginUsecase{users: users, auth: auth, token: tokenService, captcha: captcha}
}

// LoginInput HTTP 登录入参
type LoginInput struct {
	Username    string `json:"username"`
	Password    string `json:"password"`
	CaptchaID   string `json:"captchaId"`
	CaptchaCode string `json:"captchaCode"`
}

// LoginOutput HTTP 登录出参
type LoginOutput struct {
	Token string        `json:"token"`
	User  *IdentityView `json:"user"`
}

// Execute 编排登录
func (uc *LoginUsecase) Execute(ctx context.Context, in *LoginInput) (*LoginOutput, error) {
	if err := uc.captcha.Verify(ctx, in.CaptchaID, in.CaptchaCode); err != nil {
		return nil, err
	}

	out, err := uc.users.Authenticate(ctx, &userinput.AuthenticateInput{
		Username: in.Username,
		Password: in.Password,
	})
	if err != nil {
		return nil, err
	}

	identity, err := buildIdentityView(ctx, out.User, uc.auth)
	if err != nil {
		return nil, err
	}

	session := &token.TokenSession{
		UserID:     out.User.ID,
		ClientType: "admin",
	}
	session.SetContextValue("username", out.User.Username)
	session.SetContextValue("nickname", out.User.Nickname)

	tokenValue, err := uc.token.Issue(ctx, session)
	if err != nil {
		return nil, errors.Wrap(err, "签发令牌失败")
	}

	return &LoginOutput{Token: tokenValue, User: identity}, nil
}

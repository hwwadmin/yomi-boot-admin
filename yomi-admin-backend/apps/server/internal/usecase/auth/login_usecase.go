// Package auth 认证相关编排
package auth

import (
	"context"

	loginguard "yomi-admin-backend/apps/server/internal/support/loginguard"
	rbaccontract "yomi-admin-backend/domains/rbac/service_contract"
	"yomi-admin-backend/domains/user/service_contract"

	userinput "yomi-admin-backend/domains/user/pkg/shared/input"

	"gitee.com/hq_project/yomi-boot/component/common/token"
	"gitee.com/hq_project/yomi-boot/component/core/errors"
	"gitee.com/hq_project/yomi-boot/component/core/logger"
)

// captchaVerifier 校验登录图像验证码
type captchaVerifier interface {
	Verify(ctx context.Context, id, code string) error
}

// loginGuard 负责登录失败预算与临时锁定
type loginGuard interface {
	BeforeAttempt(ctx context.Context, account, clientIP string) (loginguard.Decision, error)
	RecordFailure(ctx context.Context, account, clientIP string) (loginguard.Decision, error)
	RecordSuccess(ctx context.Context, account, clientIP string) error
}

// LoginUsecase 登录编排：限流锁定 + 验证码 + 领域鉴权 + 聚合权限 + 签发 Token
type LoginUsecase struct {
	users   service_contract.UserService
	auth    rbaccontract.RbacAuthService
	token   token.TokenService
	captcha captchaVerifier
	guard   loginGuard
}

// NewLoginUsecase 创建登录用例
func NewLoginUsecase(
	users service_contract.UserService,
	auth rbaccontract.RbacAuthService,
	tokenService token.TokenService,
	captcha captchaVerifier,
	guard loginGuard,
) *LoginUsecase {
	return &LoginUsecase{users: users, auth: auth, token: tokenService, captcha: captcha, guard: guard}
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
func (uc *LoginUsecase) Execute(ctx context.Context, in *LoginInput, clientIP string) (*LoginOutput, error) {
	decision, err := uc.guard.BeforeAttempt(ctx, in.Username, clientIP)
	if err != nil {
		return nil, errors.WrapBiz(err, ErrLoginGuardUnavailable.Code, ErrLoginGuardUnavailable.Message)
	}
	switch decision.Outcome {
	case loginguard.OutcomeRateLimited:
		return nil, errors.Sentinel(ErrLoginRateLimited)
	case loginguard.OutcomeLocked:
		return nil, loginLockedError(decision.RetryAfter)
	}

	if err := uc.captcha.Verify(ctx, in.CaptchaID, in.CaptchaCode); err != nil {
		return nil, err
	}

	out, err := uc.users.Authenticate(ctx, &userinput.AuthenticateInput{
		Username: in.Username,
		Password: in.Password,
	})
	if err != nil {
		if errors.Is(err, service_contract.ErrInvalidCredential) {
			failure, guardErr := uc.guard.RecordFailure(ctx, in.Username, clientIP)
			if guardErr != nil {
				logger.Warn(ctx, "记录登录失败计数失败", "error", guardErr)
			} else if failure.Outcome == loginguard.OutcomeLocked {
				return nil, loginLockedError(failure.RetryAfter)
			}
		}
		if errors.Is(err, service_contract.ErrUserDisabled) {
			// 密码正确但账号被禁用，不计失败；清空该账号的历史失败计数
			if guardErr := uc.guard.RecordSuccess(ctx, in.Username, clientIP); guardErr != nil {
				logger.Warn(ctx, "清空账号登录失败计数失败", "error", guardErr)
			}
		}
		return nil, err
	}

	if err := uc.guard.RecordSuccess(ctx, in.Username, clientIP); err != nil {
		logger.Warn(ctx, "清空账号登录失败计数失败", "error", err)
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

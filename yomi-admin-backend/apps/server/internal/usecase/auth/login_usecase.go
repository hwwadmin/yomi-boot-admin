// Package auth 认证相关编排
package auth

import (
	"context"
	"yomi-admin-backend/domains/user/service_contract"

	loginguard "yomi-admin-backend/apps/server/internal/support/loginguard"
	rbaccontract "yomi-admin-backend/domains/rbac/service_contract"
	userinput "yomi-admin-backend/domains/user/pkg/shared/input"

	"gitee.com/hq_project/yomi-boot/component/common/token"
	"gitee.com/hq_project/yomi-boot/component/core/bc/passwd"
	"gitee.com/hq_project/yomi-boot/component/core/errors"
	"gitee.com/hq_project/yomi-boot/component/core/logger"
)

const (
	// maxLoginUsernameBytes 与用户表 username 列宽（VARCHAR(64)）一致
	maxLoginUsernameBytes = 64
	// maxUserAgentBytes 是写入令牌上下文的 User-Agent 上限，避免异常长头撑大会话快照
	maxUserAgentBytes = 512
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

// validateLoginInput 限制登录入参长度：超长输入直接拒绝，不进入限流、领域层与数据库。
//
// 用户名上限与用户表列宽一致；密码上限取自 bcrypt 的 72 字节（超过必然校验失败）。
// 空值仍由领域层按凭证错误处理，这里只拦长度。
func validateLoginInput(in *LoginInput) error {
	if len(in.Username) > maxLoginUsernameBytes || len(in.Password) > passwd.MaxPasswordBytes {
		return errors.Sentinel(service_contract.ErrInvalidInput)
	}
	return nil
}

// Execute 编排登录
func (uc *LoginUsecase) Execute(ctx context.Context, in *LoginInput, clientIP, userAgent string) (*LoginOutput, error) {
	if err := validateLoginInput(in); err != nil {
		return nil, err
	}
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

	identity, err := buildIdentityView(ctx, out.User, uc.auth)
	if err != nil {
		return nil, err
	}

	session := &token.TokenSession{
		UserID:     out.User.ID,
		ClientType: "admin",
		LoginIP:    clientIP,
	}
	session.SetContextValue("username", out.User.Username)
	session.SetContextValue("nickname", out.User.Nickname)
	session.SetContextValue("user_agent", truncateUserAgent(userAgent))

	tokenValue, err := uc.token.Issue(ctx, session)
	if err != nil {
		return nil, errors.Wrap(err, "签发令牌失败")
	}

	// 整条登录链路成功后才清空失败预算：聚合权限或签发令牌失败时保留计数，
	// 下一次尝试仍受既有预算约束。清零失败只记日志，不影响登录返回。
	if err := uc.guard.RecordSuccess(ctx, in.Username, clientIP); err != nil {
		logger.Warn(ctx, "清空账号登录失败计数失败", "error", err)
	}

	return &LoginOutput{Token: tokenValue, User: identity}, nil
}

// truncateUserAgent 按字节截断 User-Agent；仅用于会话展示，不参与鉴权判断
func truncateUserAgent(userAgent string) string {
	if len(userAgent) <= maxUserAgentBytes {
		return userAgent
	}
	return userAgent[:maxUserAgentBytes]
}

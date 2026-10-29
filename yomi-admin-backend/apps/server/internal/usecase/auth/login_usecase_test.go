package auth

import (
	"context"
	"strings"
	"testing"
	"time"
	"yomi-admin-backend/apps/server/internal/support/loginguard"

	rbacentity "yomi-admin-backend/domains/rbac/pkg/shared/entity"
	rbaccontract "yomi-admin-backend/domains/rbac/service_contract"
	userentity "yomi-admin-backend/domains/user/pkg/shared/entity"
	userinput "yomi-admin-backend/domains/user/pkg/shared/input"
	useroutput "yomi-admin-backend/domains/user/pkg/shared/output"
	usercontract "yomi-admin-backend/domains/user/service_contract"

	"gitee.com/hq_project/yomi-boot/component/common/token"
	"gitee.com/hq_project/yomi-boot/component/core/errors"
)

// fakeUserService 是用户领域服务测试桩
type fakeUserService struct {
	usercontract.UserService
	calls int
	got   *userinput.AuthenticateInput
	out   *useroutput.AuthenticateOutput
	err   error
}

func (f *fakeUserService) Authenticate(_ context.Context, in *userinput.AuthenticateInput) (*useroutput.AuthenticateOutput, error) {
	f.calls++
	f.got = in
	return f.out, f.err
}

// fakeAuthService 是 RBAC 授权查询测试桩
type fakeAuthService struct {
	rbaccontract.RbacAuthService
	permissions []string
	menus       []*rbacentity.Permission
}

func (f *fakeAuthService) ListPermissionCodes(context.Context, int64) ([]string, error) {
	return f.permissions, nil
}

func (f *fakeAuthService) ListMenus(context.Context, int64) ([]*rbacentity.Permission, error) {
	return f.menus, nil
}

// fakeTokenService 是令牌服务测试桩；可注入签发失败
type fakeTokenService struct {
	token.TokenService
	value   string
	err     error
	session *token.TokenSession
}

func (f *fakeTokenService) Issue(_ context.Context, session *token.TokenSession) (string, error) {
	f.session = session
	if f.err != nil {
		return "", f.err
	}
	return f.value, nil
}

// fakeCaptchaVerifier 是登录验证码测试桩
type fakeCaptchaVerifier struct {
	err   error
	id    string
	code  string
	calls int
}

func (f *fakeCaptchaVerifier) Verify(_ context.Context, id, code string) error {
	f.calls++
	f.id = id
	f.code = code
	return f.err
}

// fakeLoginGuard 是登录失败预算测试桩
type fakeLoginGuard struct {
	beforeDecision  loginguard.Decision
	beforeErr       error
	failureDecision loginguard.Decision
	failureErr      error
	successErr      error

	beforeCalls  int
	failureCalls int
	successCalls int
	gotAccount   string
	gotIP        string
}

func (f *fakeLoginGuard) BeforeAttempt(_ context.Context, account, clientIP string) (loginguard.Decision, error) {
	f.beforeCalls++
	f.gotAccount = account
	f.gotIP = clientIP
	return f.beforeDecision, f.beforeErr
}

func (f *fakeLoginGuard) RecordFailure(_ context.Context, account, clientIP string) (loginguard.Decision, error) {
	f.failureCalls++
	f.gotAccount = account
	f.gotIP = clientIP
	return f.failureDecision, f.failureErr
}

func (f *fakeLoginGuard) RecordSuccess(_ context.Context, account, clientIP string) error {
	f.successCalls++
	f.gotAccount = account
	f.gotIP = clientIP
	return f.successErr
}

func TestLoginVerifiesCaptchaThenIssuesToken(t *testing.T) {
	users := &fakeUserService{out: &useroutput.AuthenticateOutput{User: &userentity.User{
		ID:       7,
		Username: "admin",
		Nickname: "超级管理员",
	}}}
	auth := &fakeAuthService{
		permissions: []string{"*"},
		menus: []*rbacentity.Permission{
			{Code: "system:user", Name: "系统用户", UIPath: "/system/user"},
		},
	}
	tokens := &fakeTokenService{value: "tok"}
	captcha := &fakeCaptchaVerifier{}
	guard := &fakeLoginGuard{}

	uc := NewLoginUsecase(users, auth, tokens, captcha, guard)
	out, err := uc.Execute(context.Background(), &LoginInput{
		Username:    "admin",
		Password:    "secret",
		CaptchaID:   "cid",
		CaptchaCode: "1234",
	}, "10.0.0.1")
	if err != nil {
		t.Fatalf("登录失败: %v", err)
	}

	if guard.beforeCalls != 1 || guard.gotIP != "10.0.0.1" || guard.gotAccount != "admin" {
		t.Fatalf("登录前置校验入参不符: %+v", guard)
	}
	if captcha.calls != 1 || captcha.id != "cid" || captcha.code != "1234" {
		t.Fatalf("验证码校验入参不符: calls=%d id=%q code=%q", captcha.calls, captcha.id, captcha.code)
	}
	if users.calls != 1 || users.got == nil || users.got.Username != "admin" || users.got.Password != "secret" {
		t.Fatalf("领域鉴权入参不符: calls=%d got=%+v", users.calls, users.got)
	}
	if guard.successCalls != 1 {
		t.Fatalf("登录成功应清空失败计数: %d", guard.successCalls)
	}
	if out.Token != "tok" || out.User == nil || out.User.Username != "admin" || out.User.Nickname != "超级管理员" {
		t.Fatalf("登录出参不符: %+v", out)
	}
	if len(out.User.Permissions) != 1 || out.User.Permissions[0] != "*" {
		t.Fatalf("权限未聚合: %+v", out.User.Permissions)
	}
	if len(out.User.Menus) != 1 || out.User.Menus[0].UIPath != "/system/user" {
		t.Fatalf("菜单未聚合: %+v", out.User.Menus)
	}
	if tokens.session == nil || tokens.session.UserID != 7 || tokens.session.ClientType != "admin" {
		t.Fatalf("令牌会话不符: %+v", tokens.session)
	}
	if sessionContext := tokens.session.ContextMap(); sessionContext["username"] != "admin" || sessionContext["nickname"] != "超级管理员" {
		t.Fatalf("令牌上下文不符: %+v", sessionContext)
	}
}

func TestLoginStopsWhenCaptchaFails(t *testing.T) {
	users := &fakeUserService{}
	captcha := &fakeCaptchaVerifier{err: errors.Sentinel(ErrCaptchaMismatch)}
	guard := &fakeLoginGuard{}

	uc := NewLoginUsecase(users, &fakeAuthService{}, &fakeTokenService{}, captcha, guard)
	_, err := uc.Execute(context.Background(), &LoginInput{
		Username:    "admin",
		Password:    "secret",
		CaptchaID:   "cid",
		CaptchaCode: "0000",
	}, "10.0.0.1")
	if !errors.Is(err, ErrCaptchaMismatch) {
		t.Fatalf("期望 ErrCaptchaMismatch，实际 %v", err)
	}
	if users.calls != 0 {
		t.Fatalf("验证码失败不应继续鉴权: %d", users.calls)
	}
	if guard.failureCalls != 0 {
		t.Fatalf("验证码失败不应计入账号失败: %d", guard.failureCalls)
	}
}

func TestLoginStopsWhenGuardLocked(t *testing.T) {
	users := &fakeUserService{}
	captcha := &fakeCaptchaVerifier{}
	guard := &fakeLoginGuard{
		beforeDecision: loginguard.Decision{Outcome: loginguard.OutcomeLocked, RetryAfter: 90 * time.Second},
	}

	uc := NewLoginUsecase(users, &fakeAuthService{}, &fakeTokenService{}, captcha, guard)
	_, err := uc.Execute(context.Background(), &LoginInput{
		Username: "admin",
		Password: "secret",
	}, "10.0.0.1")
	if !errors.Is(err, ErrLoginLocked) {
		t.Fatalf("期望 ErrLoginLocked，实际 %v", err)
	}
	if captcha.calls != 0 || users.calls != 0 {
		t.Fatalf("锁定时不应继续验证码或鉴权: captcha=%d users=%d", captcha.calls, users.calls)
	}
}

func TestLoginStopsWhenRateLimited(t *testing.T) {
	guard := &fakeLoginGuard{
		beforeDecision: loginguard.Decision{Outcome: loginguard.OutcomeRateLimited},
	}

	uc := NewLoginUsecase(&fakeUserService{}, &fakeAuthService{}, &fakeTokenService{}, &fakeCaptchaVerifier{}, guard)
	_, err := uc.Execute(context.Background(), &LoginInput{Username: "admin", Password: "secret"}, "10.0.0.1")
	if !errors.Is(err, ErrLoginRateLimited) {
		t.Fatalf("期望 ErrLoginRateLimited，实际 %v", err)
	}
}

func TestLoginFailsClosedWhenGuardUnavailable(t *testing.T) {
	guard := &fakeLoginGuard{beforeErr: errors.New("redis down")}

	uc := NewLoginUsecase(&fakeUserService{}, &fakeAuthService{}, &fakeTokenService{}, &fakeCaptchaVerifier{}, guard)
	_, err := uc.Execute(context.Background(), &LoginInput{Username: "admin", Password: "secret"}, "10.0.0.1")
	if !errors.Is(err, ErrLoginGuardUnavailable) {
		t.Fatalf("期望 ErrLoginGuardUnavailable，实际 %v", err)
	}
}

func TestLoginCountsInvalidCredential(t *testing.T) {
	users := &fakeUserService{err: errors.Sentinel(usercontract.ErrInvalidCredential)}
	guard := &fakeLoginGuard{}

	uc := NewLoginUsecase(users, &fakeAuthService{}, &fakeTokenService{}, &fakeCaptchaVerifier{}, guard)
	_, err := uc.Execute(context.Background(), &LoginInput{Username: "admin", Password: "wrong"}, "10.0.0.1")
	if !errors.Is(err, usercontract.ErrInvalidCredential) {
		t.Fatalf("期望 ErrInvalidCredential，实际 %v", err)
	}
	if guard.failureCalls != 1 || guard.successCalls != 0 {
		t.Fatalf("凭证失败应记账且不清零: failure=%d success=%d", guard.failureCalls, guard.successCalls)
	}
}

func TestLoginReturnsLockedWhenFailureTriggersLock(t *testing.T) {
	users := &fakeUserService{err: errors.Sentinel(usercontract.ErrInvalidCredential)}
	guard := &fakeLoginGuard{
		failureDecision: loginguard.Decision{Outcome: loginguard.OutcomeLocked, RetryAfter: time.Minute},
	}

	uc := NewLoginUsecase(users, &fakeAuthService{}, &fakeTokenService{}, &fakeCaptchaVerifier{}, guard)
	_, err := uc.Execute(context.Background(), &LoginInput{Username: "admin", Password: "wrong"}, "10.0.0.1")
	if !errors.Is(err, ErrLoginLocked) {
		t.Fatalf("期望 ErrLoginLocked，实际 %v", err)
	}
}

func TestLoginDisabledAccountDoesNotCountFailure(t *testing.T) {
	users := &fakeUserService{err: errors.Sentinel(usercontract.ErrUserDisabled)}
	guard := &fakeLoginGuard{}

	uc := NewLoginUsecase(users, &fakeAuthService{}, &fakeTokenService{}, &fakeCaptchaVerifier{}, guard)
	_, err := uc.Execute(context.Background(), &LoginInput{Username: "admin", Password: "secret"}, "10.0.0.1")
	if !errors.Is(err, usercontract.ErrUserDisabled) {
		t.Fatalf("期望 ErrUserDisabled，实际 %v", err)
	}
	if guard.failureCalls != 0 || guard.successCalls != 1 {
		t.Fatalf("密码正确但禁用应清零而非记账: failure=%d success=%d", guard.failureCalls, guard.successCalls)
	}
}

func TestLoginInternalErrorDoesNotCountFailure(t *testing.T) {
	users := &fakeUserService{err: errors.New("db down")}
	guard := &fakeLoginGuard{}

	uc := NewLoginUsecase(users, &fakeAuthService{}, &fakeTokenService{}, &fakeCaptchaVerifier{}, guard)
	if _, err := uc.Execute(context.Background(), &LoginInput{Username: "admin", Password: "secret"}, "10.0.0.1"); err == nil {
		t.Fatal("内部错误应向上传播")
	}
	if guard.failureCalls != 0 || guard.successCalls != 0 {
		t.Fatalf("内部错误不应影响失败计数: failure=%d success=%d", guard.failureCalls, guard.successCalls)
	}
}

func TestLoginIgnoresGuardResetFailure(t *testing.T) {
	users := &fakeUserService{out: &useroutput.AuthenticateOutput{User: &userentity.User{
		ID:       7,
		Username: "admin",
		Nickname: "超级管理员",
	}}}
	guard := &fakeLoginGuard{successErr: errors.New("redis down")}

	uc := NewLoginUsecase(users, &fakeAuthService{}, &fakeTokenService{value: "tok"}, &fakeCaptchaVerifier{}, guard)
	out, err := uc.Execute(context.Background(), &LoginInput{Username: "admin", Password: "secret"}, "10.0.0.1")
	if err != nil {
		t.Fatalf("清零失败不应阻断登录: %v", err)
	}
	if out.Token != "tok" {
		t.Fatalf("登录出参不符: %+v", out)
	}
}

// TestLoginKeepsFailureBudgetWhenIssueFails 校验清零发生在整条链路成功之后：
// 聚合权限或签发令牌失败时，失败预算必须保留，下一次尝试仍受既有约束。
func TestLoginKeepsFailureBudgetWhenIssueFails(t *testing.T) {
	users := &fakeUserService{out: &useroutput.AuthenticateOutput{User: &userentity.User{
		ID:       7,
		Username: "admin",
		Nickname: "超级管理员",
	}}}
	guard := &fakeLoginGuard{}

	uc := NewLoginUsecase(users, &fakeAuthService{}, &fakeTokenService{err: errors.New("issue down")}, &fakeCaptchaVerifier{}, guard)
	if _, err := uc.Execute(context.Background(), &LoginInput{Username: "admin", Password: "secret"}, "10.0.0.1"); err == nil {
		t.Fatal("签发令牌失败应向上传播")
	}
	if guard.successCalls != 0 {
		t.Fatalf("签发令牌失败时不应清空失败预算: %d", guard.successCalls)
	}
}

// TestLoginRejectsOversizedInput 校验超长入参在入口被拒：不消耗尝试配额、不触达验证码与领域层
func TestLoginRejectsOversizedInput(t *testing.T) {
	cases := []struct {
		name     string
		username string
		password string
	}{
		{name: "用户名超过 64 字节", username: strings.Repeat("a", 65), password: "secret"},
		{name: "密码超过 72 字节", username: "admin", password: strings.Repeat("p", 73)},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			users := &fakeUserService{}
			captcha := &fakeCaptchaVerifier{}
			guard := &fakeLoginGuard{}

			uc := NewLoginUsecase(users, &fakeAuthService{}, &fakeTokenService{}, captcha, guard)
			_, err := uc.Execute(context.Background(), &LoginInput{Username: tc.username, Password: tc.password}, "10.0.0.1")
			if !errors.Is(err, usercontract.ErrInvalidInput) {
				t.Fatalf("期望 ErrInvalidInput，实际 %v", err)
			}
			if guard.beforeCalls != 0 || captcha.calls != 0 || users.calls != 0 {
				t.Fatalf("超长入参不应触达后续步骤: guard=%d captcha=%d users=%d",
					guard.beforeCalls, captcha.calls, users.calls)
			}
		})
	}
}

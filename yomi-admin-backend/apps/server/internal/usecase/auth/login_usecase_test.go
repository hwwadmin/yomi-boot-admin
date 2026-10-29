package auth

import (
	"context"
	"testing"

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

// fakeTokenService 是令牌服务测试桩
type fakeTokenService struct {
	token.TokenService
	value   string
	session *token.TokenSession
}

func (f *fakeTokenService) Issue(_ context.Context, session *token.TokenSession) (string, error) {
	f.session = session
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

	uc := NewLoginUsecase(users, auth, tokens, captcha)
	out, err := uc.Execute(context.Background(), &LoginInput{
		Username:    "admin",
		Password:    "secret",
		CaptchaID:   "cid",
		CaptchaCode: "1234",
	})
	if err != nil {
		t.Fatalf("登录失败: %v", err)
	}

	if captcha.calls != 1 || captcha.id != "cid" || captcha.code != "1234" {
		t.Fatalf("验证码校验入参不符: calls=%d id=%q code=%q", captcha.calls, captcha.id, captcha.code)
	}
	if users.calls != 1 || users.got == nil || users.got.Username != "admin" || users.got.Password != "secret" {
		t.Fatalf("领域鉴权入参不符: calls=%d got=%+v", users.calls, users.got)
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

	uc := NewLoginUsecase(users, &fakeAuthService{}, &fakeTokenService{}, captcha)
	_, err := uc.Execute(context.Background(), &LoginInput{
		Username:    "admin",
		Password:    "secret",
		CaptchaID:   "cid",
		CaptchaCode: "0000",
	})
	if !errors.Is(err, ErrCaptchaMismatch) {
		t.Fatalf("期望 ErrCaptchaMismatch，实际 %v", err)
	}
	if users.calls != 0 {
		t.Fatalf("验证码失败不应继续鉴权: %d", users.calls)
	}
}

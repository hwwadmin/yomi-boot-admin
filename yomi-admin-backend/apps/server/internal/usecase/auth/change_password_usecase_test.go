package auth

import (
	"context"
	"testing"
	"yomi-admin-backend/apps/server/internal/support"
	"yomi-admin-backend/domains/user/service_contract"

	userinput "yomi-admin-backend/domains/user/pkg/shared/input"

	"github.com/hwwadmin/yomi-boot/component/common/token"
	"github.com/hwwadmin/yomi-boot/component/core/errors"
)

// fakeChangePasswordUser 只实现编排会调用到的改密契约方法
type fakeChangePasswordUser struct {
	service_contract.UserService
	got *userinput.ChangePasswordInput
	err error
}

func (f *fakeChangePasswordUser) ChangePassword(_ context.Context, in *userinput.ChangePasswordInput) error {
	f.got = in
	return f.err
}

// fakeRevokeToken 只实现令牌撤销；可配置恒定失败
type fakeRevokeToken struct {
	token.TokenService
	revoked []int64
	err     error
}

func (f *fakeRevokeToken) RevokeByUserID(_ context.Context, userID int64) error {
	if f.err != nil {
		return f.err
	}
	f.revoked = append(f.revoked, userID)
	return nil
}

func TestChangePasswordUsesOperatorAndRevokesTokens(t *testing.T) {
	users := &fakeChangePasswordUser{}
	tokens := &fakeRevokeToken{}
	uc := NewChangePasswordUsecase(users, tokens)

	err := uc.Execute(context.Background(), &ChangePasswordInput{OldPassword: "old", NewPassword: "new"}, 7)
	if err != nil {
		t.Fatalf("修改密码失败: %v", err)
	}
	if users.got == nil || users.got.ID != 7 || users.got.OldPassword != "old" || users.got.NewPassword != "new" {
		t.Fatalf("改密入参未透传操作者: %+v", users.got)
	}
	if len(tokens.revoked) != 1 || tokens.revoked[0] != 7 {
		t.Fatalf("改密成功应撤销该用户全部令牌: %v", tokens.revoked)
	}
}

func TestChangePasswordDomainErrorSkipsRevoke(t *testing.T) {
	users := &fakeChangePasswordUser{err: errors.Sentinel(service_contract.ErrOldPasswordMismatch)}
	tokens := &fakeRevokeToken{}
	uc := NewChangePasswordUsecase(users, tokens)

	err := uc.Execute(context.Background(), &ChangePasswordInput{OldPassword: "old", NewPassword: "new"}, 7)
	if !errors.Is(err, service_contract.ErrOldPasswordMismatch) {
		t.Fatalf("期望 ErrOldPasswordMismatch，实际 %v", err)
	}
	if len(tokens.revoked) != 0 {
		t.Fatalf("改密失败不应撤销令牌: %v", tokens.revoked)
	}
}

func TestChangePasswordRevokeFailureKeepsMessage(t *testing.T) {
	users := &fakeChangePasswordUser{}
	tokens := &fakeRevokeToken{err: errors.New("redis down")}
	uc := NewChangePasswordUsecase(users, tokens)

	err := uc.Execute(context.Background(), &ChangePasswordInput{OldPassword: "old", NewPassword: "new"}, 7)
	// 调用方在幂等事务内执行：撤销失败会回滚密码更新，提示按"改密未完成"处理
	if !errors.Is(err, support.ErrPasswordChangeTokenRevokeFailed) {
		t.Fatalf("期望 ErrPasswordChangeTokenRevokeFailed，实际 %v", err)
	}
}

func TestChangePasswordRejectsMissingOperator(t *testing.T) {
	users := &fakeChangePasswordUser{}
	tokens := &fakeRevokeToken{}
	uc := NewChangePasswordUsecase(users, tokens)

	if err := uc.Execute(context.Background(), &ChangePasswordInput{OldPassword: "old", NewPassword: "new"}, 0); err == nil {
		t.Fatal("缺少操作者应报错")
	}
	if users.got != nil || len(tokens.revoked) != 0 {
		t.Fatalf("缺少操作者不应触达后续步骤: got=%v revoked=%v", users.got, tokens.revoked)
	}
}

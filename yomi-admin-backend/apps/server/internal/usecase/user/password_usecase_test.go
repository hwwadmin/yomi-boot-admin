package user

import (
	"context"
	"testing"
	"yomi-admin-backend/apps/server/internal/support"
	"yomi-admin-backend/apps/server/internal/support/superadmin"
	"yomi-admin-backend/domains/user/service_contract"

	userinput "yomi-admin-backend/domains/user/pkg/shared/input"

	"github.com/hwwadmin/yomi-boot/component/core/errors"
	"github.com/hwwadmin/yomi-boot/component/core/shared/dto"
)

// fakeResetPasswordUser 只实现编排会调用到的重置密码契约方法
type fakeResetPasswordUser struct {
	service_contract.UserService
	got *userinput.ResetPasswordInput
	err error
}

func (f *fakeResetPasswordUser) ResetPassword(_ context.Context, in *userinput.ResetPasswordInput) error {
	f.got = in
	return f.err
}

// TestResetUserPasswordSuccess 校验重置成功、透传入参并撤销令牌
func TestResetUserPasswordSuccess(t *testing.T) {
	users := &fakeResetPasswordUser{}
	guard := &fakeGuard{}
	tokens := &fakeTokenService{}
	uc := NewResetUserPasswordUsecase(users, guard, tokens)

	if err := uc.Execute(context.Background(), &ResetUserPasswordInput{
		ID:          dto.JSONInt64(8),
		NewPassword: "newpass1",
		OperatorID:  1,
	}); err != nil {
		t.Fatalf("重置密码失败: %v", err)
	}
	if guard.resetCalled != 1 || guard.resetOperatorID != 1 {
		t.Fatalf("超管校验未执行或入参错误: called=%d operator=%d", guard.resetCalled, guard.resetOperatorID)
	}
	if users.got == nil || users.got.ID != 8 || users.got.NewPassword != "newpass1" || users.got.OperatorID != 1 {
		t.Fatalf("重置入参未透传: %+v", users.got)
	}
	if len(tokens.revoked) != 1 || tokens.revoked[0] != 8 {
		t.Fatalf("重置成功应撤销目标用户全部令牌: %v", tokens.revoked)
	}
}

// TestResetUserPasswordCannotResetSelf 校验重置自己被拒，且不触达任何后续步骤
func TestResetUserPasswordCannotResetSelf(t *testing.T) {
	users := &fakeResetPasswordUser{}
	guard := &fakeGuard{}
	tokens := &fakeTokenService{}
	uc := NewResetUserPasswordUsecase(users, guard, tokens)

	err := uc.Execute(context.Background(), &ResetUserPasswordInput{
		ID:          dto.JSONInt64(7),
		NewPassword: "newpass1",
		OperatorID:  7,
	})
	if !errors.Is(err, support.ErrCannotOperateSelf) {
		t.Fatalf("重置自己期望 ErrCannotOperateSelf，实际 %v", err)
	}
	if guard.resetCalled != 0 || users.got != nil || len(tokens.revoked) != 0 {
		t.Fatalf("自重置不应触达后续步骤: guard=%d got=%v revoked=%v", guard.resetCalled, users.got, tokens.revoked)
	}
}

// TestResetUserPasswordDeniedBySuperAdminGuard 校验非超管被拒，且不写库、不撤令牌
func TestResetUserPasswordDeniedBySuperAdminGuard(t *testing.T) {
	users := &fakeResetPasswordUser{}
	guard := &fakeGuard{resetErr: superadmin.ErrResetPasswordDenied}
	tokens := &fakeTokenService{}
	uc := NewResetUserPasswordUsecase(users, guard, tokens)

	err := uc.Execute(context.Background(), &ResetUserPasswordInput{
		ID:          dto.JSONInt64(8),
		NewPassword: "newpass1",
		OperatorID:  9,
	})
	if !errors.Is(err, superadmin.ErrResetPasswordDenied) {
		t.Fatalf("期望 ErrResetPasswordDenied，实际 %v", err)
	}
	if users.got != nil || len(tokens.revoked) != 0 {
		t.Fatalf("被拒绝后不应写库或撤令牌: got=%v revoked=%v", users.got, tokens.revoked)
	}
}

// TestResetUserPasswordRejectsInvalidInput 校验非法 ID 直接拒绝，不触发超管校验
func TestResetUserPasswordRejectsInvalidInput(t *testing.T) {
	users := &fakeResetPasswordUser{}
	guard := &fakeGuard{}
	tokens := &fakeTokenService{}
	uc := NewResetUserPasswordUsecase(users, guard, tokens)

	if err := uc.Execute(context.Background(), &ResetUserPasswordInput{ID: dto.JSONInt64(0), NewPassword: "newpass1", OperatorID: 1}); !errors.Is(err, service_contract.ErrInvalidInput) {
		t.Fatalf("非法 ID 期望 ErrInvalidInput，实际 %v", err)
	}
	if guard.resetCalled != 0 {
		t.Fatalf("非法 ID 不应触发超管校验: %d", guard.resetCalled)
	}
}

// TestResetUserPasswordRevokedFailsAfterRetries 校验重置成功但撤销失败时的错误口径
func TestResetUserPasswordRevokedFailsAfterRetries(t *testing.T) {
	users := &fakeResetPasswordUser{}
	guard := &fakeGuard{}
	tokens := &fakeTokenService{failures: support.TokenRevokeAttempts}
	uc := NewResetUserPasswordUsecase(users, guard, tokens)

	err := uc.Execute(context.Background(), &ResetUserPasswordInput{
		ID:          dto.JSONInt64(8),
		NewPassword: "newpass1",
		OperatorID:  1,
	})
	if !errors.Is(err, support.ErrTokenRevokeFailed) {
		t.Fatalf("期望 ErrTokenRevokeFailed，实际 %v", err)
	}
	if tokens.attempts != support.TokenRevokeAttempts {
		t.Fatalf("撤销尝试次数 = %d, want %d", tokens.attempts, support.TokenRevokeAttempts)
	}
}

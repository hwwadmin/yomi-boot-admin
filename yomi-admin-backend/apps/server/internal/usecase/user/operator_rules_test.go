package user

import (
	"context"
	"testing"
	"time"
	"yomi-admin-backend/apps/server/internal/support"
	"yomi-admin-backend/domains/user/pkg/shared/enums"
	"yomi-admin-backend/domains/user/service_contract"

	userinput "yomi-admin-backend/domains/user/pkg/shared/input"

	"gitee.com/hq_project/yomi-boot/component/common/token"
	"gitee.com/hq_project/yomi-boot/component/core/errors"
)

// fakeUserService 只实现编排会调用到的用户契约方法
type fakeUserService struct {
	service_contract.UserService
	updated *userinput.UpdateUserInput
	deleted *userinput.DeleteUserInput
}

func (f *fakeUserService) Update(_ context.Context, in *userinput.UpdateUserInput) error {
	f.updated = in
	return nil
}

func (f *fakeUserService) Delete(_ context.Context, in *userinput.DeleteUserInput) error {
	f.deleted = in
	return nil
}

// fakeExecutor 在当前协程内直接执行事务闭包
type fakeExecutor struct{}

func (fakeExecutor) Run(ctx context.Context, fn func(ctx context.Context) error) error {
	return fn(ctx)
}

func (fakeExecutor) InTx(context.Context) bool { return false }

func (fakeExecutor) RunAutonomous(ctx context.Context, _ time.Duration, fn func(ctx context.Context) error) error {
	return fn(ctx)
}

// fakeGuard 记录最后一名超管策略的调用次数
type fakeGuard struct {
	called int
}

func (f *fakeGuard) EnsureNotLastSuper(context.Context, int64) error {
	f.called++
	return nil
}

// fakeTokenService 只实现令牌撤销
type fakeTokenService struct {
	token.TokenService
	revoked []int64
}

func (f *fakeTokenService) RevokeByUserID(_ context.Context, userID int64) error {
	f.revoked = append(f.revoked, userID)
	return nil
}

func TestUpdateUserCannotDisableSelf(t *testing.T) {
	users := &fakeUserService{}
	guard := &fakeGuard{}
	tokens := &fakeTokenService{}
	uc := NewUpdateUserUsecase(users, guard, fakeExecutor{}, tokens)

	err := uc.Execute(context.Background(), 7, &UpdateUserInput{
		Status:     enums.UserStatusDisabled,
		OperatorID: 7,
	})
	if !errors.Is(err, support.ErrCannotOperateSelf) {
		t.Fatalf("禁用自己期望 ErrCannotOperateSelf，实际 %v", err)
	}
	if users.updated != nil || guard.called != 0 || len(tokens.revoked) != 0 {
		t.Fatalf("自操作不应触达后续步骤: updated=%v guard=%d revoked=%v", users.updated, guard.called, tokens.revoked)
	}
}

func TestUpdateOtherUserDisableRevokesToken(t *testing.T) {
	users := &fakeUserService{}
	guard := &fakeGuard{}
	tokens := &fakeTokenService{}
	uc := NewUpdateUserUsecase(users, guard, fakeExecutor{}, tokens)

	if err := uc.Execute(context.Background(), 8, &UpdateUserInput{
		Nickname:   "新昵称",
		Status:     enums.UserStatusDisabled,
		OperatorID: 1,
	}); err != nil {
		t.Fatalf("禁用他人失败: %v", err)
	}
	if guard.called != 1 {
		t.Fatalf("最后超管策略调用次数 = %d, want 1", guard.called)
	}
	if users.updated == nil || users.updated.ID != 8 {
		t.Fatalf("更新未透传: %+v", users.updated)
	}
	if len(tokens.revoked) != 1 || tokens.revoked[0] != 8 {
		t.Fatalf("令牌未撤销: %v", tokens.revoked)
	}
}

func TestUpdateSelfEnableAllowed(t *testing.T) {
	users := &fakeUserService{}
	guard := &fakeGuard{}
	tokens := &fakeTokenService{}
	uc := NewUpdateUserUsecase(users, guard, fakeExecutor{}, tokens)

	if err := uc.Execute(context.Background(), 7, &UpdateUserInput{
		Status:     enums.UserStatusEnabled,
		OperatorID: 7,
	}); err != nil {
		t.Fatalf("启用自己应放行: %v", err)
	}
	if guard.called != 0 || len(tokens.revoked) != 0 {
		t.Fatalf("启用不应触发策略或撤令牌: guard=%d revoked=%v", guard.called, tokens.revoked)
	}
}

func TestDeleteUserCannotDeleteSelf(t *testing.T) {
	users := &fakeUserService{}
	guard := &fakeGuard{}
	tokens := &fakeTokenService{}
	uc := NewDeleteUserUsecase(users, guard, fakeExecutor{}, tokens)

	if err := uc.Execute(context.Background(), 7, 7); !errors.Is(err, support.ErrCannotOperateSelf) {
		t.Fatalf("删除自己期望 ErrCannotOperateSelf，实际 %v", err)
	}
	if users.deleted != nil || guard.called != 0 || len(tokens.revoked) != 0 {
		t.Fatalf("自删除不应触达后续步骤: deleted=%v guard=%d revoked=%v", users.deleted, guard.called, tokens.revoked)
	}
}

func TestDeleteOtherUserRevokesToken(t *testing.T) {
	users := &fakeUserService{}
	guard := &fakeGuard{}
	tokens := &fakeTokenService{}
	uc := NewDeleteUserUsecase(users, guard, fakeExecutor{}, tokens)

	if err := uc.Execute(context.Background(), 8, 1); err != nil {
		t.Fatalf("删除他人失败: %v", err)
	}
	if guard.called != 1 {
		t.Fatalf("最后超管策略调用次数 = %d, want 1", guard.called)
	}
	if users.deleted == nil || users.deleted.ID != 8 {
		t.Fatalf("删除未透传: %+v", users.deleted)
	}
	if len(tokens.revoked) != 1 || tokens.revoked[0] != 8 {
		t.Fatalf("令牌未撤销: %v", tokens.revoked)
	}
}

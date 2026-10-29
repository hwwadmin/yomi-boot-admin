package user

import (
	"context"
	"fmt"
	"testing"
	"time"
	"yomi-admin-backend/apps/server/internal/support"
	"yomi-admin-backend/apps/server/internal/support/superadmin"
	rbaccontract "yomi-admin-backend/domains/rbac/service_contract"
	"yomi-admin-backend/domains/user/pkg/shared/enums"
	"yomi-admin-backend/domains/user/service_contract"

	userinput "yomi-admin-backend/domains/user/pkg/shared/input"

	"gitee.com/hq_project/yomi-boot/component/common/token"
	"gitee.com/hq_project/yomi-boot/component/core/errors"
	"gitee.com/hq_project/yomi-boot/component/core/shared/dto"
)

// fakeUserService 只实现编排会调用到的用户契约方法
type fakeUserService struct {
	service_contract.UserService
	updated   *userinput.UpdateUserInput
	deleted   *userinput.DeleteUserInput
	deleteErr error
}

func (f *fakeUserService) Update(_ context.Context, in *userinput.UpdateUserInput) error {
	f.updated = in
	return nil
}

func (f *fakeUserService) Delete(_ context.Context, in *userinput.DeleteUserInput) error {
	f.deleted = in
	return f.deleteErr
}

// fakeRbacAdmin 只实现删除用户角色绑定，其余方法由嵌入接口兜底
type fakeRbacAdmin struct {
	rbaccontract.RbacAdminService
	cleanedUserID int64
}

func (f *fakeRbacAdmin) DeleteUserRoles(_ context.Context, userID int64) error {
	f.cleanedUserID = userID
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

// fakeGuard 记录超管策略与锁的调用，并可注入拒绝
type fakeGuard struct {
	called     int
	locks      int
	operatorID int64
	targetID   int64
	err        error
}

func (f *fakeGuard) WithLock(ctx context.Context, fn func(context.Context) error) error {
	f.locks++
	return fn(ctx)
}

func (f *fakeGuard) EnsureCanManageUser(_ context.Context, operatorID, targetUserID int64) error {
	f.called++
	f.operatorID = operatorID
	f.targetID = targetUserID
	return f.err
}

// fakeTokenService 只实现令牌撤销；可配置前 N 次失败
type fakeTokenService struct {
	token.TokenService
	failures int
	attempts int
	revoked  []int64
}

func (f *fakeTokenService) RevokeByUserID(_ context.Context, userID int64) error {
	f.attempts++
	if f.attempts <= f.failures {
		return fmt.Errorf("redis unavailable")
	}
	f.revoked = append(f.revoked, userID)
	return nil
}

// TestUpdateUserRejectsInvalidStatus 校验状态缺省或非法时直接报参数错误，不进入超管锁与后续流程
func TestUpdateUserRejectsInvalidStatus(t *testing.T) {
	for _, status := range []enums.UserStatus{"", "UNKNOWN"} {
		users := &fakeUserService{}
		guard := &fakeGuard{}
		tokens := &fakeTokenService{}
		uc := NewUpdateUserUsecase(users, guard, fakeExecutor{}, tokens)

		err := uc.Execute(context.Background(), &UpdateUserInput{ID: dto.JSONInt64(8), Status: status, OperatorID: 1})
		if !errors.Is(err, service_contract.ErrInvalidInput) {
			t.Fatalf("状态 %q 期望 ErrInvalidInput，实际 %v", status, err)
		}
		if guard.locks != 0 || guard.called != 0 || users.updated != nil || len(tokens.revoked) != 0 {
			t.Fatalf("状态 %q 不应触达锁/策略/写库/撤令牌: locks=%d called=%d updated=%v revoked=%v",
				status, guard.locks, guard.called, users.updated, tokens.revoked)
		}
	}
}

func TestUpdateUserCannotDisableSelf(t *testing.T) {
	users := &fakeUserService{}
	guard := &fakeGuard{}
	tokens := &fakeTokenService{}
	uc := NewUpdateUserUsecase(users, guard, fakeExecutor{}, tokens)

	err := uc.Execute(context.Background(), &UpdateUserInput{
		ID:         dto.JSONInt64(7),
		Status:     enums.UserStatusDisabled,
		OperatorID: 7,
	})
	if !errors.Is(err, support.ErrCannotOperateSelf) {
		t.Fatalf("禁用自己期望 ErrCannotOperateSelf，实际 %v", err)
	}
	if users.updated != nil || guard.called != 0 || guard.locks != 0 || len(tokens.revoked) != 0 {
		t.Fatalf("自操作不应触达后续步骤: updated=%v guard=%d locks=%d revoked=%v", users.updated, guard.called, guard.locks, tokens.revoked)
	}
}

func TestUpdateOtherUserDisableRevokesToken(t *testing.T) {
	users := &fakeUserService{}
	guard := &fakeGuard{}
	tokens := &fakeTokenService{}
	uc := NewUpdateUserUsecase(users, guard, fakeExecutor{}, tokens)

	if err := uc.Execute(context.Background(), &UpdateUserInput{
		ID:         dto.JSONInt64(8),
		Nickname:   "新昵称",
		Status:     enums.UserStatusDisabled,
		OperatorID: 1,
	}); err != nil {
		t.Fatalf("禁用他人失败: %v", err)
	}
	if guard.called != 1 {
		t.Fatalf("最后超管策略调用次数 = %d, want 1", guard.called)
	}
	if guard.locks != 1 {
		t.Fatalf("超管锁调用次数 = %d, want 1", guard.locks)
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

	if err := uc.Execute(context.Background(), &UpdateUserInput{
		ID:         dto.JSONInt64(7),
		Status:     enums.UserStatusEnabled,
		OperatorID: 7,
	}); err != nil {
		t.Fatalf("启用自己应放行: %v", err)
	}
	if guard.called != 0 || guard.locks != 0 || len(tokens.revoked) != 0 {
		t.Fatalf("启用不应触发策略或撤令牌: guard=%d locks=%d revoked=%v", guard.called, guard.locks, tokens.revoked)
	}
}

func TestDeleteUserCannotDeleteSelf(t *testing.T) {
	users := &fakeUserService{}
	rbacAdmin := &fakeRbacAdmin{}
	guard := &fakeGuard{}
	tokens := &fakeTokenService{}
	uc := NewDeleteUserUsecase(users, rbacAdmin, guard, fakeExecutor{}, tokens)

	if err := uc.Execute(context.Background(), &DeleteUserInput{ID: dto.JSONInt64(7), OperatorID: 7}); !errors.Is(err, support.ErrCannotOperateSelf) {
		t.Fatalf("删除自己期望 ErrCannotOperateSelf，实际 %v", err)
	}
	if users.deleted != nil || guard.called != 0 || guard.locks != 0 || len(tokens.revoked) != 0 {
		t.Fatalf("自删除不应触达后续步骤: deleted=%v guard=%d locks=%d revoked=%v", users.deleted, guard.called, guard.locks, tokens.revoked)
	}
}

func TestDeleteOtherUserRevokesToken(t *testing.T) {
	users := &fakeUserService{}
	rbacAdmin := &fakeRbacAdmin{}
	guard := &fakeGuard{}
	tokens := &fakeTokenService{}
	uc := NewDeleteUserUsecase(users, rbacAdmin, guard, fakeExecutor{}, tokens)

	if err := uc.Execute(context.Background(), &DeleteUserInput{ID: dto.JSONInt64(8), OperatorID: 1}); err != nil {
		t.Fatalf("删除他人失败: %v", err)
	}
	if guard.called != 1 {
		t.Fatalf("最后超管策略调用次数 = %d, want 1", guard.called)
	}
	if guard.locks != 1 {
		t.Fatalf("超管锁调用次数 = %d, want 1", guard.locks)
	}
	if users.deleted == nil || users.deleted.ID != 8 {
		t.Fatalf("删除未透传: %+v", users.deleted)
	}
	if rbacAdmin.cleanedUserID != 8 {
		t.Fatalf("用户角色未清理: %d", rbacAdmin.cleanedUserID)
	}
	if len(tokens.revoked) != 1 || tokens.revoked[0] != 8 {
		t.Fatalf("令牌未撤销: %v", tokens.revoked)
	}
}

func TestUpdateUserTokenRevokeRetries(t *testing.T) {
	users := &fakeUserService{}
	guard := &fakeGuard{}
	tokens := &fakeTokenService{failures: 1}
	uc := NewUpdateUserUsecase(users, guard, fakeExecutor{}, tokens)

	if err := uc.Execute(context.Background(), &UpdateUserInput{
		ID:         dto.JSONInt64(8),
		Status:     enums.UserStatusDisabled,
		OperatorID: 1,
	}); err != nil {
		t.Fatalf("重试后应成功: %v", err)
	}
	if tokens.attempts != 2 {
		t.Fatalf("撤销尝试次数 = %d, want 2", tokens.attempts)
	}
	if len(tokens.revoked) != 1 || tokens.revoked[0] != 8 {
		t.Fatalf("重试后未撤销令牌: %v", tokens.revoked)
	}
}

func TestUpdateUserTokenRevokeFailsAfterRetries(t *testing.T) {
	users := &fakeUserService{}
	guard := &fakeGuard{}
	tokens := &fakeTokenService{failures: tokenRevokeAttempts}
	uc := NewUpdateUserUsecase(users, guard, fakeExecutor{}, tokens)

	err := uc.Execute(context.Background(), &UpdateUserInput{
		ID:         dto.JSONInt64(8),
		Status:     enums.UserStatusDisabled,
		OperatorID: 1,
	})
	if !errors.Is(err, support.ErrTokenRevokeFailed) {
		t.Fatalf("期望 ErrTokenRevokeFailed，实际 %v", err)
	}
	if tokens.attempts != tokenRevokeAttempts {
		t.Fatalf("撤销尝试次数 = %d, want %d", tokens.attempts, tokenRevokeAttempts)
	}
	if len(tokens.revoked) != 0 {
		t.Fatalf("失败时不应记录撤销成功: %v", tokens.revoked)
	}
}

func TestDeleteUserRevokesTokenWhenAlreadyDeleted(t *testing.T) {
	users := &fakeUserService{deleteErr: errors.Sentinel(service_contract.ErrUserNotFound)}
	rbacAdmin := &fakeRbacAdmin{}
	guard := &fakeGuard{}
	tokens := &fakeTokenService{}
	uc := NewDeleteUserUsecase(users, rbacAdmin, guard, fakeExecutor{}, tokens)

	err := uc.Execute(context.Background(), &DeleteUserInput{ID: dto.JSONInt64(8), OperatorID: 1})
	if !errors.Is(err, service_contract.ErrUserNotFound) {
		t.Fatalf("期望 ErrUserNotFound，实际 %v", err)
	}
	if tokens.attempts != 1 || len(tokens.revoked) != 1 || tokens.revoked[0] != 8 {
		t.Fatalf("已删除用户也应补撤令牌: attempts=%d revoked=%v", tokens.attempts, tokens.revoked)
	}
	if guard.locks != 1 {
		t.Fatalf("超管锁调用次数 = %d, want 1", guard.locks)
	}
}

func TestDeleteUserTokenRevokeFailsAfterRetries(t *testing.T) {
	users := &fakeUserService{}
	rbacAdmin := &fakeRbacAdmin{}
	guard := &fakeGuard{}
	tokens := &fakeTokenService{failures: tokenRevokeAttempts}
	uc := NewDeleteUserUsecase(users, rbacAdmin, guard, fakeExecutor{}, tokens)

	err := uc.Execute(context.Background(), &DeleteUserInput{ID: dto.JSONInt64(8), OperatorID: 1})
	if !errors.Is(err, support.ErrTokenRevokeFailed) {
		t.Fatalf("期望 ErrTokenRevokeFailed，实际 %v", err)
	}
	if tokens.attempts != tokenRevokeAttempts {
		t.Fatalf("撤销尝试次数 = %d, want %d", tokens.attempts, tokenRevokeAttempts)
	}
}

// TestUpdateUserDeniedBySuperAdminGuard 校验普通管理员停用超管时被拒，且不产生任何副作用
func TestUpdateUserDeniedBySuperAdminGuard(t *testing.T) {
	users := &fakeUserService{}
	guard := &fakeGuard{err: superadmin.ErrManageSuperDenied}
	tokens := &fakeTokenService{}
	uc := NewUpdateUserUsecase(users, guard, fakeExecutor{}, tokens)

	err := uc.Execute(context.Background(), &UpdateUserInput{
		ID:         dto.JSONInt64(8),
		Status:     enums.UserStatusDisabled,
		OperatorID: 9,
	})
	if !errors.Is(err, superadmin.ErrManageSuperDenied) {
		t.Fatalf("期望 ErrManageSuperDenied，实际 %v", err)
	}
	if guard.operatorID != 9 || guard.targetID != 8 {
		t.Fatalf("策略入参错误: operator=%d target=%d", guard.operatorID, guard.targetID)
	}
	if users.updated != nil || len(tokens.revoked) != 0 {
		t.Fatalf("被拒绝后不应写库或撤令牌: updated=%v revoked=%v", users.updated, tokens.revoked)
	}
}

// TestDeleteUserDeniedBySuperAdminGuard 校验普通管理员删除超管时被拒，且不写库、不撤令牌
func TestDeleteUserDeniedBySuperAdminGuard(t *testing.T) {
	users := &fakeUserService{}
	rbacAdmin := &fakeRbacAdmin{}
	guard := &fakeGuard{err: superadmin.ErrManageSuperDenied}
	tokens := &fakeTokenService{}
	uc := NewDeleteUserUsecase(users, rbacAdmin, guard, fakeExecutor{}, tokens)

	err := uc.Execute(context.Background(), &DeleteUserInput{ID: dto.JSONInt64(8), OperatorID: 9})
	if !errors.Is(err, superadmin.ErrManageSuperDenied) {
		t.Fatalf("期望 ErrManageSuperDenied，实际 %v", err)
	}
	if guard.operatorID != 9 || guard.targetID != 8 {
		t.Fatalf("策略入参错误: operator=%d target=%d", guard.operatorID, guard.targetID)
	}
	if users.deleted != nil || rbacAdmin.cleanedUserID != 0 || len(tokens.revoked) != 0 {
		t.Fatalf("被拒绝后不应删除或清理: deleted=%v cleaned=%d revoked=%v",
			users.deleted, rbacAdmin.cleanedUserID, tokens.revoked)
	}
}

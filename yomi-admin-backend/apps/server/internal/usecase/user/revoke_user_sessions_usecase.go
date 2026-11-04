package user

import (
	"context"
	"yomi-admin-backend/apps/server/internal/support"

	"github.com/hwwadmin/yomi-boot/component/core/errors"
	"github.com/hwwadmin/yomi-boot/component/core/logger"
	"github.com/hwwadmin/yomi-boot/component/core/shared/dto"
	sharederr "github.com/hwwadmin/yomi-boot/component/core/shared/error"
)

// RevokeUserSessionsInput HTTP 强制下线用户全部会话入参
type RevokeUserSessionsInput struct {
	UserID     dto.JSONInt64 `json:"userId"`
	OperatorID int64         `json:"-"`
}

// RevokeUserSessionsUsecase 强制下线指定用户的全部会话
type RevokeUserSessionsUsecase struct {
	tokens SessionTokens
	guard  kickGuard
}

// NewRevokeUserSessionsUsecase 创建按用户全部下线用例
func NewRevokeUserSessionsUsecase(tokens SessionTokens, guard kickGuard) *RevokeUserSessionsUsecase {
	return &RevokeUserSessionsUsecase{tokens: tokens, guard: guard}
}

// Execute 撤销目标用户的全部令牌；目标是操作者自己时拒绝，避免把自己踢出管理页
func (uc *RevokeUserSessionsUsecase) Execute(ctx context.Context, in *RevokeUserSessionsInput) error {
	userID := in.UserID.Int64()
	if userID <= 0 {
		return errors.Sentinel(sharederr.ErrInvalidInput)
	}
	if userID == in.OperatorID {
		return support.ErrCannotOperateSelf
	}
	if err := uc.guard.EnsureCanKickUser(ctx, in.OperatorID, userID); err != nil {
		return err
	}

	if err := support.RevokeUserSessions(ctx, uc.tokens, userID); err != nil {
		return errors.WrapBiz(err, support.ErrSessionRevokeFailed.Code, support.ErrSessionRevokeFailed.Message)
	}
	logger.Info(ctx, "强制下线用户全部会话", "operatorID", in.OperatorID, "userID", userID)
	return nil
}

package user

import (
	"context"
	"strings"
	"yomi-admin-backend/apps/server/internal/support"

	"github.com/hwwadmin/yomi-boot/component/core/errors"
	"github.com/hwwadmin/yomi-boot/component/core/logger"
	"github.com/hwwadmin/yomi-boot/component/core/shared/dto"
	sharederr "github.com/hwwadmin/yomi-boot/component/core/shared/error"
)

// RevokeUserSessionInput HTTP 强制下线单个会话入参
type RevokeUserSessionInput struct {
	UserID     dto.JSONInt64 `json:"userId"`
	SessionID  string        `json:"sessionId"`
	OperatorID int64         `json:"-"`
}

// RevokeUserSessionUsecase 强制下线指定用户的一个会话
type RevokeUserSessionUsecase struct {
	tokens SessionTokens
	guard  kickGuard
}

// NewRevokeUserSessionUsecase 创建单个会话下线用例
func NewRevokeUserSessionUsecase(tokens SessionTokens, guard kickGuard) *RevokeUserSessionUsecase {
	return &RevokeUserSessionUsecase{tokens: tokens, guard: guard}
}

// Execute 校验会话归属后撤销该会话；已过期或不属于目标用户时按幂等成功处理
func (uc *RevokeUserSessionUsecase) Execute(ctx context.Context, in *RevokeUserSessionInput) error {
	userID := in.UserID.Int64()
	sessionID := strings.TrimSpace(in.SessionID)
	if userID <= 0 || sessionID == "" {
		return errors.Sentinel(sharederr.ErrInvalidInput)
	}

	sessions, err := uc.tokens.ListByUserID(ctx, userID)
	if err != nil {
		return err
	}

	found := false
	for _, session := range sessions {
		if session != nil && session.TokenHash == sessionID {
			found = true
			break
		}
	}
	if !found {
		// 已登出、已过期或不属于该用户：不撤销任何会话，避免拿别人的摘要绕过目标用户校验
		return nil
	}

	if userID == in.OperatorID && sessionID == currentTokenHash(ctx) {
		return support.ErrCannotOperateSelf
	}
	if err := uc.guard.EnsureCanKickUser(ctx, in.OperatorID, userID); err != nil {
		return err
	}

	if err := support.RevokeTokenHash(ctx, uc.tokens, sessionID); err != nil {
		return errors.WrapBiz(err, support.ErrSessionRevokeFailed.Code, support.ErrSessionRevokeFailed.Message)
	}
	logger.Info(ctx, "强制下线会话", "operatorID", in.OperatorID, "userID", userID, "tokenHash", sessionID)
	return nil
}

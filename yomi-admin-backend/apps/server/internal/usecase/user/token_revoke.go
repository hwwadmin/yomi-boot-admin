package user

import (
	"context"
	"time"
	"yomi-admin-backend/apps/server/internal/support"

	"gitee.com/hq_project/yomi-boot/component/common/token"
	"gitee.com/hq_project/yomi-boot/component/core/logger"
)

const (
	tokenRevokeAttempts = 3
	tokenRevokeTimeout  = 5 * time.Second
	tokenRevokeBackoff  = 100 * time.Millisecond
)

// revokeUserTokens 撤销用户全部令牌；失败时短重试，最终返回 ErrTokenRevokeFailed。
//
// 调用本函数时账号状态已经提交，所以失败必须原样返回给调用方：
// 提示要讲清"状态已变更、令牌未撤销"，让管理员重试补撤令牌。
// 使用脱离请求取消的 context，避免客户端断开导致安全动作被跳过。
func revokeUserTokens(ctx context.Context, tokens token.TokenService, userID int64) error {
	revokeCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), tokenRevokeTimeout)
	defer cancel()

	var lastErr error
	for attempt := 1; attempt <= tokenRevokeAttempts; attempt++ {
		lastErr = tokens.RevokeByUserID(revokeCtx, userID)
		if lastErr == nil {
			return nil
		}
		logger.Error(ctx, "撤销用户令牌失败", "userID", userID, "attempt", attempt, "error", lastErr)

		if attempt == tokenRevokeAttempts {
			break
		}
		select {
		case <-revokeCtx.Done():
		case <-time.After(tokenRevokeBackoff * time.Duration(attempt)):
		}
		if revokeCtx.Err() != nil {
			break
		}
	}

	logger.Error(ctx, "撤销用户令牌最终失败", "userID", userID, "attempts", tokenRevokeAttempts, "error", lastErr)
	return support.ErrTokenRevokeFailed
}

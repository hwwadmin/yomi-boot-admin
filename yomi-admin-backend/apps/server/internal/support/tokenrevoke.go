package support

import (
	"context"
	"time"

	"github.com/hwwadmin/yomi-boot/component/common/token"
	"github.com/hwwadmin/yomi-boot/component/core/logger"
)

// TokenRevokeAttempts 是撤销令牌的最大尝试次数
const TokenRevokeAttempts = 3

const (
	tokenRevokeTimeout = 5 * time.Second
	tokenRevokeBackoff = 100 * time.Millisecond
)

// RevokeUserTokens 撤销用户全部令牌；失败时短重试，最终返回账号变更场景的业务错误。
//
// 调用本函数时账号状态（停用、删除）或密码已经提交，所以失败必须原样返回给调用方：
// 提示要讲清"变更已生效、令牌未撤销"，让管理员重试补撤令牌。
func RevokeUserTokens(ctx context.Context, tokens token.TokenService, userID int64) error {
	if err := RevokeUserSessions(ctx, tokens, userID); err != nil {
		return ErrTokenRevokeFailed
	}
	return nil
}

// RevokeUserSessions 撤销用户全部令牌；失败时短重试，返回最后一次原始错误。
//
// 与 RevokeUserTokens 的区别只在错误投影：强制下线场景要返回"会话未撤销"的提示，
// 不能复用"账号状态已变更"的文案，所以由调用方包装自己的业务码。
func RevokeUserSessions(ctx context.Context, tokens token.TokenService, userID int64) error {
	return revokeWithRetry(ctx, "userID", userID, func(revokeCtx context.Context) error {
		return tokens.RevokeByUserID(revokeCtx, userID)
	})
}

// RevokeTokenHash 撤销指定哈希对应的单个令牌；失败时短重试，返回最后一次原始错误。
func RevokeTokenHash(ctx context.Context, tokens token.TokenService, tokenHash string) error {
	return revokeWithRetry(ctx, "tokenHash", tokenHash, func(revokeCtx context.Context) error {
		return tokens.RevokeByTokenHash(revokeCtx, tokenHash)
	})
}

// revokeWithRetry 在脱离请求取消的 context 中执行撤销动作并短重试。
//
// 使用脱离请求取消的 context，避免客户端断开导致安全动作被跳过。
func revokeWithRetry(ctx context.Context, field string, value any, action func(context.Context) error) error {
	revokeCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), tokenRevokeTimeout)
	defer cancel()

	var lastErr error
	for attempt := 1; attempt <= TokenRevokeAttempts; attempt++ {
		lastErr = action(revokeCtx)
		if lastErr == nil {
			return nil
		}
		logger.Error(ctx, "撤销令牌失败", field, value, "attempt", attempt, "error", lastErr)

		if attempt == TokenRevokeAttempts {
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

	logger.Error(ctx, "撤销令牌最终失败", field, value, "attempts", TokenRevokeAttempts, "error", lastErr)
	return lastErr
}

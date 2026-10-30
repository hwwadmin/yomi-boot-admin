package user

import (
	"context"

	"gitee.com/hq_project/yomi-boot/component/common/token"
	"gitee.com/hq_project/yomi-boot/component/core/bc/hash"
)

// SessionTokens 是会话管理需要的令牌能力：基础库的 TokenService 加上按用户列举。
//
// ListByUserID 已由 *token.Service 实现，但尚未进入 TokenService 接口，
// 因此由 bootstrap 做类型断言桥接；基础库补上接口后可以去掉这层断言。
type SessionTokens interface {
	token.TokenService
	ListByUserID(ctx context.Context, userID int64) ([]*token.UserTokenSession, error)
}

// kickGuard 判断操作者能否强制下线目标用户的会话
type kickGuard interface {
	EnsureCanKickUser(ctx context.Context, operatorID, targetUserID int64) error
}

// currentTokenHash 返回当前请求令牌的 SHA-256 摘要；没有令牌时返回空串。
func currentTokenHash(ctx context.Context) string {
	value, ok := token.TokenValueFromContext(ctx)
	if !ok || value == "" {
		return ""
	}
	return hash.SHA256Hex([]byte(value))
}

// sessionContextString 读取令牌上下文里的字符串字段；缺失或类型不符时返回空串。
func sessionContextString(session *token.TokenSession, key string) string {
	if session == nil {
		return ""
	}
	value, ok := session.ContextValue(key)
	if !ok {
		return ""
	}
	text, _ := value.(string)
	return text
}

package user

import (
	"context"

	"gitee.com/hq_project/yomi-boot/component/core/errors"
	"gitee.com/hq_project/yomi-boot/component/core/shared/dto"
	sharederr "gitee.com/hq_project/yomi-boot/component/core/shared/error"
)

// UserSessionItem 是用户在线会话列表项
type UserSessionItem struct {
	SessionID  string        `json:"sessionId"`  // SessionID 是令牌的 SHA-256 摘要，不是明文令牌
	UserID     dto.JSONInt64 `json:"userId"`     // UserID 是会话所属用户
	Username   string        `json:"username"`   // Username 是登录时写入的用户名快照
	Nickname   string        `json:"nickname"`   // Nickname 是登录时写入的昵称快照
	ClientType string        `json:"clientType"` // ClientType 是端类型，如 admin
	LoginIP    string        `json:"loginIp"`    // LoginIP 是登录 IP
	UserAgent  string        `json:"userAgent"`  // UserAgent 是登录时的 User-Agent
	LoginAt    int64         `json:"loginAt"`    // LoginAt 是登录时间（Unix 毫秒）
	ExpireAt   int64         `json:"expireAt"`   // ExpireAt 是到期时间（Unix 毫秒）
	Current    bool          `json:"current"`    // Current 表示该会话是否就是调用者当前使用的会话
}

// ListUserSessionsUsecase 查询指定用户的在线会话
type ListUserSessionsUsecase struct {
	tokens SessionTokens
}

// NewListUserSessionsUsecase 创建用户会话列表用例
func NewListUserSessionsUsecase(tokens SessionTokens) *ListUserSessionsUsecase {
	return &ListUserSessionsUsecase{tokens: tokens}
}

// Execute 返回指定用户当前有效的会话；没有会话时返回空切片而不是 nil
func (uc *ListUserSessionsUsecase) Execute(ctx context.Context, userID int64) ([]UserSessionItem, error) {
	if userID <= 0 {
		return nil, errors.Sentinel(sharederr.ErrInvalidInput)
	}

	sessions, err := uc.tokens.ListByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	currentHash := currentTokenHash(ctx)
	items := make([]UserSessionItem, 0, len(sessions))
	for _, session := range sessions {
		if session == nil || session.Session == nil {
			continue
		}
		snapshot := session.Session
		items = append(items, UserSessionItem{
			SessionID:  session.TokenHash,
			UserID:     dto.JSONInt64(snapshot.UserID),
			Username:   sessionContextString(snapshot, "username"),
			Nickname:   sessionContextString(snapshot, "nickname"),
			ClientType: snapshot.ClientType,
			LoginIP:    snapshot.LoginIP,
			UserAgent:  sessionContextString(snapshot, "user_agent"),
			LoginAt:    snapshot.IssuedAt.UnixMilli(),
			ExpireAt:   snapshot.ExpireAt.UnixMilli(),
			Current:    currentHash != "" && session.TokenHash == currentHash,
		})
	}
	return items, nil
}

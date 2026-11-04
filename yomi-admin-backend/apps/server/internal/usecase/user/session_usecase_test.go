package user

import (
	"context"
	"testing"
	"time"
	"yomi-admin-backend/apps/server/internal/support"
	"yomi-admin-backend/apps/server/internal/support/superadmin"

	"github.com/hwwadmin/yomi-boot/component/common/token"
	"github.com/hwwadmin/yomi-boot/component/core/bc/hash"
	"github.com/hwwadmin/yomi-boot/component/core/errors"
	"github.com/hwwadmin/yomi-boot/component/core/shared/dto"
	sharederr "github.com/hwwadmin/yomi-boot/component/core/shared/error"
)

// fakeSessionTokens 是会话令牌服务测试桩
type fakeSessionTokens struct {
	token.TokenService

	sessions []*token.UserTokenSession
	listErr  error

	revokedHashes []string
	revokedUsers  []int64
	revokeHashErr error
	revokeUserErr error
}

func (f *fakeSessionTokens) ListByUserID(context.Context, int64) ([]*token.UserTokenSession, error) {
	return f.sessions, f.listErr
}

func (f *fakeSessionTokens) RevokeByTokenHash(_ context.Context, tokenHash string) error {
	f.revokedHashes = append(f.revokedHashes, tokenHash)
	return f.revokeHashErr
}

func (f *fakeSessionTokens) RevokeByUserID(_ context.Context, userID int64) error {
	f.revokedUsers = append(f.revokedUsers, userID)
	return f.revokeUserErr
}

// fakeKickGuard 记录超管保护校验入参
type fakeKickGuard struct {
	called     int
	operatorID int64
	targetID   int64
	err        error
}

func (f *fakeKickGuard) EnsureCanKickUser(_ context.Context, operatorID, targetUserID int64) error {
	f.called++
	f.operatorID = operatorID
	f.targetID = targetUserID
	return f.err
}

// newTestSession 构造带登录元数据的会话快照
func newTestSession(userID int64, tokenHash, username string, issuedAt time.Time) *token.UserTokenSession {
	session := &token.TokenSession{
		UserID:     userID,
		ClientType: "admin",
		LoginIP:    "10.0.0.1",
		IssuedAt:   issuedAt,
		ExpireAt:   issuedAt.Add(2 * time.Hour),
	}
	session.SetContextValue("username", username)
	session.SetContextValue("nickname", username+"-nick")
	session.SetContextValue("user_agent", "test-agent")
	return &token.UserTokenSession{TokenHash: tokenHash, Session: session}
}

func TestListUserSessionsMapsSnapshot(t *testing.T) {
	issuedAt := time.Now().Add(-time.Hour)
	currentHash := hash.SHA256Hex([]byte("current-token"))
	tokens := &fakeSessionTokens{sessions: []*token.UserTokenSession{
		newTestSession(7, currentHash, "admin", issuedAt),
		newTestSession(7, "other-hash", "admin", issuedAt.Add(time.Minute)),
	}}
	uc := NewListUserSessionsUsecase(tokens)
	ctx := token.WithTokenValue(context.Background(), "current-token")

	items, err := uc.Execute(ctx, 7)
	if err != nil {
		t.Fatalf("查询会话失败: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("期望 2 条会话，实际 %d", len(items))
	}

	first := items[0]
	if first.SessionID != currentHash || first.UserID.Int64() != 7 ||
		first.Username != "admin" || first.Nickname != "admin-nick" ||
		first.ClientType != "admin" || first.LoginIP != "10.0.0.1" ||
		first.UserAgent != "test-agent" || !first.Current {
		t.Fatalf("第一条会话字段不符: %+v", first)
	}
	if first.LoginAt != issuedAt.UnixMilli() || first.ExpireAt != issuedAt.Add(2*time.Hour).UnixMilli() {
		t.Fatalf("时间字段不符: %+v", first)
	}
	if items[1].Current {
		t.Fatalf("非当前会话不应标记 current: %+v", items[1])
	}
}

func TestListUserSessionsRejectsInvalidUserID(t *testing.T) {
	uc := NewListUserSessionsUsecase(&fakeSessionTokens{})
	if _, err := uc.Execute(context.Background(), 0); !errors.Is(err, sharederr.ErrInvalidInput) {
		t.Fatalf("期望 ErrInvalidInput，实际 %v", err)
	}
}

func TestRevokeUserSessionIgnoresForeignKey(t *testing.T) {
	tokens := &fakeSessionTokens{sessions: []*token.UserTokenSession{
		newTestSession(7, "hash-1", "admin", time.Now()),
	}}
	guard := &fakeKickGuard{}
	uc := NewRevokeUserSessionUsecase(tokens, guard)

	in := &RevokeUserSessionInput{UserID: dto.JSONInt64(7), SessionID: "foreign-hash", OperatorID: 1}
	if err := uc.Execute(context.Background(), in); err != nil {
		t.Fatalf("不属于该用户的摘要应按幂等成功处理: %v", err)
	}
	if len(tokens.revokedHashes) != 0 || guard.called != 0 {
		t.Fatalf("不应撤销或校验超管: revoked=%v guard=%d", tokens.revokedHashes, guard.called)
	}
}

func TestRevokeUserSessionRejectsSelfCurrent(t *testing.T) {
	currentHash := hash.SHA256Hex([]byte("current-token"))
	tokens := &fakeSessionTokens{sessions: []*token.UserTokenSession{
		newTestSession(7, currentHash, "admin", time.Now()),
	}}
	guard := &fakeKickGuard{}
	uc := NewRevokeUserSessionUsecase(tokens, guard)
	ctx := token.WithTokenValue(context.Background(), "current-token")

	in := &RevokeUserSessionInput{UserID: dto.JSONInt64(7), SessionID: currentHash, OperatorID: 7}
	if err := uc.Execute(ctx, in); !errors.Is(err, support.ErrCannotOperateSelf) {
		t.Fatalf("期望 ErrCannotOperateSelf，实际 %v", err)
	}
	if len(tokens.revokedHashes) != 0 {
		t.Fatalf("不应撤销当前会话: %v", tokens.revokedHashes)
	}
}

func TestRevokeUserSessionAllowsSelfOther(t *testing.T) {
	currentHash := hash.SHA256Hex([]byte("current-token"))
	tokens := &fakeSessionTokens{sessions: []*token.UserTokenSession{
		newTestSession(7, currentHash, "admin", time.Now()),
		newTestSession(7, "other-hash", "admin", time.Now()),
	}}
	guard := &fakeKickGuard{}
	uc := NewRevokeUserSessionUsecase(tokens, guard)
	ctx := token.WithTokenValue(context.Background(), "current-token")

	in := &RevokeUserSessionInput{UserID: dto.JSONInt64(7), SessionID: "other-hash", OperatorID: 7}
	if err := uc.Execute(ctx, in); err != nil {
		t.Fatalf("下线自己的其它会话应放行: %v", err)
	}
	if len(tokens.revokedHashes) != 1 || tokens.revokedHashes[0] != "other-hash" {
		t.Fatalf("撤销目标不符: %v", tokens.revokedHashes)
	}
	if guard.called != 1 || guard.targetID != 7 {
		t.Fatalf("超管保护入参不符: %+v", guard)
	}
}

func TestRevokeUserSessionGuardDenied(t *testing.T) {
	tokens := &fakeSessionTokens{sessions: []*token.UserTokenSession{
		newTestSession(7, "hash-1", "admin", time.Now()),
	}}
	guard := &fakeKickGuard{err: superadmin.ErrKickSuperDenied}
	uc := NewRevokeUserSessionUsecase(tokens, guard)

	in := &RevokeUserSessionInput{UserID: dto.JSONInt64(7), SessionID: "hash-1", OperatorID: 1}
	if err := uc.Execute(context.Background(), in); !errors.Is(err, superadmin.ErrKickSuperDenied) {
		t.Fatalf("期望超管保护错误，实际 %v", err)
	}
	if len(tokens.revokedHashes) != 0 {
		t.Fatalf("超管保护拒绝后不应撤销: %v", tokens.revokedHashes)
	}
}

func TestRevokeUserSessionRevokeFailure(t *testing.T) {
	tokens := &fakeSessionTokens{
		sessions:      []*token.UserTokenSession{newTestSession(7, "hash-1", "admin", time.Now())},
		revokeHashErr: errors.New("redis down"),
	}
	uc := NewRevokeUserSessionUsecase(tokens, &fakeKickGuard{})

	in := &RevokeUserSessionInput{UserID: dto.JSONInt64(7), SessionID: "hash-1", OperatorID: 1}
	if err := uc.Execute(context.Background(), in); !errors.Is(err, support.ErrSessionRevokeFailed) {
		t.Fatalf("期望 ErrSessionRevokeFailed，实际 %v", err)
	}
}

func TestRevokeUserSessionsRejectsSelf(t *testing.T) {
	tokens := &fakeSessionTokens{}
	guard := &fakeKickGuard{}
	uc := NewRevokeUserSessionsUsecase(tokens, guard)

	in := &RevokeUserSessionsInput{UserID: dto.JSONInt64(7), OperatorID: 7}
	if err := uc.Execute(context.Background(), in); !errors.Is(err, support.ErrCannotOperateSelf) {
		t.Fatalf("期望 ErrCannotOperateSelf，实际 %v", err)
	}
	if len(tokens.revokedUsers) != 0 || guard.called != 0 {
		t.Fatalf("拒绝后不应撤销或校验超管: revoked=%v guard=%d", tokens.revokedUsers, guard.called)
	}
}

func TestRevokeUserSessionsGuardDenied(t *testing.T) {
	tokens := &fakeSessionTokens{}
	guard := &fakeKickGuard{err: superadmin.ErrKickSuperDenied}
	uc := NewRevokeUserSessionsUsecase(tokens, guard)

	in := &RevokeUserSessionsInput{UserID: dto.JSONInt64(7), OperatorID: 1}
	if err := uc.Execute(context.Background(), in); !errors.Is(err, superadmin.ErrKickSuperDenied) {
		t.Fatalf("期望超管保护错误，实际 %v", err)
	}
	if len(tokens.revokedUsers) != 0 {
		t.Fatalf("超管保护拒绝后不应撤销: %v", tokens.revokedUsers)
	}
}

func TestRevokeUserSessionsSuccess(t *testing.T) {
	tokens := &fakeSessionTokens{}
	guard := &fakeKickGuard{}
	uc := NewRevokeUserSessionsUsecase(tokens, guard)

	in := &RevokeUserSessionsInput{UserID: dto.JSONInt64(7), OperatorID: 1}
	if err := uc.Execute(context.Background(), in); err != nil {
		t.Fatalf("按用户下线失败: %v", err)
	}
	if len(tokens.revokedUsers) != 1 || tokens.revokedUsers[0] != 7 {
		t.Fatalf("撤销用户不符: %v", tokens.revokedUsers)
	}
	if guard.called != 1 || guard.operatorID != 1 || guard.targetID != 7 {
		t.Fatalf("超管保护入参不符: %+v", guard)
	}
}

func TestRevokeUserSessionsRevokeFailure(t *testing.T) {
	tokens := &fakeSessionTokens{revokeUserErr: errors.New("redis down")}
	uc := NewRevokeUserSessionsUsecase(tokens, &fakeKickGuard{})

	in := &RevokeUserSessionsInput{UserID: dto.JSONInt64(7), OperatorID: 1}
	if err := uc.Execute(context.Background(), in); !errors.Is(err, support.ErrSessionRevokeFailed) {
		t.Fatalf("期望 ErrSessionRevokeFailed，实际 %v", err)
	}
}

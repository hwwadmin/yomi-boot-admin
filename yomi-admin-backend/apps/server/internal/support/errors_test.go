package support

import (
	"strings"
	"testing"
)

// TestErrTokenRevokeFailedMessage 校验提示明确表达"状态已变更、令牌未撤销、需要重试"。
//
// 该错误在账号状态已提交后才返回，措辞是管理员判断"是否需要重试"的唯一依据，
// 因此这里锁定三个关键信息，避免后续改文案时丢失。
func TestErrTokenRevokeFailedMessage(t *testing.T) {
	message := ErrTokenRevokeFailed.Error()
	for _, want := range []string{"已变更", "未撤销", "请重试"} {
		if !strings.Contains(message, want) {
			t.Fatalf("错误提示缺少关键字 %q: %s", want, message)
		}
	}
	if ErrTokenRevokeFailed.Code != "ADMIN_TOKEN_REVOKE_FAILED" {
		t.Fatalf("错误码变更会影响前端与日志检索: %s", ErrTokenRevokeFailed.Code)
	}
}

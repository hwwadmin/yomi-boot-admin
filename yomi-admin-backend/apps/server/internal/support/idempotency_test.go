package support

import (
	"context"
	"testing"

	frameworkidem "gitee.com/hq_project/yomi-boot/component/common/idempotency"
	"gitee.com/hq_project/yomi-boot/component/core/errors"
)

// fakeRunner 在当前协程执行事务闭包并记录调用次数
type fakeRunner struct {
	runs int
}

func (f *fakeRunner) Run(ctx context.Context, fn func(ctx context.Context) error) error {
	f.runs++
	return fn(ctx)
}

func (f *fakeRunner) InTx(context.Context) bool { return false }

func TestIdempotencyKey(t *testing.T) {
	t.Run("缺省返回零值键", func(t *testing.T) {
		key, err := IdempotencyKey("  ", "CREATE_ROLE")
		if err != nil {
			t.Fatalf("缺省幂等键不应报错: %v", err)
		}
		if key.BizID != "" || key.BizSubID != "" || key.OperationType != "" {
			t.Fatalf("缺省幂等键应为零值: %+v", key)
		}
	})

	t.Run("有效键按操作类型落子键", func(t *testing.T) {
		key, err := IdempotencyKey(" 3f2a-uuid ", "CREATE_ROLE")
		if err != nil {
			t.Fatalf("有效幂等键不应报错: %v", err)
		}
		if key.BizID != "3f2a-uuid" {
			t.Fatalf("BizID = %q, want 3f2a-uuid", key.BizID)
		}
		if key.BizSubID != "CREATE_ROLE" || key.OperationType != "CREATE_ROLE" {
			t.Fatalf("子键/操作类型不符合预期: %+v", key)
		}
	})

	t.Run("超长键拒绝", func(t *testing.T) {
		long := make([]byte, maxIdempotencyKeyBytes+1)
		for i := range long {
			long[i] = 'a'
		}
		_, err := IdempotencyKey(string(long), "CREATE_ROLE")
		if !errors.Is(err, ErrIdempotencyKeyInvalid) {
			t.Fatalf("超长幂等键应返回 ErrIdempotencyKeyInvalid, got %v", err)
		}
	})
}

func TestRunIdempotentWithoutKeyRunsOnce(t *testing.T) {
	runner := &fakeRunner{}
	calls := 0

	out, err := RunIdempotent(nil, runner, context.Background(), frameworkidem.Key{}, func(ctx context.Context) (string, error) {
		calls++
		return "ok", nil
	})
	if err != nil {
		t.Fatalf("无幂等键执行失败: %v", err)
	}
	if out != "ok" {
		t.Fatalf("out = %q, want ok", out)
	}
	if runner.runs != 1 || calls != 1 {
		t.Fatalf("runner.runs = %d, calls = %d, want 1/1", runner.runs, calls)
	}
}

package support

import (
	"context"
	"regexp"
	"strings"

	dbtx "github.com/hwwadmin/yomi-boot/component/common/db/tx"
	frameworkidem "github.com/hwwadmin/yomi-boot/component/common/idempotency"
	"github.com/hwwadmin/yomi-boot/component/core/errors"
)

// idempotencyPartPattern 与 yomi-boot idempotency.Key 的 biz_id / biz_sub_id 格式一致：
// 1～100 位字母、数字、下划线或连字符，且不含冒号。
var idempotencyPartPattern = regexp.MustCompile(`^[A-Za-z0-9_-]{1,100}$`)

const (
	// IdempotencyKeyHeader 是调用方传递幂等键的请求头。
	//
	// 同一次提交意图（含超时重试）复用同一个值；缺省时不做幂等保护。
	IdempotencyKeyHeader = "X-Idempotency-Key"
	// maxIdempotencyKeyBytes 与 t_sys_biz.biz_id 列宽（VARCHAR(100)）一致
	maxIdempotencyKeyBytes = 100
)

// IdempotencyKey 由请求头幂等键和操作类型构造框架幂等键。
//
// raw 为空表示调用方未提供幂等键，返回零值 Key，由 RunIdempotent 退化为一次性事务执行。
// 键与操作类型都必须是 1～100 位字母、数字、下划线或连字符，否则直接拒绝。
func IdempotencyKey(raw, operation string) (frameworkidem.Key, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return frameworkidem.Key{}, nil
	}
	if !idempotencyPartPattern.MatchString(raw) || !idempotencyPartPattern.MatchString(operation) {
		return frameworkidem.Key{}, errors.Sentinel(ErrIdempotencyKeyInvalid)
	}
	// BizSubID 用操作类型而不是随机值：同一个键复用到不同操作时落在不同的幂等记录上，
	// 不会把 A 操作的结果回放给 B 操作。
	return frameworkidem.Key{
		BizID:         raw,
		BizSubID:      operation,
		OperationType: operation,
	}, nil
}

// RunIdempotent 执行幂等保护：提供了幂等键时走框架 idempotency.Execute，
// 未提供时退化为一次性事务执行（不写 t_sys_biz，保持原有语义）。
//
// fn 内的所有写操作必须使用传入的 txCtx，才能与幂等抢占记录同事务提交。
func RunIdempotent[T any](
	guard *frameworkidem.Guard,
	runner dbtx.Runner,
	ctx context.Context,
	key frameworkidem.Key,
	fn func(ctx context.Context) (T, error),
) (T, error) {
	if key.BizID == "" {
		var out T
		err := runner.Run(ctx, func(txCtx context.Context) error {
			var innerErr error
			out, innerErr = fn(txCtx)
			return innerErr
		})
		return out, err
	}
	return frameworkidem.Execute(guard, runner, ctx, key, fn)
}

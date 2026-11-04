package handler

import (
	"context"
	"yomi-admin-backend/apps/server/internal/support/audit"

	"github.com/hwwadmin/yomi-boot/component/common/token"
	"github.com/hwwadmin/yomi-boot/component/core/errors"
)

// auditOperator 从令牌会话读取操作者快照；取不到时返回零值
func auditOperator(ctx context.Context) audit.Operator {
	id, _ := token.OperatorID(ctx)
	name := ""
	if value, ok := token.TokenContextValueFromContext(ctx, "username"); ok {
		name, _ = value.(string)
	}
	return audit.Operator{ID: id, Name: name}
}

// auditRun 在业务前写 STARTED、业务后写 FINISHED；审计失败不影响业务。
//
// ctx 由调用方提供：普通请求传请求上下文；幂等包裹内传事务上下文，保证审计与业务同属一个执行边界。
func auditRun[T any](
	ctx context.Context,
	recorder audit.Recorder,
	start audit.Spec,
	finish func(out T, err error) audit.Spec,
	fn func(ctx context.Context) (T, error),
) (T, error) {
	if recorder == nil {
		return fn(ctx)
	}

	ref := recorder.Start(ctx, start)
	var out T
	var err error

	defer func() {
		if rec := recover(); rec != nil {
			panicErr := errors.New("panic")
			recorder.Finish(ctx, ref, finish(out, panicErr), audit.OutcomeFromError(panicErr))
			panic(rec)
		}
		recorder.Finish(ctx, ref, finish(out, err), audit.OutcomeFromError(err))
	}()

	out, err = fn(ctx)
	return out, err
}

// auditVoid 是 auditRun 的无返回值版本
func auditVoid(
	ctx context.Context,
	recorder audit.Recorder,
	start audit.Spec,
	finish func(err error) audit.Spec,
	fn func(ctx context.Context) error,
) error {
	_, err := auditRun(ctx, recorder, start, func(_ struct{}, err error) audit.Spec {
		return finish(err)
	}, func(ctx context.Context) (struct{}, error) {
		return struct{}{}, fn(ctx)
	})
	return err
}

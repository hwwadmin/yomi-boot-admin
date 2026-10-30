package handler

import (
	"context"
	"yomi-admin-backend/apps/server/internal/support/audit"

	"gitee.com/hq_project/yomi-boot/component/common/token"
	"gitee.com/hq_project/yomi-boot/component/core/errors"

	"github.com/gin-gonic/gin"
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

// auditRun 在业务前写 STARTED、业务后写 FINISHED；审计失败不影响业务
func auditRun[T any](
	c *gin.Context,
	recorder audit.Recorder,
	start audit.Spec,
	finish func(out T, err error) audit.Spec,
	fn func(ctx context.Context) (T, error),
) (T, error) {
	if recorder == nil {
		return fn(c.Request.Context())
	}

	ref := recorder.Start(c.Request.Context(), start)
	var out T
	var err error

	defer func() {
		if rec := recover(); rec != nil {
			panicErr := errors.New("panic")
			recorder.Finish(c.Request.Context(), ref, finish(out, panicErr), audit.OutcomeFromError(panicErr))
			panic(rec)
		}
		recorder.Finish(c.Request.Context(), ref, finish(out, err), audit.OutcomeFromError(err))
	}()

	out, err = fn(c.Request.Context())
	return out, err
}

// auditVoid 是 auditRun 的无返回值版本
func auditVoid(
	c *gin.Context,
	recorder audit.Recorder,
	start audit.Spec,
	finish func(err error) audit.Spec,
	fn func(ctx context.Context) error,
) error {
	_, err := auditRun(c, recorder, start, func(_ struct{}, err error) audit.Spec {
		return finish(err)
	}, func(ctx context.Context) (struct{}, error) {
		return struct{}{}, fn(ctx)
	})
	return err
}

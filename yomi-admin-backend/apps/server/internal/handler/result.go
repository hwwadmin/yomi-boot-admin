package handler

import (
	"context"

	"gitee.com/hq_project/yomi-boot/component/core/errors"
	corehandler "gitee.com/hq_project/yomi-boot/component/core/handler"
	sharederr "gitee.com/hq_project/yomi-boot/component/core/shared/error"

	"github.com/gin-gonic/gin"
)

// wireSentinels 是框架层直接作为 code 使用的通用错误，属于对外契约，不能收敛。
var wireSentinels = []error{
	sharederr.ErrInvalidBody,
	sharederr.ErrInvalidInput,
	sharederr.ErrRequestTimeout,
	sharederr.ErrRequestBodyTooLarge,
}

// failureFallback 判断错误是否需要兜底为 INTERNAL_ERROR；nil 表示保留原错误。
func failureFallback(err error) error {
	for _, sentinel := range wireSentinels {
		if errors.Is(err, sentinel) {
			return nil
		}
	}
	return sharederr.ErrInternal
}

// handle 执行业务用例并输出响应。
// 用例只应返回 BizError 或内部错误：BizError 与框架通用码保留自己的 code，
// 其余无法识别的错误统一收敛为 INTERNAL_ERROR，避免内部错误文本被当作业务码返回给客户端。
func handle[T any](c *gin.Context, fn func(ctx context.Context) (T, error)) {
	output, err := fn(c.Request.Context())
	if err != nil {
		corehandler.WriteBusinessErrorWithFallback(c, err, failureFallback(err))
		return
	}
	corehandler.WriteSuccess(c, output)
}

// handleVoid 同 handle，用于无返回数据的用例。
func handleVoid(c *gin.Context, fn func(ctx context.Context) error) {
	if err := fn(c.Request.Context()); err != nil {
		corehandler.WriteBusinessErrorWithFallback(c, err, failureFallback(err))
		return
	}
	corehandler.WriteSuccess(c, nil)
}

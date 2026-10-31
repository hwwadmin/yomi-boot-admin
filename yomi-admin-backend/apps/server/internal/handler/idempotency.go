package handler

import (
	"yomi-admin-backend/apps/server/internal/support"

	"github.com/gin-gonic/gin"
)

// idempotencyKey 读取请求头中的幂等键；缺省时返回空串，用例退化为一次性事务执行。
func idempotencyKey(c *gin.Context) string {
	return c.GetHeader(support.IdempotencyKeyHeader)
}

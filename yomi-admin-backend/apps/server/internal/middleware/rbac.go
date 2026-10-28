// Package middleware 编排层 HTTP 中间件
package middleware

import (
	"net/http"
	"strings"
	rbaccontract "yomi-admin-backend/domains/rbac/service_contract"

	"gitee.com/hq_project/yomi-boot/component/common/token"
	"gitee.com/hq_project/yomi-boot/component/core/errors"
	corehandler "gitee.com/hq_project/yomi-boot/component/core/handler"

	"github.com/gin-gonic/gin"
)

// RBAC 按接口路径模板校验接口权限；skip 中的路径与 OPTIONS 请求直接放行
func RBAC(auth rbaccontract.RbacAuthService, skip map[string]struct{}) gin.HandlerFunc {
	return func(c *gin.Context) {
		if _, ok := skip[c.Request.URL.Path]; ok {
			c.Next()
			return
		}
		if c.Request.Method == http.MethodOptions {
			c.Next()
			return
		}

		path := c.FullPath()
		if path == "" {
			c.Next()
			return
		}

		userID, err := token.OperatorID(c.Request.Context())
		if err != nil {
			corehandler.WriteBusinessError(c, err)
			c.Abort()
			return
		}

		allowed, err := auth.CanAccessAPI(c.Request.Context(), userID, strings.ToUpper(c.Request.Method), path)
		if err != nil {
			corehandler.WriteBusinessError(c, err)
			c.Abort()
			return
		}
		if !allowed {
			corehandler.WriteBusinessError(c, errors.Sentinel(rbaccontract.ErrPermissionDenied))
			c.Abort()
			return
		}
		c.Next()
	}
}

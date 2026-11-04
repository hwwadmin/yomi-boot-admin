// Package middleware 编排层 HTTP 中间件
package middleware

import (
	"net/http"
	"strings"
	rbaccontract "yomi-admin-backend/domains/rbac/service_contract"

	"github.com/hwwadmin/yomi-boot/component/common/token"
	"github.com/hwwadmin/yomi-boot/component/core/errors"
	corehandler "github.com/hwwadmin/yomi-boot/component/core/handler"

	"github.com/gin-gonic/gin"
)

// RBAC 按接口路径模板校验接口权限；skip 中的"方法 + 路径"与 OPTIONS 请求直接放行。
//
// skip 由 biz.auth.public_paths 生成，键格式为 "GET /api/auth/login"，
// 方法与路径都匹配才跳过，避免同名路径的其它方法被顺带放行。
func RBAC(auth rbaccontract.RbacAuthService, skip map[string]struct{}) gin.HandlerFunc {
	return func(c *gin.Context) {
		if _, ok := skip[skipKey(c.Request.Method, c.Request.URL.Path)]; ok {
			c.Next()
			return
		}
		if c.Request.Method == http.MethodOptions {
			c.Next()
			return
		}

		path := c.FullPath()
		if path == "" {
			// 未匹配到路由（404/405）：默认拒绝，避免将来挂上兜底处理器时绕过鉴权
			corehandler.WriteBusinessError(c, errors.Sentinel(rbaccontract.ErrPermissionDenied))
			c.Abort()
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

// skipKey 组装跳过清单的键：方法统一大写，路径按请求原始路径
func skipKey(method, path string) string {
	return strings.ToUpper(method) + " " + path
}

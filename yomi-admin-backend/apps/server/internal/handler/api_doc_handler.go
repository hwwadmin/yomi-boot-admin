// Package handler HTTP 适配层
package handler

import (
	"embed"
	"fmt"
	"net/http"
	"yomi-admin-backend/apps/server/api"

	"github.com/gin-gonic/gin"
)

//go:embed swaggerui/swagger-ui-bundle.js swaggerui/swagger-ui.css
var swaggerUI embed.FS

// API 文档路由段（相对 router group；实际路径带 /api 前缀）
const (
	apiDocPagePath   = "/system/api-doc"
	apiDocSchemaPath = "/system/api-doc/schema"
	apiDocCSSPath    = "/system/api-doc/swagger-ui.css"
	apiDocJSPath     = "/system/api-doc/swagger-ui-bundle.js"
)

// apiDocPageTemplate 是 Swagger UI 页面模板；三个 %s 依次为 css/js/schema 绝对地址
const apiDocPageTemplate = `<!DOCTYPE html>
<html lang="zh-CN">
<head>
  <meta charset="UTF-8"/>
  <title>yomi-admin API 文档</title>
  <!-- data:, 声明空图标，避免浏览器请求 /favicon.ico 落进鉴权中间件产生 ERROR 日志 -->
  <link rel="icon" href="data:,"/>
  <link rel="stylesheet" href="%s"/>
</head>
<body>
<div id="swagger-ui"></div>
<script src="%s"></script>
<script>
  window.onload = function () {
    window.ui = SwaggerUIBundle({
      url: '%s',
      dom_id: '#swagger-ui',
      persistAuthorization: true,
      validatorUrl: null
    });
  };
</script>
</body>
</html>
`

// ApiDocHandler API 文档路由（Swagger UI + 内嵌 OpenAPI 3.1 schema）
type ApiDocHandler struct {
	pageHTML []byte
	cssData  []byte
	jsData   []byte
}

// NewApiDocHandler 创建 API 文档 Handler
func NewApiDocHandler() *ApiDocHandler {
	// go:embed 保证文件存在，读取不会失败
	css, _ := swaggerUI.ReadFile("swaggerui/swagger-ui.css")
	js, _ := swaggerUI.ReadFile("swaggerui/swagger-ui-bundle.js")
	return &ApiDocHandler{cssData: css, jsData: js}
}

// Routes 注册文档路由；页面内三个地址均按当前 router 前缀拼接为绝对路径
func (h *ApiDocHandler) Routes(router *gin.RouterGroup) {
	base := router.BasePath()
	h.pageHTML = []byte(fmt.Sprintf(apiDocPageTemplate,
		base+apiDocCSSPath, base+apiDocJSPath, base+apiDocSchemaPath))
	router.GET(apiDocPagePath, h.page)
	router.GET(apiDocSchemaPath, h.schema)
	router.GET(apiDocCSSPath, h.css)
	router.GET(apiDocJSPath, h.js)
}

func (h *ApiDocHandler) page(c *gin.Context) {
	c.Header("Cache-Control", "no-store")
	c.Data(http.StatusOK, "text/html; charset=utf-8", h.pageHTML)
}

func (h *ApiDocHandler) schema(c *gin.Context) {
	c.Header("Cache-Control", "no-store")
	c.Data(http.StatusOK, "application/yaml; charset=utf-8", api.Spec)
}

func (h *ApiDocHandler) css(c *gin.Context) {
	c.Header("Cache-Control", "public, max-age=86400")
	c.Data(http.StatusOK, "text/css; charset=utf-8", h.cssData)
}

func (h *ApiDocHandler) js(c *gin.Context) {
	c.Header("Cache-Control", "public, max-age=86400")
	c.Data(http.StatusOK, "text/javascript; charset=utf-8", h.jsData)
}

// Package handler API 文档与路由对账
package handler

import (
	"sort"
	"strings"
	"testing"
	"yomi-admin-backend/apps/server/api"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/gin-gonic/gin"
)

// apidocSelfRoutes 是 API 文档自身路由，不写入文档
var apidocSelfRoutes = map[string]struct{}{
	"GET /api/system/api-doc":                      {},
	"GET /api/system/api-doc/schema":               {},
	"GET /api/system/api-doc/swagger-ui.css":       {},
	"GET /api/system/api-doc/swagger-ui-bundle.js": {},
}

// TestAPIDocCoverage 校验 openapi.yaml 与已注册路由双向一致。
// API 文档自身的 4 条路由不写入文档，健康检查探针不注册在本测试引擎中。
func TestAPIDocCoverage(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	// 路由的 /api 前缀已写在各 handler 的 URI 里，测试引擎不再加 group 前缀
	group := engine.Group("")
	(&AuthHandler{}).Routes(group)
	(&UserHandler{}).Routes(group)
	(&UserSessionHandler{}).Routes(group)
	(&RoleHandler{}).Routes(group)
	(&PermissionHandler{}).Routes(group)
	(&PermissionGroupHandler{}).Routes(group)
	(&HttpRecordHandler{}).Routes(group)
	(&RbacCacheHandler{}).Routes(group)
	(&SysParamHandler{}).Routes(group)
	NewApiDocHandler().Routes(group)

	registered := make(map[string]struct{})
	for _, route := range engine.Routes() {
		registered[route.Method+" "+route.Path] = struct{}{}
	}

	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromData(api.Spec)
	if err != nil {
		t.Fatalf("openapi.yaml 解析失败: %v", err)
	}
	documented := make(map[string]struct{})
	for _, path := range doc.Paths.Keys() {
		item := doc.Paths.Value(path)
		if item == nil {
			continue
		}
		for method := range item.Operations() {
			documented[strings.ToUpper(method)+" "+path] = struct{}{}
		}
	}

	var missing []string
	for key := range registered {
		if _, ok := apidocSelfRoutes[key]; ok {
			continue
		}
		if _, ok := documented[key]; !ok {
			missing = append(missing, key)
		}
	}
	sort.Strings(missing)
	if len(missing) > 0 {
		t.Fatalf("以下路由未写入 openapi.yaml: %v", missing)
	}

	var stale []string
	for key := range documented {
		if _, ok := registered[key]; !ok {
			stale = append(stale, key)
		}
	}
	sort.Strings(stale)
	if len(stale) > 0 {
		t.Fatalf("openapi.yaml 包含未注册的路径: %v", stale)
	}
}

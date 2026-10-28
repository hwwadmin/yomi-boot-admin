// Package handler 路由与权限目录对账
package handler

import (
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"testing"

	"github.com/gin-gonic/gin"
)

// skipRBACCatalog 无需 API 权限的公开路由
var skipRBACCatalog = map[string]struct{}{
	"POST /api/auth/login":  {},
	"GET /api/auth/profile": {},
	"POST /api/auth/logout": {},
}

// TestRBACRouteCatalogCoverage 校验已注册路由与权限种子双向一致
func TestRBACRouteCatalogCoverage(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	group := engine.Group("/api")
	(&AuthHandler{}).Routes(group)
	(&UserHandler{}).Routes(group)
	(&RbacHandler{}).Routes(group)
	(&HttpRecordHandler{}).Routes(group)

	registered := make(map[string]struct{})
	for _, route := range engine.Routes() {
		key := route.Method + " " + route.Path
		if _, skip := skipRBACCatalog[key]; skip {
			continue
		}
		registered[key] = struct{}{}
	}

	catalog := parseAPIPermissionCatalog(t)

	var missing []string
	for key := range registered {
		if _, ok := catalog[key]; !ok {
			missing = append(missing, key)
		}
	}
	sort.Strings(missing)
	if len(missing) > 0 {
		t.Fatalf("权限种子缺少已注册路由: %v", missing)
	}

	var stale []string
	for key := range catalog {
		if _, ok := registered[key]; !ok {
			stale = append(stale, key)
		}
	}
	sort.Strings(stale)
	if len(stale) > 0 {
		t.Fatalf("权限种子包含不存在的路由: %v", stale)
	}
}

// parseAPIPermissionCatalog 从 RBAC 种子 SQL 解析 API 权限键
func parseAPIPermissionCatalog(t *testing.T) map[string]struct{} {
	t.Helper()

	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("获取测试文件路径失败")
	}
	sqlPath := filepath.Join(filepath.Dir(file), "..", "..", "..", "..", "script", "sql", "020_rbac.sql")
	content, err := os.ReadFile(sqlPath)
	if err != nil {
		t.Fatalf("读取权限种子失败: %v", err)
	}

	pattern := regexp.MustCompile(`'API',\s*NULL,\s*'([A-Z]+)',\s*'(/api/[^']+)'`)
	matches := pattern.FindAllStringSubmatch(string(content), -1)
	if len(matches) == 0 {
		t.Fatal("权限种子中未解析到 API 权限")
	}

	catalog := make(map[string]struct{}, len(matches))
	for _, match := range matches {
		catalog[match[1]+" "+match[2]] = struct{}{}
	}
	return catalog
}

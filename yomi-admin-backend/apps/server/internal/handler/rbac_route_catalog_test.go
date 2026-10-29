// Package handler 路由与权限种子对账
package handler

import (
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"testing"

	"github.com/gin-gonic/gin"
)

// skipRBACCatalog 无需 API 权限的公开路由
var skipRBACCatalog = map[string]struct{}{
	"GET /api/auth/captcha": {},
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
	(&RoleHandler{}).Routes(group)
	(&PermissionHandler{}).Routes(group)
	(&PermissionGroupHandler{}).Routes(group)
	(&HttpRecordHandler{}).Routes(group)
	(&RbacCacheHandler{}).Routes(group)

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

// TestRBACSeedPermissionGroupCoverage 校验种子权限全部被权限组覆盖
func TestRBACSeedPermissionGroupCoverage(t *testing.T) {
	content := readRBACSeed(t)

	permissionPattern := regexp.MustCompile("(?s)INSERT IGNORE INTO `t_rbac_permission`[^;]*VALUES \\((\\d+),")
	permissionIDs := make(map[int64]struct{})
	for _, match := range permissionPattern.FindAllStringSubmatch(content, -1) {
		permissionIDs[parseSeedID(t, match[1])] = struct{}{}
	}
	if len(permissionIDs) == 0 {
		t.Fatal("权限种子中未解析到权限")
	}

	groupPattern := regexp.MustCompile("(?s)INSERT IGNORE INTO `t_rbac_permission_group`[^;]*VALUES \\((\\d+),")
	groupIDs := make(map[int64]struct{})
	for _, match := range groupPattern.FindAllStringSubmatch(content, -1) {
		groupIDs[parseSeedID(t, match[1])] = struct{}{}
	}
	if len(groupIDs) == 0 {
		t.Fatal("权限种子中未解析到权限组")
	}

	relPattern := regexp.MustCompile("INSERT IGNORE INTO `t_rbac_group_permission_rel`[^;]*VALUES \\((\\d+), (\\d+), (\\d+)\\)")
	covered := make(map[int64]struct{})
	for _, match := range relPattern.FindAllStringSubmatch(content, -1) {
		groupID := parseSeedID(t, match[2])
		permissionID := parseSeedID(t, match[3])
		if _, ok := groupIDs[groupID]; !ok {
			t.Fatalf("权限组关系引用了不存在的权限组: %d", groupID)
		}
		if _, ok := permissionIDs[permissionID]; !ok {
			t.Fatalf("权限组关系引用了不存在的权限: %d", permissionID)
		}
		covered[permissionID] = struct{}{}
	}

	var uncovered []int64
	for permissionID := range permissionIDs {
		if _, ok := covered[permissionID]; !ok {
			uncovered = append(uncovered, permissionID)
		}
	}
	sort.Slice(uncovered, func(i, j int) bool { return uncovered[i] < uncovered[j] })
	if len(uncovered) > 0 {
		t.Fatalf("以下权限未加入任何权限组: %v", uncovered)
	}
}

// parseAPIPermissionCatalog 从 RBAC 种子 SQL 解析 API 权限键
func parseAPIPermissionCatalog(t *testing.T) map[string]struct{} {
	t.Helper()

	content := readRBACSeed(t)
	pattern := regexp.MustCompile(`'API',\s*NULL,\s*'([A-Z]+)',\s*'(/api/[^']+)'`)
	matches := pattern.FindAllStringSubmatch(content, -1)
	if len(matches) == 0 {
		t.Fatal("权限种子中未解析到 API 权限")
	}

	catalog := make(map[string]struct{}, len(matches))
	for _, match := range matches {
		catalog[match[1]+" "+match[2]] = struct{}{}
	}
	return catalog
}

// readRBACSeed 读取 RBAC 种子 SQL
func readRBACSeed(t *testing.T) string {
	t.Helper()

	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("获取测试文件路径失败")
	}
	sqlPath := filepath.Join(filepath.Dir(file), "..", "..", "..", "..", "script", "sql", "020_rbac.sql")
	content, err := os.ReadFile(sqlPath)
	if err != nil {
		t.Fatalf("读取 RBAC 种子失败: %v", err)
	}
	return string(content)
}

// parseSeedID 解析种子 SQL 中的雪花 ID
func parseSeedID(t *testing.T, raw string) int64 {
	t.Helper()
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		t.Fatalf("非法种子 ID %q: %v", raw, err)
	}
	return id
}

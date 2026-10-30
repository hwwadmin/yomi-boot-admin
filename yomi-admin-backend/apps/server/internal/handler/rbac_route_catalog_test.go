// Package handler 路由与权限种子对账
package handler

import (
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"testing"
	svcconfig "yomi-admin-backend/apps/server/internal/config"

	"gitee.com/hq_project/yomi-boot/component/core/appctx"
	coreconfig "gitee.com/hq_project/yomi-boot/component/core/config"
	corehandler "gitee.com/hq_project/yomi-boot/component/core/handler"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

// TestRBACRouteCatalogCoverage 校验已注册路由、权限种子与公开路由配置三方一致。
//
// 公开路由（免接口权限）只来自 config/dev.yml 的 biz.auth.public_paths，
// 这里不再维护任何硬编码清单：
//   - 已注册但既不是公开路由、也没有权限种子 → 说明有人加了接口忘记配权限；
//   - 权限种子里有、但没有对应路由 → 说明权限种子过期；
//   - 配置声明为公开、但没有对应路由 → 多半是路径写错，会让本该受保护的路由意外公开。
func TestRBACRouteCatalogCoverage(t *testing.T) {
	gin.SetMode(gin.TestMode)
	prefix, publicRoutes := readPublicRoutes(t)

	engine := gin.New()
	group := engine.Group(prefix)
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
	corehandler.NewHealthHandler("/api/v1/health", "/api/v1/health/ops",
		appctx.NewAppCtx("yomi-admin-test", "test", "test")).Routes(group)

	registered := make(map[string]struct{})
	for _, route := range engine.Routes() {
		registered[route.Method+" "+route.Path] = struct{}{}
	}

	catalog := parseAPIPermissionCatalog(t)

	var missing []string
	for key := range registered {
		if _, public := publicRoutes[key]; public {
			continue
		}
		if _, ok := catalog[key]; !ok {
			missing = append(missing, key)
		}
	}
	sort.Strings(missing)
	if len(missing) > 0 {
		t.Fatalf("以下路由既不是公开路由、也没有权限种子（请在 biz.auth.public_paths 或 020_rbac.sql 中登记）: %v", missing)
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

	var unregistered []string
	for key := range publicRoutes {
		if _, ok := registered[key]; !ok {
			unregistered = append(unregistered, key)
		}
	}
	sort.Strings(unregistered)
	if len(unregistered) > 0 {
		t.Fatalf("biz.auth.public_paths 声明了不存在的路由（请检查路径是否写了路由前缀或拼写错误）: %v", unregistered)
	}
}

// readPublicRoutes 从 config/dev.yml 解析路由前缀与公开路由。
//
// 走与运行时一致的读取方式：先展开 ${VAR:default} 占位符再交给 viper 解析。
func readPublicRoutes(t *testing.T) (string, map[string]struct{}) {
	t.Helper()

	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("获取测试文件路径失败")
	}
	configPath := filepath.Join(filepath.Dir(file), "..", "..", "config", "dev.yml")

	raw, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("读取开发配置失败: %v", err)
	}

	v := viper.New()
	v.SetConfigType("yml")
	if err := v.ReadConfig(strings.NewReader(coreconfig.ExpandEnvDefaults(string(raw)))); err != nil {
		t.Fatalf("解析开发配置失败: %v", err)
	}

	var bizCfg svcconfig.BootConfig
	if err := v.UnmarshalKey(svcconfig.SectionKey, &bizCfg); err != nil {
		t.Fatalf("解析 biz 配置段失败: %v", err)
	}
	// router_prefix 未配置时为空串：路由的 /api 前缀已写在各 handler 的 URI 里
	prefix := strings.TrimSpace(v.GetString("http.router_prefix"))

	routes := make(map[string]struct{}, len(bizCfg.Auth.PublicPaths))
	for _, route := range bizCfg.Auth.PublicRoutes() {
		routes[route.Method+" "+route.Path] = struct{}{}
	}
	if len(routes) == 0 {
		t.Fatal("biz.auth.public_paths 未解析到公开路由")
	}
	return prefix, routes
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

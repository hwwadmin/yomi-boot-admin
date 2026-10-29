package snapshot

import (
	"testing"
	"yomi-admin-backend/domains/rbac/internal/repo_contract"
	"yomi-admin-backend/domains/rbac/pkg/shared/entity"
	"yomi-admin-backend/domains/rbac/pkg/shared/enums"
)

// testGraph 构造覆盖超管、正常链路、停用角色、停用权限组的六表数据
func testGraph() *repo_contract.AuthGraph {
	return &repo_contract.AuthGraph{
		Roles: []*entity.Role{
			{ID: 1, Code: enums.RoleCodeSuperAdmin, Status: enums.StatusEnabled},
			{ID: 2, Code: "viewer", Status: enums.StatusEnabled},
			{ID: 3, Code: "auditor", Status: enums.StatusEnabled},
			{ID: 4, Code: "disabled_role", Status: enums.StatusDisabled},
		},
		Permissions: []*entity.Permission{
			{ID: 100, Code: "system:user", Name: "用户", Type: enums.PermissionTypeMenu, UIPath: "/system/user", Status: enums.StatusEnabled},
			{ID: 101, Code: "system:user:list", Type: enums.PermissionTypeAPI, HTTPMethod: "GET", HTTPPath: "/api/system/user", Status: enums.StatusEnabled},
			{ID: 102, Code: "system:user:create", Type: enums.PermissionTypeAPI, HTTPMethod: "POST", HTTPPath: "/api/system/user", Status: enums.StatusEnabled},
			{ID: 103, Code: "system:role", Name: "角色", Type: enums.PermissionTypeMenu, UIPath: "/system/role", Status: enums.StatusEnabled},
			{ID: 104, Code: "system:role:list", Type: enums.PermissionTypeAPI, HTTPMethod: "GET", HTTPPath: "/api/system/role", Status: enums.StatusEnabled},
			{ID: 105, Code: "system:hidden", Name: "隐藏", Type: enums.PermissionTypeMenu, UIPath: "/system/hidden", Status: enums.StatusDisabled},
			{ID: 106, Code: "system:hidden:list", Type: enums.PermissionTypeAPI, HTTPMethod: "GET", HTTPPath: "/api/system/hidden", Status: enums.StatusDisabled},
		},
		PermissionGroups: []*entity.PermissionGroup{
			{ID: 10, Code: "read_group", Status: enums.StatusEnabled},
			{ID: 11, Code: "write_group", Status: enums.StatusDisabled},
		},
		UserRoles: map[int64][]int64{
			1: {1},
			2: {2},
			3: {3},
			4: {4},
			7: {2, 3},
		},
		RoleGroups: map[int64][]int64{
			2: {10},
			3: {10, 11},
			4: {10},
		},
		GroupPermissions: map[int64][]int64{
			10: {100, 101, 103, 104, 105, 106},
			11: {102},
		},
	}
}

func newTestSnapshot() *Snapshot {
	return Build("v1", 1000, testGraph())
}

func TestSnapshotSuperAdmin(t *testing.T) {
	snap := newTestSnapshot()

	if !snap.CanAccessAPI(1, "GET", "/api/never-registered") {
		t.Fatal("超管应放行任意路由")
	}
	codes := snap.ListPermissionCodes(1)
	if len(codes) != 1 || codes[0] != "*" {
		t.Fatalf("超管权限编码应为 [*]: %v", codes)
	}
	menus := snap.ListMenus(1)
	if len(menus) != 2 || menus[0].ID != 100 || menus[1].ID != 103 {
		t.Fatalf("超管应看到全部启用菜单且按 ID 升序: %+v", menus)
	}
}

func TestSnapshotNormalUser(t *testing.T) {
	snap := newTestSnapshot()

	if !snap.CanAccessAPI(2, "GET", "/api/system/user") {
		t.Fatal("已授权接口应放行")
	}
	if snap.CanAccessAPI(2, "POST", "/api/system/user") {
		t.Fatal("停用权限组内的接口不应放行")
	}
	if snap.CanAccessAPI(2, "GET", "/api/system/hidden") {
		t.Fatal("停用权限不应放行")
	}

	codes := snap.ListPermissionCodes(2)
	want := []string{"system:role", "system:role:list", "system:user", "system:user:list"}
	if len(codes) != len(want) {
		t.Fatalf("权限编码数量不符: %v", codes)
	}
	for i, code := range want {
		if codes[i] != code {
			t.Fatalf("权限编码应按字母升序 %v，实际 %v", want, codes)
		}
	}

	menus := snap.ListMenus(2)
	if len(menus) != 2 || menus[0].ID != 100 || menus[1].ID != 103 {
		t.Fatalf("菜单不符: %+v", menus)
	}
}

func TestSnapshotFiltersDisabledChain(t *testing.T) {
	snap := newTestSnapshot()

	// 用户 3 的有效链路只剩启用权限组 10，停用权限组 11 提供的 POST 不应生效
	if !snap.CanAccessAPI(3, "GET", "/api/system/role") {
		t.Fatal("启用权限组内的接口应放行")
	}
	if snap.CanAccessAPI(3, "POST", "/api/system/user") {
		t.Fatal("停用权限组内的接口不应放行")
	}

	// 用户 4 绑定的角色已停用
	if snap.CanAccessAPI(4, "GET", "/api/system/user") {
		t.Fatal("停用角色不应放行")
	}
	if len(snap.ListMenus(4)) != 0 {
		t.Fatal("停用角色不应有菜单")
	}
}

func TestSnapshotDedupesAcrossRoles(t *testing.T) {
	snap := newTestSnapshot()

	if !snap.CanAccessAPI(7, "GET", "/api/system/user") || !snap.CanAccessAPI(7, "GET", "/api/system/role") {
		t.Fatal("多角色并集接口应放行")
	}

	menus := snap.ListMenus(7)
	if len(menus) != 2 {
		t.Fatalf("跨角色菜单应去重: %+v", menus)
	}
	codes := snap.ListPermissionCodes(7)
	if len(codes) != 4 {
		t.Fatalf("跨角色权限编码应去重: %v", codes)
	}
}

func TestSnapshotNormalizesMethodAndPath(t *testing.T) {
	snap := newTestSnapshot()

	if !snap.CanAccessAPI(2, " get ", " /api/system/user ") {
		t.Fatal("方法应大写归一、路径应去空白后匹配")
	}
	if snap.CanAccessAPI(2, "", "/api/system/user") || snap.CanAccessAPI(2, "GET", "") {
		t.Fatal("空方法或空路径不应放行")
	}
	if snap.CanAccessAPI(0, "GET", "/api/system/user") {
		t.Fatal("非法用户不应放行")
	}
	if len(snap.ListMenus(99)) != 0 || len(snap.ListPermissionCodes(99)) != 0 {
		t.Fatal("未绑定角色的用户不应有权限")
	}
}

func TestSnapshotEncodeDecodeRoundTrip(t *testing.T) {
	snap := newTestSnapshot()

	payload, err := Encode(snap)
	if err != nil {
		t.Fatalf("序列化失败: %v", err)
	}
	decoded, err := Decode(payload)
	if err != nil {
		t.Fatalf("反序列化失败: %v", err)
	}

	if decoded.Version != snap.Version || decoded.BuiltAt != snap.BuiltAt {
		t.Fatalf("元信息丢失: %+v", decoded)
	}
	if !decoded.CanAccessAPI(2, "GET", "/api/system/user") || decoded.CanAccessAPI(2, "POST", "/api/system/user") {
		t.Fatal("反序列化后求值结果应与原快照一致")
	}
	if len(decoded.ListMenus(2)) != 2 || len(decoded.ListPermissionCodes(2)) != 4 {
		t.Fatal("反序列化后菜单与权限编码应与原快照一致")
	}
	if decoded.RoleCount() != 4 || decoded.PermissionCount() != 7 || decoded.PermissionGroupCount() != 2 {
		t.Fatalf("行数统计不符: %d/%d/%d", decoded.RoleCount(), decoded.PermissionCount(), decoded.PermissionGroupCount())
	}
	// 用户角色 6 条、角色权限组 4 条、权限组权限 7 条
	if decoded.RelationCount() != 6+4+7 {
		t.Fatalf("关系条目统计不符: %d", decoded.RelationCount())
	}
}

func TestSnapshotDecodeRejectsOtherSchema(t *testing.T) {
	payload := []byte(`{"schema":999,"roles":[],"permissions":[],"permissionGroups":[]}`)
	if _, err := Decode(payload); err == nil {
		t.Fatal("结构版本不一致应返回错误")
	}
}

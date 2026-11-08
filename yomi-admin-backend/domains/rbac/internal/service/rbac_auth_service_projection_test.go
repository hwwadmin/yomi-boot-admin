package service

import (
	"context"
	"testing"
	"time"
	"yomi-admin-backend/domains/rbac/internal/projection"
	"yomi-admin-backend/domains/rbac/pkg/shared/entity"
	"yomi-admin-backend/domains/rbac/pkg/shared/enums"

	"github.com/hwwadmin/yomi-boot/component/common/distlock"
	"github.com/hwwadmin/yomi-boot/component/common/redis"
	"github.com/hwwadmin/yomi-boot/component/common/redis/redistest"

	redisv9 "github.com/redis/go-redis/v9"
)

// parityRoutes 是比对用的请求矩阵
var parityRoutes = []struct {
	method string
	path   string
}{
	{"GET", "/api/system/user"},
	{"POST", "/api/system/user"},
	{"GET", "/api/system/role"},
	{"GET", "/api/system/secret"},
	{"", "/api/system/user"},
}

// parityUsers 覆盖超管、单角色、多角色、停用角色、无角色与非法用户
var parityUsers = []int64{0, 1, 2, 3, 4, 7, 99}

// seedParityRepo 构造包含停用链路与跨组重复权限的数据集
func seedParityRepo() *fakeRepo {
	repo := newFakeRepo()
	repo.roles[1] = &entity.Role{ID: 1, Code: enums.RoleCodeSuperAdmin, Status: enums.StatusEnabled}
	repo.roles[2] = &entity.Role{ID: 2, Code: "viewer", Status: enums.StatusEnabled}
	repo.roles[3] = &entity.Role{ID: 3, Code: "auditor", Status: enums.StatusEnabled}
	repo.roles[4] = &entity.Role{ID: 4, Code: "disabled_role", Status: enums.StatusDisabled}

	repo.permissionGroups[10] = &entity.PermissionGroup{ID: 10, Code: "read_group", Status: enums.StatusEnabled}
	repo.permissionGroups[11] = &entity.PermissionGroup{ID: 11, Code: "write_group", Status: enums.StatusDisabled}

	repo.permissions[100] = &entity.Permission{ID: 100, Code: "system:user", Name: "用户", Type: enums.PermissionTypeMenu, UIPath: "/system/user", Status: enums.StatusEnabled}
	repo.permissions[101] = &entity.Permission{ID: 101, Code: "system:user:list", Type: enums.PermissionTypeAPI, HTTPMethod: "GET", HTTPPath: "/api/system/user", Status: enums.StatusEnabled}
	repo.permissions[102] = &entity.Permission{ID: 102, Code: "system:user:create", Type: enums.PermissionTypeAPI, HTTPMethod: "POST", HTTPPath: "/api/system/user", Status: enums.StatusEnabled}
	repo.permissions[103] = &entity.Permission{ID: 103, Code: "system:role", Name: "角色", Type: enums.PermissionTypeMenu, UIPath: "/system/role", Status: enums.StatusEnabled}
	repo.permissions[104] = &entity.Permission{ID: 104, Code: "system:role:list", Type: enums.PermissionTypeAPI, HTTPMethod: "GET", HTTPPath: "/api/system/role", Status: enums.StatusEnabled}
	repo.permissions[105] = &entity.Permission{ID: 105, Code: "system:role", Name: "角色", Type: enums.PermissionTypeMenu, UIPath: "/system/role", Status: enums.StatusEnabled}
	repo.permissions[106] = &entity.Permission{ID: 106, Code: "system:secret:list", Type: enums.PermissionTypeAPI, HTTPMethod: "GET", HTTPPath: "/api/system/secret", Status: enums.StatusDisabled}

	repo.userRoles[1] = []int64{1}
	repo.userRoles[2] = []int64{2}
	repo.userRoles[3] = []int64{3}
	repo.userRoles[4] = []int64{4}
	repo.userRoles[7] = []int64{2, 3}

	repo.roleGroups[2] = []int64{10}
	repo.roleGroups[3] = []int64{10, 11}
	repo.roleGroups[4] = []int64{10}

	// 105 与 103 是同一份菜单的重复登记，用于校验跨组去重
	repo.groupPermissions[10] = []int64{100, 101, 103, 104, 105, 106}
	repo.groupPermissions[11] = []int64{102}

	return repo
}

// newParityStore 创建以 fakeRepo 为数据源的投影存储
func newParityStore(t *testing.T, repo *fakeRepo) *projection.Store {
	t.Helper()

	client, cleanup := redistest.NewClient(t)
	t.Cleanup(cleanup)

	factory, err := distlock.NewFactory(client)
	if err != nil {
		t.Fatalf("创建分布式锁工厂失败: %v", err)
	}
	store, err := projection.NewStore(projection.Config{
		Key:      "test:rbac:parity",
		LockKey:  "test:rbac:parity:lock",
		MaxBytes: 1 << 20,
		LockWait: time.Second,
	}, redis.NewRedisHelper(client), factory, repo, parityTx{}, &parityIDs{})
	if err != nil {
		t.Fatalf("创建投影存储失败: %v", err)
	}
	return store
}

// parityTx 是透传事务执行器
type parityTx struct{}

func (parityTx) Run(ctx context.Context, fn func(context.Context) error) error { return fn(ctx) }
func (parityTx) InTx(context.Context) bool                                     { return false }
func (parityTx) RunAutonomous(ctx context.Context, _ time.Duration, fn func(context.Context) error) error {
	return fn(ctx)
}

// parityIDs 递增生成投影版本号
type parityIDs struct {
	next int64
}

func (g *parityIDs) NextID() (int64, error) {
	g.next++
	return g.next, nil
}

// redisClient 连接指定库号的 Redis。
func redisClient(t *testing.T, addr string, db int) *redisv9.Client {
	t.Helper()
	// 关闭重试并缩短拨号超时：Redis 故障场景应快速失败，避免拖慢用例。
	return redisv9.NewClient(&redisv9.Options{
		Addr:        addr,
		DB:          db,
		MaxRetries:  -1,
		DialTimeout: 200 * time.Millisecond,
	})
}

// TestAuthServiceProjectionMatchesSQL 校验投影求值与 SQL 直查的鉴权结果完全一致
func TestAuthServiceProjectionMatchesSQL(t *testing.T) {
	ctx := context.Background()
	repo := seedParityRepo()

	sqlService := NewRbacAuthService(repo, nil)
	projectionService := NewRbacAuthService(repo, newParityStore(t, repo))

	for _, userID := range parityUsers {
		for _, route := range parityRoutes {
			want, err := sqlService.CanAccessAPI(ctx, userID, route.method, route.path)
			if err != nil {
				t.Fatalf("SQL 路径求值失败: %v", err)
			}
			got, err := projectionService.CanAccessAPI(ctx, userID, route.method, route.path)
			if err != nil {
				t.Fatalf("投影路径求值失败: %v", err)
			}
			if got != want {
				t.Fatalf("鉴权结果不一致 user=%d %s %s: SQL=%v 投影=%v",
					userID, route.method, route.path, want, got)
			}
		}

		wantMenus, err := sqlService.ListMenus(ctx, userID)
		if err != nil {
			t.Fatalf("SQL 路径查询菜单失败: %v", err)
		}
		gotMenus, err := projectionService.ListMenus(ctx, userID)
		if err != nil {
			t.Fatalf("投影路径查询菜单失败: %v", err)
		}
		if !samePermissionIDs(wantMenus, gotMenus) {
			t.Fatalf("菜单不一致 user=%d: SQL=%v 投影=%v", userID, permissionIDs(wantMenus), permissionIDs(gotMenus))
		}

		wantCodes, err := sqlService.ListPermissionCodes(ctx, userID)
		if err != nil {
			t.Fatalf("SQL 路径查询权限编码失败: %v", err)
		}
		gotCodes, err := projectionService.ListPermissionCodes(ctx, userID)
		if err != nil {
			t.Fatalf("投影路径查询权限编码失败: %v", err)
		}
		if !sameStrings(wantCodes, gotCodes) {
			t.Fatalf("权限编码不一致 user=%d: SQL=%v 投影=%v", userID, wantCodes, gotCodes)
		}
	}
}

// TestAuthServiceFallsBackToSQLWhenRedisUnavailable 校验 Redis 故障时仍按 SQL 给出正确结果
func TestAuthServiceFallsBackToSQLWhenRedisUnavailable(t *testing.T) {
	ctx := context.Background()
	repo := seedParityRepo()

	addr, db, _ := redistest.NewServer(t)
	client := redisClient(t, addr, db)
	t.Cleanup(func() { _ = client.Close() })

	factory, err := distlock.NewFactory(client)
	if err != nil {
		t.Fatalf("创建分布式锁工厂失败: %v", err)
	}
	store, err := projection.NewStore(projection.Config{
		Key:      "test:rbac:fallback",
		LockKey:  "test:rbac:fallback:lock",
		MaxBytes: 1 << 20,
	}, redis.NewRedisHelper(client), factory, repo, parityTx{}, &parityIDs{})
	if err != nil {
		t.Fatalf("创建投影存储失败: %v", err)
	}

	svc := NewRbacAuthService(repo, store)
	if allowed, err := svc.CanAccessAPI(ctx, 2, "GET", "/api/system/user"); err != nil || !allowed {
		t.Fatalf("Redis 正常时应放行: allowed=%v err=%v", allowed, err)
	}

	// redistest 不再提供可关闭的进程内 Redis，关闭客户端模拟不可用。
	_ = client.Close()

	allowed, err := svc.CanAccessAPI(ctx, 2, "GET", "/api/system/user")
	if err != nil || !allowed {
		t.Fatalf("Redis 不可用时应降级 SQL 并放行: allowed=%v err=%v", allowed, err)
	}
	if allowed, err := svc.CanAccessAPI(ctx, 2, "POST", "/api/system/user"); err != nil || allowed {
		t.Fatalf("Redis 不可用时停用权限组仍应拒绝: allowed=%v err=%v", allowed, err)
	}
	if codes, err := svc.ListPermissionCodes(ctx, 2); err != nil || len(codes) != 4 {
		t.Fatalf("Redis 不可用时权限编码应来自 SQL: codes=%v err=%v", codes, err)
	}
}

// permissionIDs 提取权限 ID
func permissionIDs(permissions []*entity.Permission) []int64 {
	ids := make([]int64, 0, len(permissions))
	for _, permission := range permissions {
		ids = append(ids, permission.ID)
	}
	return ids
}

// samePermissionIDs 判断两组权限 ID 顺序与内容一致
func samePermissionIDs(left, right []*entity.Permission) bool {
	leftIDs, rightIDs := permissionIDs(left), permissionIDs(right)
	if len(leftIDs) != len(rightIDs) {
		return false
	}
	for i := range leftIDs {
		if leftIDs[i] != rightIDs[i] {
			return false
		}
	}
	return true
}

// sameStrings 判断两个字符串切片顺序与内容一致
func sameStrings(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for i := range left {
		if left[i] != right[i] {
			return false
		}
	}
	return true
}

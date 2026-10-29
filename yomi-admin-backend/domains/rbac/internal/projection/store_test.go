package projection

import (
	"context"
	"testing"
	"time"
	"yomi-admin-backend/domains/rbac/internal/repo_contract"
	"yomi-admin-backend/domains/rbac/pkg/shared/entity"
	"yomi-admin-backend/domains/rbac/pkg/shared/enums"

	"gitee.com/hq_project/yomi-boot/component/common/distlock"
	"gitee.com/hq_project/yomi-boot/component/common/redis"
	"gitee.com/hq_project/yomi-boot/component/common/redis/redistest"
	"gitee.com/hq_project/yomi-boot/component/core/errors"

	redisv9 "github.com/redis/go-redis/v9"
)

const (
	testKey     = "test:rbac:projection"
	testLockKey = "test:rbac:projection:lock"
)

// fakeLoader 记录调用次数并返回固定六表数据
type fakeLoader struct {
	calls int
	graph *repo_contract.AuthGraph
}

func (l *fakeLoader) LoadAuthGraph(context.Context) (*repo_contract.AuthGraph, error) {
	l.calls++
	return l.graph, nil
}

// fakeTx 是不带事务的透传执行器
type fakeTx struct{}

func (fakeTx) Run(ctx context.Context, fn func(context.Context) error) error {
	return fn(ctx)
}

func (fakeTx) InTx(context.Context) bool {
	return false
}

func (fakeTx) RunAutonomous(ctx context.Context, _ time.Duration, fn func(context.Context) error) error {
	return fn(ctx)
}

// fakeIDs 递增生成投影版本号
type fakeIDs struct {
	next int64
}

func (g *fakeIDs) NextID() (int64, error) {
	g.next++
	return g.next, nil
}

// testGraph 构造最小可求值数据：超管 + 一个普通角色
func testGraph() *repo_contract.AuthGraph {
	return &repo_contract.AuthGraph{
		Roles: []*entity.Role{
			{ID: 1, Code: enums.RoleCodeSuperAdmin, Status: enums.StatusEnabled},
			{ID: 2, Code: "viewer", Status: enums.StatusEnabled},
		},
		Permissions: []*entity.Permission{
			{ID: 100, Code: "system:user:list", Type: enums.PermissionTypeAPI, HTTPMethod: "GET", HTTPPath: "/api/system/user", Status: enums.StatusEnabled},
		},
		PermissionGroups: []*entity.PermissionGroup{
			{ID: 10, Code: "read_group", Status: enums.StatusEnabled},
		},
		UserRoles:        map[int64][]int64{1: {1}, 2: {2}},
		RoleGroups:       map[int64][]int64{2: {10}},
		GroupPermissions: map[int64][]int64{10: {100}},
	}
}

func testConfig() Config {
	return Config{
		Key:      testKey,
		LockKey:  testLockKey,
		MaxBytes: 1 << 20,
		LockWait: 200 * time.Millisecond,
	}
}

// newTestStore 创建基于内存 Redis 的投影存储
func newTestStore(t *testing.T, cfg Config, loader Loader) (*Store, *redisv9.Client) {
	t.Helper()

	client, cleanup := redistest.NewClient(t)
	t.Cleanup(cleanup)

	factory, err := distlock.NewFactory(client)
	if err != nil {
		t.Fatalf("创建分布式锁工厂失败: %v", err)
	}
	store, err := NewStore(cfg, redis.NewRedisHelper(client), factory, loader, fakeTx{}, &fakeIDs{})
	if err != nil {
		t.Fatalf("创建投影存储失败: %v", err)
	}
	return store, client
}

func TestStoreLoadBuildsOnMissAndReusesProjection(t *testing.T) {
	ctx := context.Background()
	loader := &fakeLoader{graph: testGraph()}
	store, client := newTestStore(t, testConfig(), loader)

	first, err := store.Load(ctx)
	if err != nil {
		t.Fatalf("首次读取失败: %v", err)
	}
	if !first.CanAccessAPI(2, "GET", "/api/system/user") {
		t.Fatal("首次构建的投影应可用于求值")
	}
	if loader.calls != 1 {
		t.Fatalf("首次读取应回源一次，实际 %d", loader.calls)
	}

	exists, err := client.Exists(ctx, testKey).Result()
	if err != nil || exists != 1 {
		t.Fatalf("投影应写入 Redis: exists=%d err=%v", exists, err)
	}

	second, err := store.Load(ctx)
	if err != nil {
		t.Fatalf("二次读取失败: %v", err)
	}
	if second.Version != first.Version {
		t.Fatalf("二次读取应命中同一投影版本: %s vs %s", second.Version, first.Version)
	}
	if loader.calls != 1 {
		t.Fatalf("命中后不应回源，实际回源 %d 次", loader.calls)
	}
}

func TestStoreRebuildsUnusableProjection(t *testing.T) {
	ctx := context.Background()
	loader := &fakeLoader{graph: testGraph()}
	store, client := newTestStore(t, testConfig(), loader)

	cases := map[string]string{
		"损坏载荷":   "{not-json",
		"结构版本不符": `{"schema":999,"roles":[],"permissions":[],"permissionGroups":[]}`,
	}
	for name, payload := range cases {
		t.Run(name, func(t *testing.T) {
			if err := client.Set(ctx, testKey, payload, 0).Err(); err != nil {
				t.Fatalf("写入测试载荷失败: %v", err)
			}
			before := loader.calls
			snap, err := store.Load(ctx)
			if err != nil {
				t.Fatalf("不可用投影应触发重建: %v", err)
			}
			if snap.Version == "" || loader.calls != before+1 {
				t.Fatalf("应重建投影并回源一次: version=%q calls=%d", snap.Version, loader.calls-before)
			}
		})
	}
}

func TestStoreRefreshWritesNewVersion(t *testing.T) {
	ctx := context.Background()
	loader := &fakeLoader{graph: testGraph()}
	store, _ := newTestStore(t, testConfig(), loader)

	first, err := store.Load(ctx)
	if err != nil {
		t.Fatalf("首次读取失败: %v", err)
	}
	refreshed, err := store.Refresh(ctx)
	if err != nil {
		t.Fatalf("刷新失败: %v", err)
	}
	if refreshed.Version == first.Version {
		t.Fatal("刷新应生成新的投影版本")
	}
	if refreshed.RoleCount() != 2 || refreshed.PermissionCount() != 1 || refreshed.PermissionGroupCount() != 1 {
		t.Fatalf("刷新结果行数不符: %+v", refreshed)
	}

	loaded, err := store.Load(ctx)
	if err != nil {
		t.Fatalf("刷新后读取失败: %v", err)
	}
	if loaded.Version != refreshed.Version {
		t.Fatalf("刷新后应读到新版本: %s vs %s", loaded.Version, refreshed.Version)
	}
}

func TestStoreRefreshBusyWhenLockHeld(t *testing.T) {
	ctx := context.Background()
	loader := &fakeLoader{graph: testGraph()}
	store, client := newTestStore(t, testConfig(), loader)

	factory, err := distlock.NewFactory(client)
	if err != nil {
		t.Fatalf("创建分布式锁工厂失败: %v", err)
	}
	held := factory.GetLock(testLockKey)
	locked, err := held.TryLock(ctx)
	if err != nil || !locked {
		t.Fatalf("占位锁失败: locked=%v err=%v", locked, err)
	}
	defer func() { _ = held.Unlock(ctx) }()

	if _, err := store.Refresh(ctx); !errors.Is(err, ErrBusy) {
		t.Fatalf("锁被占用时应返回 ErrBusy，实际 %v", err)
	}
	if _, err := store.Load(ctx); !errors.Is(err, ErrBusy) {
		t.Fatalf("锁被占用且无投影时读取应返回 ErrBusy，实际 %v", err)
	}
}

func TestStoreRefreshRejectsOversizedProjection(t *testing.T) {
	ctx := context.Background()
	cfg := testConfig()
	cfg.MaxBytes = 16
	store, _ := newTestStore(t, cfg, &fakeLoader{graph: testGraph()})

	if _, err := store.Refresh(ctx); !errors.Is(err, ErrTooLarge) {
		t.Fatalf("超限应返回 ErrTooLarge，实际 %v", err)
	}
}

func TestStoreLoadFailsWhenRedisUnavailable(t *testing.T) {
	addr, stopRedis := redistest.NewServer(t)
	// 关闭重试并缩短拨号超时：Redis 停机场景应快速失败，避免拖慢用例
	client := redisv9.NewClient(&redisv9.Options{
		Addr:        addr,
		MaxRetries:  -1,
		DialTimeout: 200 * time.Millisecond,
	})
	t.Cleanup(func() { _ = client.Close() })

	factory, err := distlock.NewFactory(client)
	if err != nil {
		t.Fatalf("创建分布式锁工厂失败: %v", err)
	}
	store, err := NewStore(testConfig(), redis.NewRedisHelper(client), factory,
		&fakeLoader{graph: testGraph()}, fakeTx{}, &fakeIDs{})
	if err != nil {
		t.Fatalf("创建投影存储失败: %v", err)
	}

	stopRedis()

	if _, err := store.Load(context.Background()); err == nil {
		t.Fatal("Redis 不可用时应返回错误，由调用方降级")
	}
}

func TestNewStoreRejectsInvalidConfig(t *testing.T) {
	cases := map[string]Config{
		"缺少投影键":   {LockKey: testLockKey, MaxBytes: 1},
		"缺少刷新锁键":  {Key: testKey, MaxBytes: 1},
		"体积上限非正数": {Key: testKey, LockKey: testLockKey},
	}
	for name, cfg := range cases {
		t.Run(name, func(t *testing.T) {
			client, cleanup := redistest.NewClient(t)
			defer cleanup()
			factory, _ := distlock.NewFactory(client)
			if _, err := NewStore(cfg, redis.NewRedisHelper(client), factory,
				&fakeLoader{graph: testGraph()}, fakeTx{}, &fakeIDs{}); err == nil {
				t.Fatal("非法配置应返回错误")
			}
		})
	}
}

func TestStoreWarmUpReusesExistingProjection(t *testing.T) {
	ctx := context.Background()
	loader := &fakeLoader{graph: testGraph()}
	store, client := newTestStore(t, testConfig(), loader)

	// 投影缺失：预热负责构建
	if err := store.WarmUp(ctx); err != nil {
		t.Fatalf("首次预热失败: %v", err)
	}
	if loader.calls != 1 {
		t.Fatalf("投影缺失时应构建一次，实际回源 %d 次", loader.calls)
	}
	firstVersion, _ := client.Get(ctx, testKey).Result()

	// 投影已存在：预热不再重建，滚动重启不会覆盖已有数据
	if err := store.WarmUp(ctx); err != nil {
		t.Fatalf("二次预热失败: %v", err)
	}
	if loader.calls != 1 {
		t.Fatalf("投影可用时不应重建，实际回源 %d 次", loader.calls)
	}
	raw, _ := client.Get(ctx, testKey).Result()
	if raw != firstVersion {
		t.Fatal("预热不应改写已有投影")
	}
}

func TestStoreWarmUpRebuildsUnusableProjection(t *testing.T) {
	ctx := context.Background()
	loader := &fakeLoader{graph: testGraph()}
	store, client := newTestStore(t, testConfig(), loader)

	// 结构版本不符的投影视为不可用：预热应重建
	if err := client.Set(ctx, testKey, `{"schema":999,"roles":[],"permissions":[],"permissionGroups":[]}`, 0).Err(); err != nil {
		t.Fatalf("写入测试载荷失败: %v", err)
	}
	if err := store.WarmUp(ctx); err != nil {
		t.Fatalf("预热失败: %v", err)
	}
	if loader.calls != 1 {
		t.Fatalf("不可用投影应触发重建，实际回源 %d 次", loader.calls)
	}

	snap, err := store.Load(ctx)
	if err != nil || snap.Version == "" {
		t.Fatalf("重建后应能读取投影: err=%v", err)
	}
}

func TestStoreLoadServesProjectionUntilRefresh(t *testing.T) {
	ctx := context.Background()
	loader := &fakeLoader{graph: testGraph()}
	store, _ := newTestStore(t, testConfig(), loader)

	if _, err := store.Load(ctx); err != nil {
		t.Fatalf("构建投影失败: %v", err)
	}

	// 模拟库里新增角色：没有刷新时读到的仍是旧投影
	loader.graph.Roles = append(loader.graph.Roles, &entity.Role{ID: 3, Code: "editor", Status: enums.StatusEnabled})
	loader.graph.UserRoles[3] = []int64{3}
	loader.graph.RoleGroups[3] = []int64{10}

	cached, err := store.Load(ctx)
	if err != nil {
		t.Fatalf("刷新前读取失败: %v", err)
	}
	if cached.RoleCount() != 2 {
		t.Fatalf("刷新前应命中旧投影，实际角色数 %d", cached.RoleCount())
	}

	refreshed, err := store.Refresh(ctx)
	if err != nil {
		t.Fatalf("刷新失败: %v", err)
	}
	if refreshed.RoleCount() != 3 {
		t.Fatalf("刷新后应读到新数据，实际角色数 %d", refreshed.RoleCount())
	}
}

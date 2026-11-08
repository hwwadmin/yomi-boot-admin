package projection

import (
	"context"
	"testing"
	"time"
	"yomi-admin-backend/domains/rbac/internal/repo_contract"
	"yomi-admin-backend/domains/rbac/internal/snapshot"
	"yomi-admin-backend/domains/rbac/pkg/shared/entity"
	"yomi-admin-backend/domains/rbac/pkg/shared/enums"

	"github.com/hwwadmin/yomi-boot/component/common/distlock"
	"github.com/hwwadmin/yomi-boot/component/common/redis"
	"github.com/hwwadmin/yomi-boot/component/common/redis/redistest"
	"github.com/hwwadmin/yomi-boot/component/core/errors"

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

// memoryConfig 在测试配置上打开内存层（5s 周期）
func memoryConfig() Config {
	cfg := testConfig()
	cfg.MemorySync = 5 * time.Second
	return cfg
}

// encodeGraph 将六表数据编码为 Redis 载荷
func encodeGraph(t *testing.T, version string, graph *repo_contract.AuthGraph) string {
	t.Helper()
	payload, err := snapshot.Encode(snapshot.Build(version, time.Now().UnixMilli(), graph))
	if err != nil {
		t.Fatalf("编码投影失败: %v", err)
	}
	return string(payload)
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
	addr, db, _ := redistest.NewServer(t)
	// 关闭重试并缩短拨号超时：Redis 故障场景应快速失败，避免拖慢用例。
	// redistest 的清理只归还库号，不停止容器，关闭客户端让后续命令失败。
	client := redisv9.NewClient(&redisv9.Options{
		Addr:        addr,
		DB:          db,
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

	_ = client.Close()

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

// ———————— 内存快照层 ————————

// TestStoreMemoryLayerServesWithoutRedis 校验内存命中后不再访问 Redis
func TestStoreMemoryLayerServesWithoutRedis(t *testing.T) {
	ctx := context.Background()
	loader := &fakeLoader{graph: testGraph()}
	store, client := newTestStore(t, memoryConfig(), loader)

	if _, err := store.Refresh(ctx); err != nil {
		t.Fatalf("构建投影失败: %v", err)
	}
	first := store.mem.Load()
	if first == nil {
		t.Fatal("刷新后内存应持有快照")
	}

	// 删除 Redis 键：内存仍可服务，鉴权不再依赖 Redis
	if err := client.Del(ctx, testKey).Err(); err != nil {
		t.Fatalf("删除 Redis 键失败: %v", err)
	}
	snap, err := store.Load(ctx)
	if err != nil || snap.Version != first.Version {
		t.Fatalf("内存命中应直接返回: version=%s err=%v", snap.Version, err)
	}
}

// TestStoreMemoryFallbackDoesNotBackfill 校验内存未就绪时降级读 Redis，且不回填内存
func TestStoreMemoryFallbackDoesNotBackfill(t *testing.T) {
	ctx := context.Background()
	loader := &fakeLoader{graph: testGraph()}
	store, client := newTestStore(t, memoryConfig(), loader)

	if err := client.Set(ctx, testKey, encodeGraph(t, "seed-v1", testGraph()), 0).Err(); err != nil {
		t.Fatalf("写入投影失败: %v", err)
	}
	snap, err := store.Load(ctx)
	if err != nil || snap.Version != "seed-v1" {
		t.Fatalf("降级读应命中 Redis: version=%s err=%v", snap.Version, err)
	}
	if store.mem.Load() != nil {
		t.Fatal("降级读不应回填内存")
	}

	if err := store.syncMemoryOnce(ctx); err != nil {
		t.Fatalf("同步内存失败: %v", err)
	}
	if got := store.mem.Load(); got == nil || got.Version != "seed-v1" {
		t.Fatalf("同步后内存应持有投影: %+v", got)
	}
}

// TestStoreMemorySyncPicksUpRedisChanges 校验同步循环把 Redis 的新投影切换进内存
func TestStoreMemorySyncPicksUpRedisChanges(t *testing.T) {
	ctx := context.Background()
	loader := &fakeLoader{graph: testGraph()}
	store, client := newTestStore(t, memoryConfig(), loader)

	if _, err := store.Refresh(ctx); err != nil {
		t.Fatalf("构建投影失败: %v", err)
	}
	before := store.mem.Load()
	if before == nil {
		t.Fatal("刷新后内存应持有快照")
	}

	// 模拟其它实例刷新：直接改写 Redis 载荷
	changed := testGraph()
	changed.Roles = append(changed.Roles, &entity.Role{ID: 3, Code: "editor", Status: enums.StatusEnabled})
	changed.UserRoles[3] = []int64{3}
	changed.RoleGroups[3] = []int64{10}
	if err := client.Set(ctx, testKey, encodeGraph(t, "peer-v2", changed), 0).Err(); err != nil {
		t.Fatalf("改写投影失败: %v", err)
	}

	if err := store.syncMemoryOnce(ctx); err != nil {
		t.Fatalf("同步内存失败: %v", err)
	}
	loaded, err := store.Load(ctx)
	if err != nil || loaded.Version != "peer-v2" {
		t.Fatalf("同步后应读到新投影: version=%s err=%v", loaded.Version, err)
	}
	if store.mem.Load() == before {
		t.Fatal("同步应切换内存指针")
	}
	if !loaded.CanAccessAPI(3, "GET", "/api/system/user") {
		t.Fatal("新投影的求值索引应已预构建")
	}
}

// TestStoreMemorySyncKeepsOldOnRedisFailure 校验 Redis 不可用时同步失败但旧快照继续服务
func TestStoreMemorySyncKeepsOldOnRedisFailure(t *testing.T) {
	ctx := context.Background()
	addr, db, _ := redistest.NewServer(t)
	client := redisv9.NewClient(&redisv9.Options{
		Addr:        addr,
		DB:          db,
		MaxRetries:  -1,
		DialTimeout: 200 * time.Millisecond,
	})
	t.Cleanup(func() { _ = client.Close() })
	factory, err := distlock.NewFactory(client)
	if err != nil {
		t.Fatalf("创建分布式锁工厂失败: %v", err)
	}
	store, err := NewStore(memoryConfig(), redis.NewRedisHelper(client), factory,
		&fakeLoader{graph: testGraph()}, fakeTx{}, &fakeIDs{})
	if err != nil {
		t.Fatalf("创建投影存储失败: %v", err)
	}

	if _, err := store.Refresh(ctx); err != nil {
		t.Fatalf("构建投影失败: %v", err)
	}
	before := store.mem.Load()
	if before == nil {
		t.Fatal("刷新后内存应持有快照")
	}

	_ = client.Close()
	if err := store.syncMemoryOnce(ctx); err == nil {
		t.Fatal("Redis 不可用时同步应失败")
	}
	if store.mem.Load() != before {
		t.Fatal("同步失败应保留旧快照")
	}
	if snap, err := store.Load(ctx); err != nil || snap.Version != before.Version {
		t.Fatalf("内存快照应继续服务: version=%s err=%v", snap.Version, err)
	}
}

// TestStoreMemorySyncKeepsOldOnCorruptPayload 校验载荷损坏时保留旧快照且不触发重建
func TestStoreMemorySyncKeepsOldOnCorruptPayload(t *testing.T) {
	ctx := context.Background()
	loader := &fakeLoader{graph: testGraph()}
	store, client := newTestStore(t, memoryConfig(), loader)

	if _, err := store.Refresh(ctx); err != nil {
		t.Fatalf("构建投影失败: %v", err)
	}
	before := store.mem.Load()
	if err := client.Set(ctx, testKey, "{not-json", 0).Err(); err != nil {
		t.Fatalf("写入损坏载荷失败: %v", err)
	}

	if err := store.syncMemoryOnce(ctx); err != nil {
		t.Fatalf("损坏载荷应跳过同步而非报错: %v", err)
	}
	if store.mem.Load() != before {
		t.Fatal("载荷损坏应保留旧快照")
	}
	if loader.calls != 1 {
		t.Fatalf("同步不应触发重建，实际回源 %d 次", loader.calls)
	}
}

// TestStoreDisabledMemoryLayerReadsRedis 校验内存层关闭时行为与纯 Redis 模式一致
func TestStoreDisabledMemoryLayerReadsRedis(t *testing.T) {
	ctx := context.Background()
	loader := &fakeLoader{graph: testGraph()}
	store, _ := newTestStore(t, testConfig(), loader)

	if _, err := store.Refresh(ctx); err != nil {
		t.Fatalf("构建投影失败: %v", err)
	}
	if store.mem.Load() != nil {
		t.Fatal("内存层关闭时不应写内存")
	}
	if _, err := store.Load(ctx); err != nil {
		t.Fatalf("读取失败: %v", err)
	}
	if store.mem.Load() != nil {
		t.Fatal("内存层关闭时 Load 不应写内存")
	}
}

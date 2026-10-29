// Package projection 维护 RBAC 授权投影：Redis 中的全量快照，供鉴权热路径内存求值。
//
// 缓存分两层：进程内保留一份最近从 Redis 读到的快照（按 MemorySync 周期切换指针，
// 读路径命中后不再访问 Redis），Redis 是唯一事实源；任何一层不可用都向上返回错误，
// 由调用方降级到 SQL 直查。
package projection

import (
	"context"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	"yomi-admin-backend/domains/rbac/internal/repo_contract"
	"yomi-admin-backend/domains/rbac/internal/snapshot"

	"gitee.com/hq_project/yomi-boot/component/common/db/tx"
	"gitee.com/hq_project/yomi-boot/component/common/distlock"
	"gitee.com/hq_project/yomi-boot/component/common/redis"
	"gitee.com/hq_project/yomi-boot/component/common/snowflake"
	"gitee.com/hq_project/yomi-boot/component/core/errors"
	"gitee.com/hq_project/yomi-boot/component/core/logger"
)

const (
	// lockTTL 覆盖一次六表全量构建与写回的耗时上限，超时自动释放。
	//
	// 必须明显大于一次构建的耗时：构建期间锁一旦过期，其它实例就会进来重复重建，
	// 先写入的新快照还可能被后写入的旧快照覆盖。代价是进程异常退出后残留锁最多
	// 压制一个 TTL 的重建，这期间读请求按未命中降级为 SQL 直查，不影响正确性。
	lockTTL = 30 * time.Second
	// unlockTimeout 是释放锁的最长等待时间
	unlockTimeout = 3 * time.Second
	// fallbackWarnInterval 是降级告警的最小间隔，避免 Redis 故障期间刷日志
	fallbackWarnInterval = 30 * time.Second
)

// ErrBusy 表示投影正在被本实例或其它实例重建，调用方应降级到 SQL 直查
var ErrBusy = errors.New("rbac projection: 投影重建进行中")

// ErrTooLarge 表示序列化后的投影超过配置的体积上限
var ErrTooLarge = errors.New("rbac projection: 投影体积超过上限")

// Config 是投影运行参数
type Config struct {
	Key             string        // 投影键
	LockKey         string        // 刷新锁键
	MaxBytes        int           // 投影载荷体积上限
	LockWait        time.Duration // 获取刷新锁的最长等待
	BootstrapWarmup bool          // 启动时投影不可用才构建
	RefreshInterval time.Duration // 定时刷新间隔，<=0 表示关闭
	MemorySync      time.Duration // 内存快照同步周期；<=0 表示关闭内存层，热路径直接读 Redis
}

// Loader 提供六表全量数据
type Loader interface {
	// LoadAuthGraph 在调用方事务内读取六表全量数据
	LoadAuthGraph(ctx context.Context) (*repo_contract.AuthGraph, error)
}

// Store 负责投影的读取、重建与刷新
type Store struct {
	cfg    Config
	redis  redis.RedisService
	locks  distlock.DistLockFactory
	loader Loader
	tx     tx.Executor
	ids    snowflake.Generator
	now    func() time.Time

	// mem 缓存最近一份从 Redis 读到的快照：热路径命中后不再访问 Redis。
	// 写入方只有同步循环与 buildAndStore，指针切换前后都是完整可用的快照；nil 表示内存层未就绪。
	mem atomic.Pointer[snapshot.Snapshot]

	// mu 保证本实例同一时刻只有一个重建在跑；用 TryLock 而非 Lock，竞争期间直接降级
	mu sync.Mutex
	// lastFallbackWarn 记录上次降级告警的时间（Unix 纳秒）
	lastFallbackWarn atomic.Int64
}

// NewStore 创建投影存储
func NewStore(cfg Config, rds redis.RedisService, locks distlock.DistLockFactory,
	loader Loader, executor tx.Executor, ids snowflake.Generator,
) (*Store, error) {
	normalized, err := cfg.withDefaults()
	if err != nil {
		return nil, err
	}
	if rds == nil || locks == nil || loader == nil || executor == nil || ids == nil {
		return nil, errors.New("rbac projection: 依赖不能为空")
	}
	return &Store{
		cfg:    normalized,
		redis:  rds,
		locks:  locks,
		loader: loader,
		tx:     executor,
		ids:    ids,
		now:    time.Now,
	}, nil
}

// Load 读取投影；优先命中内存，未就绪时读 Redis，缺失时尝试重建，重建不成功则返回错误由调用方降级。
//
// 内存未就绪时的 Redis 读取是一次性降级：只读不回填内存、不触发同步，内存由同步循环按周期补齐。
func (s *Store) Load(ctx context.Context) (*snapshot.Snapshot, error) {
	if snap := s.mem.Load(); snap != nil {
		return snap, nil
	}

	snap, hit, err := s.read(ctx)
	if err != nil {
		return nil, err
	}
	if hit {
		return snap, nil
	}

	// 本实例已有重建在跑：本次直接降级，不排队等待
	if !s.mu.TryLock() {
		return nil, ErrBusy
	}
	defer s.mu.Unlock()

	// 等锁期间可能已被其它请求或实例写入
	if snap, hit, err = s.read(ctx); err == nil && hit {
		return snap, nil
	}

	lock := s.locks.GetLock(s.cfg.LockKey, distlock.WithTTL(lockTTL))
	acquired, err := lock.TryLockWithTimeout(ctx, s.cfg.LockWait)
	if err != nil {
		s.warnFallback(ctx, "获取 RBAC 投影刷新锁失败", err)
		return nil, err
	}
	if !acquired {
		return nil, ErrBusy
	}
	defer s.releaseLock(ctx, lock)

	return s.buildAndStore(ctx)
}

// Refresh 强制全量重建投影，供手动刷新与定时刷新使用
func (s *Store) Refresh(ctx context.Context) (*snapshot.Snapshot, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	lock := s.locks.GetLock(s.cfg.LockKey, distlock.WithTTL(lockTTL))
	acquired, err := lock.TryLockWithTimeout(ctx, s.cfg.LockWait)
	if err != nil {
		return nil, errors.Wrap(err, "获取 RBAC 投影刷新锁失败")
	}
	if !acquired {
		return nil, ErrBusy
	}
	defer s.releaseLock(ctx, lock)

	return s.buildAndStore(ctx)
}

// StartBackground 启动预热、内存同步与定时刷新；ctx 取消后全部退出
func (s *Store) StartBackground(ctx context.Context) {
	if s.cfg.BootstrapWarmup {
		go func() {
			// 多实例或滚动重启时，已有可用投影的实例不重复构建
			if err := s.WarmUp(ctx); err != nil && !errors.Is(err, ErrBusy) {
				logger.Warn(ctx, "[RBAC] 启动期预热授权投影失败", "error", err)
			}
		}()
	}
	if s.cfg.MemorySync > 0 {
		go func() {
			// 先同步一次：滚动重启后第一个请求即可命中内存，失败由循环按周期重试
			if err := s.syncMemoryOnce(ctx); err != nil {
				logger.Warn(ctx, "[RBAC] 启动期同步内存投影失败", "error", err)
			}
			s.memLoop(ctx)
		}()
	}
	if s.cfg.RefreshInterval <= 0 {
		return
	}
	logger.Info(ctx, "[RBAC] 定时刷新授权投影已启用", "interval", s.cfg.RefreshInterval.String())
	go s.refreshLoop(ctx)
}

// memLoop 按固定周期把 Redis 中的投影切换进内存；失败保留旧快照，由下个周期重试
func (s *Store) memLoop(ctx context.Context) {
	ticker := time.NewTicker(s.cfg.MemorySync)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// 失败已在内部限流告警，保留旧快照继续服务
			_ = s.syncMemoryOnce(ctx)
		}
	}
}

// syncMemoryOnce 从 Redis 读取一份投影并切换进内存，只读不重建：
// Redis 缺失或载荷损坏都不在这里抢锁修复，交给预热与定时刷新链路。
func (s *Store) syncMemoryOnce(ctx context.Context) error {
	raw, exists, err := s.redis.OpsStr4Get(ctx, s.cfg.Key)
	if err != nil {
		s.warnFallback(ctx, "同步内存投影失败", err)
		return err
	}
	if !exists {
		return nil
	}

	snap, err := snapshot.Decode([]byte(raw))
	if err != nil {
		// 载荷不可用：保留旧内存，等刷新链路重写 Redis 后自动接上
		s.warnFallback(ctx, "内存投影同步跳过：Redis 载荷不可用", err)
		return nil
	}
	snap.PrebuildIndex()
	s.mem.Store(snap)
	return nil
}

// WarmUp 启动预热：投影可用时直接复用，缺失、损坏或结构版本不符时才从数据库重建。
// 已存在可用投影时不重建，避免重启覆盖掉别处刚刷新过的数据。
func (s *Store) WarmUp(ctx context.Context) error {
	existing, hit, err := s.read(ctx)
	if err != nil {
		return err
	}
	if hit {
		logger.Info(ctx, "[RBAC] 启动期发现可用授权投影，跳过重建",
			"key", s.cfg.Key,
			"version", existing.Version,
			"built_at", existing.BuiltAt)
		return nil
	}

	startedAt := s.now()
	built, err := s.Refresh(ctx)
	if err != nil {
		return err
	}
	logger.Info(ctx, "[RBAC] 启动期授权投影已构建",
		"version", built.Version,
		"roles", built.RoleCount(),
		"permissions", built.PermissionCount(),
		"permission_groups", built.PermissionGroupCount(),
		"duration_ms", s.now().Sub(startedAt).Milliseconds())
	return nil
}

// refreshLoop 按固定间隔刷新投影；抢不到锁说明其它实例正在刷新，本轮跳过
func (s *Store) refreshLoop(ctx context.Context) {
	ticker := time.NewTicker(s.cfg.RefreshInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if _, err := s.Refresh(ctx); err != nil && !errors.Is(err, ErrBusy) {
				logger.Warn(ctx, "[RBAC] 定时刷新授权投影失败", "error", err)
			}
		}
	}
}

// read 读取并解析投影；未命中或载荷不可用时返回 hit=false
func (s *Store) read(ctx context.Context) (*snapshot.Snapshot, bool, error) {
	raw, exists, err := s.redis.OpsStr4Get(ctx, s.cfg.Key)
	if err != nil {
		s.warnFallback(ctx, "读取 RBAC 授权投影失败", err)
		return nil, false, err
	}
	if !exists {
		return nil, false, nil
	}

	snap, err := snapshot.Decode([]byte(raw))
	if err != nil {
		// 结构不兼容或内容损坏：按未命中处理，由重建覆盖
		s.warnFallback(ctx, "RBAC 授权投影不可用，将重建", err)
		return nil, false, nil
	}
	return snap, true, nil
}

// buildAndStore 从数据库全量构建投影并写回 Redis；调用方须持有刷新锁
func (s *Store) buildAndStore(ctx context.Context) (*snapshot.Snapshot, error) {
	graph, err := s.loadGraph(ctx)
	if err != nil {
		return nil, err
	}
	if graph == nil {
		return nil, errors.New("rbac projection: 未读取到鉴权数据")
	}

	id, err := s.ids.NextID()
	if err != nil {
		return nil, errors.Wrap(err, "生成 RBAC 投影版本号失败")
	}
	snap := snapshot.Build(strconv.FormatInt(int64(id), 10), s.now().UnixMilli(), graph)

	payload, err := snapshot.Encode(snap)
	if err != nil {
		return nil, err
	}
	if len(payload) > s.cfg.MaxBytes {
		return nil, errors.Wrap(ErrTooLarge, "投影 %d 字节，上限 %d 字节", len(payload), s.cfg.MaxBytes)
	}
	if err := s.redis.OpsStr4Set(ctx, s.cfg.Key, string(payload), 0); err != nil {
		return nil, err
	}
	// 新投影即刻切换进内存：手动刷新在本实例立即生效，其它实例随同步循环（≤ 一个同步周期）跟进
	if s.cfg.MemorySync > 0 {
		snap.PrebuildIndex()
		s.mem.Store(snap)
	}
	return snap, nil
}

// loadGraph 在事务内读取六表，保证投影来自同一数据库快照
func (s *Store) loadGraph(ctx context.Context) (*repo_contract.AuthGraph, error) {
	var graph *repo_contract.AuthGraph
	err := s.tx.Run(ctx, func(txCtx context.Context) error {
		loaded, err := s.loader.LoadAuthGraph(txCtx)
		if err != nil {
			return err
		}
		graph = loaded
		return nil
	})
	if err != nil {
		return nil, err
	}
	return graph, nil
}

// releaseLock 以独立超时上下文释放锁，避免调用方 ctx 已取消时锁悬挂
func (s *Store) releaseLock(ctx context.Context, lock distlock.Lock) {
	unlockCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), unlockTimeout)
	defer cancel()
	if err := lock.Unlock(unlockCtx); err != nil {
		logger.Warn(ctx, "[RBAC] 释放投影刷新锁失败", "key", lock.Key(), "error", err)
	}
}

// warnFallback 按固定间隔输出降级告警，防止故障期间日志被单个请求刷满
func (s *Store) warnFallback(ctx context.Context, msg string, err error) {
	now := s.now().UnixNano()
	last := s.lastFallbackWarn.Load()
	if now-last < int64(fallbackWarnInterval) {
		return
	}
	if !s.lastFallbackWarn.CompareAndSwap(last, now) {
		return
	}
	logger.Warn(ctx, "[RBAC] "+msg, "key", s.cfg.Key, "error", err)
}

// withDefaults 校验并补全配置
func (c Config) withDefaults() (Config, error) {
	c.Key = strings.TrimSpace(c.Key)
	c.LockKey = strings.TrimSpace(c.LockKey)
	if c.Key == "" {
		return Config{}, errors.New("rbac projection: 投影键不能为空")
	}
	if c.LockKey == "" {
		return Config{}, errors.New("rbac projection: 刷新锁键不能为空")
	}
	if c.MaxBytes <= 0 {
		return Config{}, errors.New("rbac projection: 投影体积上限必须大于 0")
	}
	if c.LockWait <= 0 {
		c.LockWait = 2 * time.Second
	}
	if c.MemorySync < 0 {
		return Config{}, errors.New("rbac projection: 内存同步周期不能为负数")
	}
	if c.RefreshInterval < 0 {
		c.RefreshInterval = 0
	}
	return c, nil
}

package audit

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
	"yomi-admin-backend/domains/auditlog/pkg/shared/enums"
	auditloginput "yomi-admin-backend/domains/auditlog/pkg/shared/input"
	auditlogcontract "yomi-admin-backend/domains/auditlog/service_contract"

	"gitee.com/hq_project/yomi-boot/component/common/snowflake"
	"gitee.com/hq_project/yomi-boot/component/core/logger"
)

const (
	// defaultQueueSize 是内存队列容量；审计事件按"次"产生（登录、改密），1024 足够吸收突发
	defaultQueueSize = 1024
	// writeTimeout 是单条审计写入上限，避免慢库把消费协程卡死
	writeTimeout = 500 * time.Millisecond
	// drainTimeout 是退出时冲刷剩余队列的最长时间，超时即丢弃
	drainTimeout = 2 * time.Second
	// dropLogInterval 是队列满时的日志采样间隔，避免把日志刷爆
	dropLogInterval = 100
)

// TwoPhaseRecorder 是 Recorder 的两段式实现：业务前投递 STARTED，业务后投递 FINISHED。
//
// 落库是"内存队列 + 单消费协程"的异步投递：业务侧只做非阻塞入队，队列满就丢弃，
// 全程不等数据库。审计允许少量丢失（审计日志的常规取舍），换来业务零阻塞；
// 消费协程用独立 context 写库，不参与业务事务，也不受请求取消影响。
type TwoPhaseRecorder struct {
	writer           auditlogcontract.AuditLogWriter
	ids              snowflake.Generator
	tasks            chan writeTask
	dropped          atomic.Int64
	started          atomic.Bool
	notStartedWarned atomic.Bool
	startOnce        sync.Once
	done             chan struct{}
}

// writeTask 是一条待落库的审计记录；值拷贝，投递后不再依赖请求上下文
type writeTask struct {
	ref        Ref
	spec       Spec
	phase      enums.Phase
	result     enums.Result
	bizCode    string
	createTime int64
}

var _ Recorder = (*TwoPhaseRecorder)(nil)

// NewTwoPhaseRecorder 创建两段式审计写入器
func NewTwoPhaseRecorder(writer auditlogcontract.AuditLogWriter, ids snowflake.Generator) *TwoPhaseRecorder {
	return newTwoPhaseRecorder(writer, ids, defaultQueueSize)
}

// newTwoPhaseRecorder 按指定队列容量创建，供测试使用
func newTwoPhaseRecorder(writer auditlogcontract.AuditLogWriter, ids snowflake.Generator, queueSize int) *TwoPhaseRecorder {
	return &TwoPhaseRecorder{
		writer: writer,
		ids:    ids,
		tasks:  make(chan writeTask, queueSize),
		done:   make(chan struct{}),
	}
}

// StartBackground 启动消费协程；ctx 结束时冲刷队列并退出。必须在装配处调用一次（重复调用为空操作）。
func (r *TwoPhaseRecorder) StartBackground(ctx context.Context) {
	if r == nil {
		return
	}
	r.startOnce.Do(func() {
		r.started.Store(true)
		logger.Info(context.Background(), "审计队列已启动", "queue", cap(r.tasks))
		go r.consume(ctx)
	})
}

// Start 生成 audit_id 并投递 STARTED 记录；投递不阻塞，队列满则丢弃。
//
// 雪花 ID 生成失败时返回空 Ref，Finish 会因此跳过，整对记录一起放弃。
func (r *TwoPhaseRecorder) Start(ctx context.Context, spec Spec) Ref {
	if r == nil || r.writer == nil || r.ids == nil {
		return Ref{}
	}

	auditID, err := r.ids.NextID()
	if err != nil {
		logger.Warn(ctx, "生成审计事件ID失败", "event", spec.Event, "error", err)
		return Ref{}
	}

	ref := Ref{AuditID: auditID}
	r.enqueue(writeTask{
		ref:        ref,
		spec:       spec,
		phase:      enums.PhaseStarted,
		result:     enums.ResultPending,
		createTime: time.Now().UnixMilli(),
	})
	return ref
}

// Finish 用同一个 audit_id 投递 FINISHED 记录；ref 为空（ID 生成失败）时直接跳过
func (r *TwoPhaseRecorder) Finish(ctx context.Context, ref Ref, spec Spec, outcome Outcome) {
	if r == nil || r.writer == nil || ref.AuditID <= 0 {
		return
	}
	r.enqueue(writeTask{
		ref:        ref,
		spec:       spec,
		phase:      enums.PhaseFinished,
		result:     outcome.Result,
		bizCode:    outcome.BizCode,
		createTime: time.Now().UnixMilli(),
	})
}

// enqueue 非阻塞入队；队列满直接丢弃并采样告警
func (r *TwoPhaseRecorder) enqueue(task writeTask) {
	// 忘调 StartBackground 时没有消费协程，记录会静默堆积到队列满才告警；这里在第一条就喊一声
	if !r.started.Load() && r.notStartedWarned.CompareAndSwap(false, true) {
		logger.Warn(context.Background(), "审计消费协程未启动，记录不会落库，请检查装配是否调用了 StartBackground",
			"event", task.spec.Event)
	}

	select {
	case r.tasks <- task:
	default:
		dropped := r.dropped.Add(1)
		if dropped == 1 || dropped%dropLogInterval == 0 {
			logger.Warn(context.Background(), "审计队列已满，丢弃记录",
				"totalDropped", dropped, "event", task.spec.Event, "phase", task.phase)
		}
	}
}

// consume 串行消费队列：同一请求的 STARTED 先于 FINISHED 入队，落库顺序也一致
func (r *TwoPhaseRecorder) consume(ctx context.Context) {
	defer close(r.done)
	for {
		select {
		case task := <-r.tasks:
			r.deliver(task)
		case <-ctx.Done():
			r.drain()
			return
		}
	}
}

// drain 退出前把队列里剩下的记录写完，最多等 drainTimeout
func (r *TwoPhaseRecorder) drain() {
	timer := time.NewTimer(drainTimeout)
	defer timer.Stop()
	for {
		select {
		case task := <-r.tasks:
			r.deliver(task)
		case <-timer.C:
			logger.Warn(context.Background(), "审计队列冲刷超时，丢弃剩余记录", "totalDropped", r.dropped.Load())
			return
		default:
			return
		}
	}
}

// deliver 写一条审计记录；失败或 panic 只告警，不影响后续消费
func (r *TwoPhaseRecorder) deliver(task writeTask) {
	defer func() {
		if rec := recover(); rec != nil {
			logger.Warn(context.Background(), "审计写入 panic",
				"event", task.spec.Event, "auditId", task.ref.AuditID, "panic", rec)
		}
	}()

	writeCtx, cancel := context.WithTimeout(context.Background(), writeTimeout)
	defer cancel()

	err := r.writer.Record(writeCtx, &auditloginput.RecordInput{
		AuditID:      task.ref.AuditID,
		Phase:        task.phase,
		EventCode:    task.spec.Event,
		OperatorID:   task.spec.Operator.ID,
		OperatorName: task.spec.Operator.Name,
		ClientIP:     task.spec.ClientIP,
		TargetType:   task.spec.Target.Type,
		TargetID:     task.spec.Target.ID,
		Result:       task.result,
		BizCode:      task.bizCode,
		CreateTime:   task.createTime,
	})
	if err != nil {
		logger.Warn(writeCtx, "审计写入失败", "event", task.spec.Event, "auditId", task.ref.AuditID, "error", err)
	}
}

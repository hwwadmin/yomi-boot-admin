package audit

import (
	"context"
	"time"
	"yomi-admin-backend/domains/auditlog/pkg/shared/enums"
	auditloginput "yomi-admin-backend/domains/auditlog/pkg/shared/input"
	auditlogcontract "yomi-admin-backend/domains/auditlog/service_contract"

	"gitee.com/hq_project/yomi-boot/component/common/snowflake"
	"gitee.com/hq_project/yomi-boot/component/core/logger"
)

// writeTimeout 是单条审计写入上限；同步写但必须尽快返回，不能长时间拖住业务
const writeTimeout = 500 * time.Millisecond

// TwoPhaseRecorder 是 Recorder 的两段式实现：业务前写 STARTED，业务后写 FINISHED。
//
// 写入使用独立的 background context 和独立的数据库连接，不参与业务事务；
// 失败只告警，不重试、不回滚、不影响业务结果。
type TwoPhaseRecorder struct {
	writer auditlogcontract.AuditLogWriter
	ids    snowflake.Generator
}

var _ Recorder = (*TwoPhaseRecorder)(nil)

// NewTwoPhaseRecorder 创建两段式审计写入器
func NewTwoPhaseRecorder(writer auditlogcontract.AuditLogWriter, ids snowflake.Generator) *TwoPhaseRecorder {
	return &TwoPhaseRecorder{writer: writer, ids: ids}
}

// Start 生成 audit_id 并写 STARTED 记录。
//
// 雪花 ID 生成失败时返回空 Ref，Finish 会因此跳过，整对记录一起放弃。
// STARTED 落库失败仍返回 Ref：结束后写出的是没有配对的孤立 FINISHED 记录。
// 这是刻意的取舍——孤条自带时间与结果，仍能还原"发生了什么"，比整对丢失更有用。
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
	r.write(ctx, ref, spec, enums.PhaseStarted, enums.ResultPending, "")
	return ref
}

// Finish 用同一个 audit_id 写 FINISHED 记录；ref 为空（ID 生成失败）时直接跳过
func (r *TwoPhaseRecorder) Finish(ctx context.Context, ref Ref, spec Spec, outcome Outcome) {
	if r == nil || r.writer == nil || ref.AuditID <= 0 {
		return
	}
	r.write(ctx, ref, spec, enums.PhaseFinished, outcome.Result, outcome.BizCode)
}

// write 写一条审计记录；使用独立 context，失败或 panic 只告警
func (r *TwoPhaseRecorder) write(ctx context.Context, ref Ref, spec Spec,
	phase enums.Phase, result enums.Result, bizCode string,
) {
	defer func() {
		if rec := recover(); rec != nil {
			logger.Warn(ctx, "审计写入 panic", "event", spec.Event, "auditId", ref.AuditID, "panic", rec)
		}
	}()

	// 不继承业务 context：避免把业务事务和请求取消带进审计写入
	writeCtx, cancel := context.WithTimeout(context.Background(), writeTimeout)
	defer cancel()

	err := r.writer.Record(writeCtx, &auditloginput.RecordInput{
		AuditID:      ref.AuditID,
		Phase:        phase,
		EventCode:    spec.Event,
		OperatorID:   spec.Operator.ID,
		OperatorName: spec.Operator.Name,
		ClientIP:     spec.ClientIP,
		TargetType:   spec.Target.Type,
		TargetID:     spec.Target.ID,
		Result:       result,
		BizCode:      bizCode,
	})
	if err != nil {
		logger.Warn(writeCtx, "审计写入失败", "event", spec.Event, "auditId", ref.AuditID, "error", err)
	}
}

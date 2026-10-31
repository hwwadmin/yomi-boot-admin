package audit

import (
	"context"
	"errors"
	"testing"
	"time"
	"yomi-admin-backend/domains/auditlog/pkg/shared/enums"
	auditloginput "yomi-admin-backend/domains/auditlog/pkg/shared/input"
)

// fakeWriter 收集写入入参
type fakeWriter struct {
	records chan *auditloginput.RecordInput
}

// Record 把入参送进通道
func (f *fakeWriter) Record(_ context.Context, in *auditloginput.RecordInput) error {
	f.records <- in
	return nil
}

// flakyWriter 第一次写入 panic，之后正常收集
type flakyWriter struct {
	records chan *auditloginput.RecordInput
	failed  bool
}

// Record 首次 panic，验证消费协程不会被打死
func (f *flakyWriter) Record(_ context.Context, in *auditloginput.RecordInput) error {
	if !f.failed {
		f.failed = true
		panic("boom")
	}
	f.records <- in
	return nil
}

// fakeIDs 返回自增 ID
type fakeIDs struct{ next int64 }

// NextID 返回下一个 ID
func (f *fakeIDs) NextID() (int64, error) {
	f.next++
	return f.next, nil
}

// failingIDs 模拟雪花 ID 生成失败
type failingIDs struct{}

// NextID 永远失败
func (failingIDs) NextID() (int64, error) { return 0, errors.New("snowflake down") }

// recv 等待一条写入，超时即失败
func recv(t *testing.T, ch <-chan *auditloginput.RecordInput) *auditloginput.RecordInput {
	t.Helper()
	select {
	case in := <-ch:
		return in
	case <-time.After(2 * time.Second):
		t.Fatal("等待审计写入超时")
		return nil
	}
}

// TestRecorderDeliversBothPhases 验证 STARTED/FINISHED 都以原子记录入队并按序落库
func TestRecorderDeliversBothPhases(t *testing.T) {
	writer := &fakeWriter{records: make(chan *auditloginput.RecordInput, 4)}
	rec := newTwoPhaseRecorder(writer, &fakeIDs{}, 4)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	rec.StartBackground(ctx)

	startSpec := Spec{Event: EventLogin, Operator: Operator{Name: "admin"}, ClientIP: "1.2.3.4"}
	ref := rec.Start(ctx, startSpec)
	rec.Finish(ctx, ref, Spec{Event: EventLogin, Operator: Operator{ID: 7, Name: "admin"}},
		Outcome{Result: enums.ResultSuccess, BizCode: "OK"})

	first := recv(t, writer.records)
	second := recv(t, writer.records)

	if ref.AuditID <= 0 {
		t.Fatalf("auditId 未生成: %d", ref.AuditID)
	}
	if first.Phase != enums.PhaseStarted || first.Result != enums.ResultPending ||
		first.OperatorName != "admin" || first.ClientIP != "1.2.3.4" {
		t.Fatalf("STARTED 记录错误: %+v", first)
	}
	if second.Phase != enums.PhaseFinished || second.Result != enums.ResultSuccess ||
		second.BizCode != "OK" || second.OperatorID != 7 {
		t.Fatalf("FINISHED 记录错误: %+v", second)
	}
	if first.AuditID != second.AuditID || first.AuditID != ref.AuditID {
		t.Fatalf("开始/结束未共用 audit_id: %d / %d", first.AuditID, second.AuditID)
	}
	if first.CreateTime <= 0 || second.CreateTime <= 0 {
		t.Fatalf("事件时间未写入: %d / %d", first.CreateTime, second.CreateTime)
	}
}

// TestRecorderDropsWhenQueueFull 验证队列满时非阻塞丢弃
func TestRecorderDropsWhenQueueFull(t *testing.T) {
	rec := newTwoPhaseRecorder(&fakeWriter{records: make(chan *auditloginput.RecordInput, 1)}, &fakeIDs{}, 1)
	spec := Spec{Event: EventLogin}

	rec.Start(context.Background(), spec) // 占满唯一槽位
	rec.Start(context.Background(), spec) // 队列已满：应直接丢弃，不能阻塞

	if got := rec.dropped.Load(); got != 1 {
		t.Fatalf("丢弃计数 = %d, want 1", got)
	}
}

// TestRecorderDrainsOnStop 验证退出时冲刷队列
func TestRecorderDrainsOnStop(t *testing.T) {
	writer := &fakeWriter{records: make(chan *auditloginput.RecordInput, 4)}
	rec := newTwoPhaseRecorder(writer, &fakeIDs{}, 4)
	ctx, cancel := context.WithCancel(context.Background())
	rec.StartBackground(ctx)

	rec.Start(ctx, Spec{Event: EventLogin})
	cancel()

	select {
	case <-rec.done:
	case <-time.After(3 * time.Second):
		t.Fatal("消费协程未在 ctx 结束后退出")
	}
	if len(writer.records) != 1 {
		t.Fatalf("退出时应冲刷队列，实际落库 %d 条", len(writer.records))
	}
}

// TestRecorderSurvivesWriterPanic 验证单条写入 panic 不会打死消费协程
func TestRecorderSurvivesWriterPanic(t *testing.T) {
	writer := &flakyWriter{records: make(chan *auditloginput.RecordInput, 4)}
	rec := newTwoPhaseRecorder(writer, &fakeIDs{}, 4)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	rec.StartBackground(ctx)

	rec.Start(ctx, Spec{Event: EventLogin}) // 消费时 panic
	rec.Start(ctx, Spec{Event: EventLogin}) // 之后仍应继续消费

	recv(t, writer.records)
}

// TestRecorderSkipsWhenIDFails 验证雪花 ID 失败时放弃整对记录
func TestRecorderSkipsWhenIDFails(t *testing.T) {
	writer := &fakeWriter{records: make(chan *auditloginput.RecordInput, 4)}
	rec := newTwoPhaseRecorder(writer, failingIDs{}, 4)

	ref := rec.Start(context.Background(), Spec{Event: EventLogin})
	if ref.AuditID != 0 {
		t.Fatalf("ID 生成失败应返回空 Ref: %+v", ref)
	}
	rec.Finish(context.Background(), ref, Spec{Event: EventLogin}, Outcome{Result: enums.ResultSuccess})
	if len(rec.tasks) != 0 {
		t.Fatalf("ID 失败后不应入队，实际 %d 条", len(rec.tasks))
	}
}

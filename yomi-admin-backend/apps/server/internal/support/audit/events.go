// Package audit 操作审计两段式契约：业务开始前写 STARTED，业务结束后写 FINISHED。
package audit

import (
	"context"
	"yomi-admin-backend/domains/auditlog/pkg/shared/enums"

	"github.com/hwwadmin/yomi-boot/component/core/errors"
	sharederr "github.com/hwwadmin/yomi-boot/component/core/shared/error"
)

// 审计事件编码，落库后是前端契约；新增事件在此追加编码。
const (
	// EventLogin 登录
	EventLogin = "auth.login"
	// EventChangePassword 修改自身密码
	EventChangePassword = "auth.change_password"
	// EventResetPassword 重置他人密码
	EventResetPassword = "auth.reset_password"
)

// TargetTypeUser 是用户目标的类型标识
const TargetTypeUser = "user"

// Operator 是业务操作者快照
type Operator struct {
	ID   int64
	Name string
}

// Target 是业务目标快照；只记类型与 ID，目标用户名由查询侧按 ID 实时回填，不落库
type Target struct {
	Type string
	ID   int64
}

// Spec 是一次操作要记录的业务事实
type Spec struct {
	Event    string
	Operator Operator
	ClientIP string
	Target   Target
}

// Ref 是两段记录共用的审计事件实例
type Ref struct {
	AuditID int64
}

// Outcome 是业务结果
type Outcome struct {
	Result  enums.Result
	BizCode string
}

// Recorder 是两段式审计写入契约；实现必须 best-effort，失败不得影响业务。
// 实现可以是异步投递（允许队列满时丢弃），调用方不得依赖"返回即已落库"。
type Recorder interface {
	// Start 生成 audit_id 并写 STARTED 记录
	Start(ctx context.Context, spec Spec) Ref
	// Finish 用同一个 audit_id 写 FINISHED 记录
	Finish(ctx context.Context, ref Ref, spec Spec, outcome Outcome)
}

// OutcomeFromError 把业务错误映射为审计结果与业务码，口径与统一响应一致。
func OutcomeFromError(err error) Outcome {
	if err == nil {
		return Outcome{Result: enums.ResultSuccess, BizCode: "OK"}
	}

	var biz *errors.BizError
	if errors.As(err, &biz) && biz != nil && biz.Code != "" {
		return Outcome{Result: enums.ResultFailure, BizCode: biz.Code}
	}
	return Outcome{Result: enums.ResultFailure, BizCode: sharederr.ErrInternal.Error()}
}

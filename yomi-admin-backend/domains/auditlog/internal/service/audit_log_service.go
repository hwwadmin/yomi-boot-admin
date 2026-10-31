// Package service 操作审计日志域服务实现
package service

import (
	"context"
	"strings"
	"time"
	"yomi-admin-backend/domains/auditlog/internal/repo_contract"
	"yomi-admin-backend/domains/auditlog/pkg/shared/entity"
	"yomi-admin-backend/domains/auditlog/pkg/shared/enums"
	"yomi-admin-backend/domains/auditlog/pkg/shared/input"
	"yomi-admin-backend/domains/auditlog/pkg/shared/output"
	"yomi-admin-backend/domains/auditlog/service_contract"

	"gitee.com/hq_project/yomi-boot/component/common/snowflake"
	"gitee.com/hq_project/yomi-boot/component/core/errors"
	"gitee.com/hq_project/yomi-boot/component/core/utils/utf8util"
)

// 字段写入上限（字节），超出按 UTF-8 边界截断
const (
	maxEventCodeBytes    = 64
	maxOperatorNameBytes = 64
	maxClientIPBytes     = 64
	maxTargetTypeBytes   = 32
	maxBizCodeBytes      = 64
)

// AuditLogService 操作审计日志域服务实现
type AuditLogService struct {
	logs repo_contract.AuditLogRepo
	ids  snowflake.Generator
}

var (
	_ service_contract.AuditLogWriter = (*AuditLogService)(nil)
	_ service_contract.AuditLogReader = (*AuditLogService)(nil)
)

// NewAuditLogService 创建审计日志域服务
func NewAuditLogService(logs repo_contract.AuditLogRepo, ids snowflake.Generator) *AuditLogService {
	return &AuditLogService{logs: logs, ids: ids}
}

// Record 写入一条审计记录
func (s *AuditLogService) Record(ctx context.Context, in *input.RecordInput) error {
	if in == nil {
		return errors.Sentinel(service_contract.ErrInvalidInput)
	}

	eventCode := strings.TrimSpace(in.EventCode)
	if eventCode == "" || len(eventCode) > maxEventCodeBytes {
		return errors.Sentinel(service_contract.ErrInvalidInput)
	}
	if in.AuditID <= 0 || !enums.IsValidPhase(in.Phase) {
		return errors.Sentinel(service_contract.ErrInvalidInput)
	}
	if !enums.IsValidResult(in.Result) {
		return errors.Sentinel(service_contract.ErrInvalidInput)
	}
	if in.Phase == enums.PhaseStarted && in.Result != enums.ResultPending {
		return errors.Sentinel(service_contract.ErrInvalidInput)
	}
	if in.Phase == enums.PhaseFinished && in.Result == enums.ResultPending {
		return errors.Sentinel(service_contract.ErrInvalidInput)
	}
	if in.OperatorID < 0 || in.TargetID < 0 {
		return errors.Sentinel(service_contract.ErrInvalidInput)
	}

	id, err := s.ids.NextID()
	if err != nil {
		return errors.Wrap(err, "生成审计日志 ID 失败")
	}

	createTime := in.CreateTime
	if createTime <= 0 {
		createTime = time.Now().UnixMilli()
	}

	return s.logs.Create(ctx, &entity.Log{
		ID:           id,
		AuditID:      in.AuditID,
		Phase:        in.Phase,
		EventCode:    eventCode,
		OperatorID:   in.OperatorID,
		OperatorName: cutString(strings.TrimSpace(in.OperatorName), maxOperatorNameBytes),
		ClientIP:     cutString(strings.TrimSpace(in.ClientIP), maxClientIPBytes),
		TargetType:   cutString(strings.TrimSpace(in.TargetType), maxTargetTypeBytes),
		TargetID:     in.TargetID,
		Result:       in.Result,
		BizCode:      cutString(strings.TrimSpace(in.BizCode), maxBizCodeBytes),
		CreateTime:   createTime,
	})
}

// Page 分页查询审计日志
func (s *AuditLogService) Page(ctx context.Context, in *input.PageInput) (*output.PageLogsOutput, error) {
	if in == nil {
		return nil, errors.Sentinel(service_contract.ErrInvalidInput)
	}
	if in.Result != "" && !enums.IsValidResult(in.Result) {
		return nil, errors.Sentinel(service_contract.ErrInvalidInput)
	}
	if in.Phase != "" && !enums.IsValidPhase(in.Phase) {
		return nil, errors.Sentinel(service_contract.ErrInvalidInput)
	}
	if in.AuditID < 0 {
		return nil, errors.Sentinel(service_contract.ErrInvalidInput)
	}
	// 列表查询必须带时间范围：审计表只增不减，没有窗口的翻页与统计会随表增长退化。
	// 按单个事件 ID 精确查是点查（走 audit_id 索引、最多命中两条），不受此限。
	if in.AuditID <= 0 && (in.StartTime <= 0 || in.EndTime <= 0) {
		return nil, errors.Sentinel(service_contract.ErrInvalidInput)
	}
	if in.StartTime > 0 && in.EndTime > 0 && in.StartTime > in.EndTime {
		return nil, errors.Sentinel(service_contract.ErrInvalidInput)
	}

	eventCode := strings.TrimSpace(in.EventCode)
	if len(eventCode) > maxEventCodeBytes {
		return nil, errors.Sentinel(service_contract.ErrInvalidInput)
	}

	items, total, err := s.logs.Page(ctx, &repo_contract.Filter{
		Keyword:   in.Keyword,
		EventCode: eventCode,
		Phase:     in.Phase,
		Result:    in.Result,
		AuditID:   in.AuditID,
		StartTime: in.StartTime,
		EndTime:   in.EndTime,
		Page:      in.Page,
		PageSize:  in.PageSize,
	})
	if err != nil {
		return nil, err
	}
	return &output.PageLogsOutput{Items: items, Total: total}, nil
}

// cutString 按字节截断，不切断 UTF-8 字符
func cutString(value string, maxBytes int) string {
	if len(value) <= maxBytes {
		return value
	}
	cut, _ := utf8util.Cut([]byte(value), maxBytes)
	return string(cut)
}

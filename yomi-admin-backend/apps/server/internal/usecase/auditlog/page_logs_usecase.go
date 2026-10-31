// Package auditlog 操作审计日志查询编排
package auditlog

import (
	"context"
	"strings"
	"yomi-admin-backend/apps/server/internal/support/audit"
	"yomi-admin-backend/domains/auditlog/pkg/shared/enums"
	domaininput "yomi-admin-backend/domains/auditlog/pkg/shared/input"
	domaincontract "yomi-admin-backend/domains/auditlog/service_contract"
	usercontract "yomi-admin-backend/domains/user/service_contract"

	"gitee.com/hq_project/yomi-boot/component/core/logger"
	"gitee.com/hq_project/yomi-boot/component/core/shared/dto"
)

// PageLogsInput 审计日志分页入参
type PageLogsInput struct {
	dto.PageInput
	Keyword   string       `json:"keyword"`
	EventCode string       `json:"eventCode"`
	Phase     enums.Phase  `json:"phase"`
	Result    enums.Result `json:"result"`
	AuditID   int64        `json:"auditId"`
	StartTime int64        `json:"startTime"`
	EndTime   int64        `json:"endTime"`
}

// LogItem 审计日志列表项
type LogItem struct {
	ID           dto.JSONInt64 `json:"id"`
	AuditID      dto.JSONInt64 `json:"auditId"`
	Phase        enums.Phase   `json:"phase"`
	EventCode    string        `json:"eventCode"`
	OperatorID   dto.JSONInt64 `json:"operatorId"`
	OperatorName string        `json:"operatorName"`
	ClientIP     string        `json:"clientIp"`
	TargetType   string        `json:"targetType"`
	TargetID     dto.JSONInt64 `json:"targetId"`
	TargetName   string        `json:"targetName"`
	Result       enums.Result  `json:"result"`
	BizCode      string        `json:"bizCode"`
	CreateTime   int64         `json:"createTime"`
}

// PageLogsUsecase 审计日志分页编排
type PageLogsUsecase struct {
	logs  domaincontract.AuditLogReader
	users usercontract.UserService
}

// NewPageLogsUsecase 创建审计日志分页用例；users 用于回填目标用户名
func NewPageLogsUsecase(logs domaincontract.AuditLogReader, users usercontract.UserService) *PageLogsUsecase {
	return &PageLogsUsecase{logs: logs, users: users}
}

// Execute 分页查询审计日志
func (uc *PageLogsUsecase) Execute(ctx context.Context, in *PageLogsInput) (dto.PageOutput[LogItem], error) {
	in.ValiPage()

	out, err := uc.logs.Page(ctx, &domaininput.PageInput{
		Keyword:   in.Keyword,
		EventCode: strings.TrimSpace(in.EventCode),
		Phase:     in.Phase,
		Result:    in.Result,
		AuditID:   in.AuditID,
		StartTime: in.StartTime,
		EndTime:   in.EndTime,
		Page:      in.Page,
		PageSize:  in.PageSize,
	})
	if err != nil {
		return dto.PageOutput[LogItem]{}, err
	}

	items := make([]LogItem, 0, len(out.Items))
	for _, log := range out.Items {
		items = append(items, toLogItem(log))
	}
	uc.fillTargetNames(ctx, items)
	return dto.NewPageOutput(items, out.Total), nil
}

// fillTargetNames 按本页出现的目标用户 ID 批量回填用户名；
// 目标用户名不落库，回填失败只告警，列表照常返回（前端会退化成展示目标 ID）
func (uc *PageLogsUsecase) fillTargetNames(ctx context.Context, items []LogItem) {
	if uc.users == nil || len(items) == 0 {
		return
	}

	ids := make([]int64, 0, len(items))
	for _, item := range items {
		if item.TargetType == audit.TargetTypeUser && item.TargetID > 0 {
			ids = append(ids, item.TargetID.Int64())
		}
	}
	if len(ids) == 0 {
		return
	}

	names, err := uc.users.GetNamesByIDs(ctx, ids)
	if err != nil {
		logger.Warn(ctx, "回填审计目标用户名失败", "error", err)
		return
	}
	for i := range items {
		if name, ok := names[items[i].TargetID.Int64()]; ok {
			items[i].TargetName = name
		}
	}
}

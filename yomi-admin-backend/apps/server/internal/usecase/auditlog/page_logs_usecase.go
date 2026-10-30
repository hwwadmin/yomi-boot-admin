// Package auditlog 操作审计日志查询编排
package auditlog

import (
	"context"
	"strings"
	"yomi-admin-backend/domains/auditlog/pkg/shared/enums"
	domaininput "yomi-admin-backend/domains/auditlog/pkg/shared/input"
	domaincontract "yomi-admin-backend/domains/auditlog/service_contract"

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
	logs domaincontract.AuditLogReader
}

// NewPageLogsUsecase 创建审计日志分页用例
func NewPageLogsUsecase(logs domaincontract.AuditLogReader) *PageLogsUsecase {
	return &PageLogsUsecase{logs: logs}
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
	return dto.NewPageOutput(items, out.Total), nil
}

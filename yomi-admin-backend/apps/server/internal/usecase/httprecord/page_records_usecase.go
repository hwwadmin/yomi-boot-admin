// Package httprecord HTTP 访问记录编排
package httprecord

import (
	"context"
	"strings"
	domainenums "yomi-admin-backend/domains/httprecord/pkg/shared/enums"
	domaininput "yomi-admin-backend/domains/httprecord/pkg/shared/input"
	domaincontract "yomi-admin-backend/domains/httprecord/service_contract"

	"gitee.com/hq_project/yomi-boot/component/core/errors"
	"gitee.com/hq_project/yomi-boot/component/core/shared/dto"
)

// PageRecordsInput 访问记录分页入参
type PageRecordsInput struct {
	dto.PageInput
	Keyword     string `json:"keyword"`
	Method      string `json:"method"`
	StatusClass string `json:"statusClass"`
	UserID      int64  `json:"userId"`
	StartTime   int64  `json:"startTime"`
	EndTime     int64  `json:"endTime"`
}

// Validate 校验入参
func (i *PageRecordsInput) Validate() error {
	i.ValiPage()

	i.Method = strings.ToUpper(strings.TrimSpace(i.Method))
	if i.Method != "" && !domainenums.IsValidMethod(i.Method) {
		return errors.Sentinel(domaincontract.ErrInvalidInput)
	}

	i.StatusClass = strings.ToLower(strings.TrimSpace(i.StatusClass))
	if i.StatusClass != "" && !domainenums.IsValidStatusClass(domainenums.StatusClass(i.StatusClass)) {
		return errors.Sentinel(domaincontract.ErrInvalidInput)
	}

	if i.StartTime > 0 && i.EndTime > 0 && i.StartTime > i.EndTime {
		return errors.Sentinel(domaincontract.ErrInvalidInput)
	}
	return nil
}

// RecordItem 访问记录列表项
type RecordItem struct {
	ID             dto.JSONInt64 `json:"id"`
	ServiceType    string        `json:"serviceType"`
	ServiceVersion string        `json:"serviceVersion"`
	TraceID        string        `json:"traceId"`
	RequestAt      int64         `json:"requestAt"`
	DurationMS     int64         `json:"durationMs"`
	Method         string        `json:"method"`
	Path           string        `json:"path"`
	Route          string        `json:"route"`
	ClientIP       string        `json:"clientIp"`
	UserID         dto.JSONInt64 `json:"userId"`
	ClientType     string        `json:"clientType"`
	StatusCode     int           `json:"statusCode"`
	BizCode        string        `json:"bizCode"`
	BizMsg         string        `json:"bizMsg"`
}

// PageRecordsUsecase 访问记录分页编排
type PageRecordsUsecase struct {
	records domaincontract.RecordService
}

// NewPageRecordsUsecase 创建访问记录分页用例
func NewPageRecordsUsecase(records domaincontract.RecordService) *PageRecordsUsecase {
	return &PageRecordsUsecase{records: records}
}

// Execute 分页查询访问记录
func (uc *PageRecordsUsecase) Execute(ctx context.Context, in *PageRecordsInput) (*dto.PageOutput[RecordItem], error) {
	if err := in.Validate(); err != nil {
		return nil, err
	}

	out, err := uc.records.Page(ctx, &domaininput.PageRecordsInput{
		Keyword:     in.Keyword,
		Method:      in.Method,
		StatusClass: in.StatusClass,
		UserID:      in.UserID,
		StartTime:   in.StartTime,
		EndTime:     in.EndTime,
		Page:        in.Page,
		PageSize:    in.PageSize,
	})
	if err != nil {
		return nil, err
	}

	items := make([]RecordItem, 0, len(out.Items))
	for _, record := range out.Items {
		items = append(items, toRecordItem(record))
	}
	return &dto.PageOutput[RecordItem]{Items: items, Total: out.Total}, nil
}

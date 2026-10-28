// Package httprecord HTTP 访问记录编排
package httprecord

import (
	"context"
	domaincontract "yomi-admin-backend/domains/httprecord/service_contract"
)

// GetRecordUsecase 访问记录详情编排
type GetRecordUsecase struct {
	records domaincontract.RecordService
}

// NewGetRecordUsecase 创建访问记录详情用例
func NewGetRecordUsecase(records domaincontract.RecordService) *GetRecordUsecase {
	return &GetRecordUsecase{records: records}
}

// Execute 查询访问记录详情
func (uc *GetRecordUsecase) Execute(ctx context.Context, id int64) (*RecordDetail, error) {
	record, err := uc.records.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	return &RecordDetail{
		RecordItem:            toRecordItem(record),
		Query:                 record.Query,
		UserAgent:             record.UserAgent,
		RequestHeaders:        record.RequestHeaders,
		RequestBody:           record.RequestBody,
		RequestBodyTruncated:  record.RequestBodyTruncated,
		ResponseBody:          record.ResponseBody,
		ResponseBodyTruncated: record.ResponseBodyTruncated,
		CreateTime:            record.CreateTime,
	}, nil
}

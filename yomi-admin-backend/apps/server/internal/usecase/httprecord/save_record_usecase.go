// Package httprecord HTTP 访问记录编排
package httprecord

import (
	"context"

	domaininput "yomi-admin-backend/domains/httprecord/pkg/shared/input"
	domaincontract "yomi-admin-backend/domains/httprecord/service_contract"

	"gitee.com/hq_project/yomi-boot/component/core/errors"
)

// ErrInvalidRecord 表示记录内容本身不合法；消费适配器应丢弃消息而不是重投。
var ErrInvalidRecord = errors.NewNoStack("HTTP_RECORD_INVALID")

// SaveRecordInput 访问记录消费入参：由 httprecord 插件记录映射而来
type SaveRecordInput struct {
	ServiceType    string
	ServiceVersion string
	TraceID        string
	RequestAt      int64 // 请求到达时间，Unix 毫秒时间戳
	DurationMS     int64

	Method    string
	Path      string
	Route     string
	Query     string
	ClientIP  string
	UserAgent string

	RequestHeaders       map[string]string
	RequestBody          string
	RequestBodyTruncated bool

	UserID     int64
	ClientType string

	StatusCode            int
	ResponseBody          string
	ResponseBodyTruncated bool
	BizCode               string
	BizMsg                string
}

// SaveRecordUsecase 访问记录落库编排
type SaveRecordUsecase struct {
	records domaincontract.RecordService
}

// NewSaveRecordUsecase 创建访问记录落库用例
func NewSaveRecordUsecase(records domaincontract.RecordService) *SaveRecordUsecase {
	return &SaveRecordUsecase{records: records}
}

// Execute 保存一条访问记录；记录不合法时返回 ErrInvalidRecord 并保留领域错误链
func (uc *SaveRecordUsecase) Execute(ctx context.Context, in *SaveRecordInput) error {
	err := uc.records.Save(ctx, &domaininput.SaveRecordInput{
		ServiceType:           in.ServiceType,
		ServiceVersion:        in.ServiceVersion,
		TraceID:               in.TraceID,
		RequestAt:             in.RequestAt,
		DurationMS:            in.DurationMS,
		Method:                in.Method,
		Path:                  in.Path,
		Route:                 in.Route,
		Query:                 in.Query,
		ClientIP:              in.ClientIP,
		UserAgent:             in.UserAgent,
		RequestHeaders:        in.RequestHeaders,
		RequestBody:           in.RequestBody,
		RequestBodyTruncated:  in.RequestBodyTruncated,
		UserID:                in.UserID,
		ClientType:            in.ClientType,
		StatusCode:            in.StatusCode,
		ResponseBody:          in.ResponseBody,
		ResponseBodyTruncated: in.ResponseBodyTruncated,
		BizCode:               in.BizCode,
		BizMsg:                in.BizMsg,
	})
	if err != nil && errors.Is(err, domaincontract.ErrInvalidInput) {
		return errors.Join(ErrInvalidRecord, err)
	}
	return err
}

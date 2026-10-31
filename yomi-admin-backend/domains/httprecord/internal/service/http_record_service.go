// Package service HTTP 访问记录域服务实现
package service

import (
	"context"
	"encoding/json"
	"strings"
	"time"
	"yomi-admin-backend/domains/httprecord/internal/repo_contract"
	"yomi-admin-backend/domains/httprecord/pkg/shared/entity"
	"yomi-admin-backend/domains/httprecord/pkg/shared/enums"
	"yomi-admin-backend/domains/httprecord/pkg/shared/input"
	"yomi-admin-backend/domains/httprecord/pkg/shared/output"
	"yomi-admin-backend/domains/httprecord/service_contract"

	"gitee.com/hq_project/yomi-boot/component/common/snowflake"
	"gitee.com/hq_project/yomi-boot/component/core/errors"
	"gitee.com/hq_project/yomi-boot/component/core/utils/utf8util"
)

// 各字段写入上限（字节），超出按 UTF-8 边界截断，避免超长请求撑爆列宽
const (
	maxServiceTypeBytes    = 64
	maxServiceVersionBytes = 32
	maxTraceIDBytes        = 64
	maxMethodBytes         = 16
	maxPathBytes           = 512
	maxRouteBytes          = 512
	maxQueryBytes          = 8192
	maxClientIPBytes       = 64
	maxUserAgentBytes      = 1024
	maxRequestHeadersBytes = 8192
	maxClientTypeBytes     = 32
	maxBizCodeBytes        = 64
	maxBizMsgBytes         = 1024
)

// RecordService HTTP 访问记录域服务实现
type RecordService struct {
	records     repo_contract.RecordRepo
	idGenerator snowflake.Generator
}

var _ service_contract.RecordService = (*RecordService)(nil)

// NewRecordService 创建访问记录领域服务
func NewRecordService(records repo_contract.RecordRepo, idGenerator snowflake.Generator) *RecordService {
	return &RecordService{records: records, idGenerator: idGenerator}
}

// Save 持久化一条访问记录
func (s *RecordService) Save(ctx context.Context, in *input.SaveRecordInput) error {
	if in == nil {
		return errors.Sentinel(service_contract.ErrInvalidInput)
	}

	method := cutString(strings.ToUpper(strings.TrimSpace(in.Method)), maxMethodBytes)
	path := cutString(strings.TrimSpace(in.Path), maxPathBytes)
	if method == "" || path == "" {
		return errors.Sentinel(service_contract.ErrInvalidInput)
	}

	requestAt := in.RequestAt
	if requestAt <= 0 {
		requestAt = time.Now().UnixMilli()
	}

	id, err := s.idGenerator.NextID()
	if err != nil {
		return errors.Wrap(err, "生成访问记录 ID 失败")
	}

	headers, err := marshalHeaders(in.RequestHeaders)
	if err != nil {
		return err
	}

	return s.records.Create(ctx, &entity.Record{
		ID:                    id,
		ServiceType:           cutString(in.ServiceType, maxServiceTypeBytes),
		ServiceVersion:        cutString(in.ServiceVersion, maxServiceVersionBytes),
		TraceID:               cutString(in.TraceID, maxTraceIDBytes),
		RequestAt:             requestAt,
		DurationMS:            in.DurationMS,
		Method:                method,
		Path:                  path,
		Route:                 cutString(in.Route, maxRouteBytes),
		Query:                 cutString(in.Query, maxQueryBytes),
		ClientIP:              cutString(in.ClientIP, maxClientIPBytes),
		UserAgent:             cutString(in.UserAgent, maxUserAgentBytes),
		RequestHeaders:        headers,
		RequestBody:           in.RequestBody,
		RequestBodyTruncated:  in.RequestBodyTruncated,
		UserID:                in.UserID,
		ClientType:            cutString(in.ClientType, maxClientTypeBytes),
		StatusCode:            in.StatusCode,
		ResponseBody:          in.ResponseBody,
		ResponseBodyTruncated: in.ResponseBodyTruncated,
		BizCode:               cutString(in.BizCode, maxBizCodeBytes),
		BizMsg:                cutString(in.BizMsg, maxBizMsgBytes),
	})
}

// Page 分页查询访问记录
func (s *RecordService) Page(ctx context.Context, in *input.PageRecordsInput) (*output.PageRecordsOutput, error) {
	if in == nil {
		return nil, errors.Sentinel(service_contract.ErrInvalidInput)
	}

	method := strings.ToUpper(strings.TrimSpace(in.Method))
	if method != "" && !enums.IsValidMethod(method) {
		return nil, errors.Sentinel(service_contract.ErrInvalidInput)
	}

	statusClass := enums.StatusClass(strings.ToLower(strings.TrimSpace(in.StatusClass)))
	if statusClass != "" && !enums.IsValidStatusClass(statusClass) {
		return nil, errors.Sentinel(service_contract.ErrInvalidInput)
	}
	// 列表查询必须带时间范围：访问记录只增不减，没有窗口的翻页与统计会随表增长退化。
	if in.StartTime <= 0 || in.EndTime <= 0 {
		return nil, errors.Sentinel(service_contract.ErrInvalidInput)
	}
	if in.StartTime > in.EndTime {
		return nil, errors.Sentinel(service_contract.ErrInvalidInput)
	}

	items, total, err := s.records.Page(ctx, &repo_contract.RecordFilter{
		Keyword:     in.Keyword,
		Method:      method,
		StatusClass: string(statusClass),
		UserID:      in.UserID,
		StartTime:   in.StartTime,
		EndTime:     in.EndTime,
		Page:        in.Page,
		PageSize:    in.PageSize,
	})
	if err != nil {
		return nil, err
	}
	return &output.PageRecordsOutput{Items: items, Total: total}, nil
}

// Get 按主键查询访问记录详情
func (s *RecordService) Get(ctx context.Context, id int64) (*entity.Record, error) {
	if id <= 0 {
		return nil, errors.Sentinel(service_contract.ErrRecordNotFound)
	}

	record, found, err := s.records.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, errors.Sentinel(service_contract.ErrRecordNotFound)
	}
	return record, nil
}

// marshalHeaders 将请求头序列化为 JSON；空集合返回空串
func marshalHeaders(headers map[string]string) (string, error) {
	if len(headers) == 0 {
		return "", nil
	}
	payload, err := json.Marshal(headers)
	if err != nil {
		return "", errors.Wrap(err, "序列化请求头失败")
	}
	return cutString(string(payload), maxRequestHeadersBytes), nil
}

// cutString 按字节上限截断字符串并保证 UTF-8 合法
func cutString(value string, maxBytes int) string {
	if value == "" || len(value) <= maxBytes {
		return value
	}
	cut, _ := utf8util.Cut([]byte(value), maxBytes)
	return string(cut)
}

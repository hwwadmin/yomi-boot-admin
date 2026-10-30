package handler

import (
	"context"
	httprecoreusecase "yomi-admin-backend/apps/server/internal/usecase/httprecord"

	"gitee.com/hq_project/yomi-boot/component/common/httprecord"
	"gitee.com/hq_project/yomi-boot/component/core/errors"
	"gitee.com/hq_project/yomi-boot/component/core/logger"
)

// recordSaver 保存一条访问记录；由 usecase/httprecord.SaveRecordUsecase 实现。
type recordSaver interface {
	Execute(ctx context.Context, in *httprecoreusecase.SaveRecordInput) error
}

// NewHttpRecordConsumer 构造访问记录消费适配器：把插件记录映射为用例入参并落库。
//
// 记录不合法说明消息本身有问题，直接丢弃避免毒消息反复重投；落库失败按可重试处理。
func NewHttpRecordConsumer(save *httprecoreusecase.SaveRecordUsecase) httprecord.RecordHandler {
	return newHttpRecordConsumer(save)
}

// newHttpRecordConsumer 接受接口，便于单测注入保存结果。
func newHttpRecordConsumer(save recordSaver) httprecord.RecordHandler {
	return func(ctx context.Context, record *httprecord.HttpRecord) httprecord.RecordHandleResult {
		if record == nil {
			return httprecord.HandleDrop()
		}

		err := save.Execute(ctx, &httprecoreusecase.SaveRecordInput{
			ServiceType:           record.ServiceType,
			ServiceVersion:        record.ServiceVersion,
			TraceID:               record.TraceID,
			RequestAt:             record.RequestAt.UnixMilli(),
			DurationMS:            record.Duration.Milliseconds(),
			Method:                record.Method,
			Path:                  record.Path,
			Route:                 record.Route,
			Query:                 record.Query,
			ClientIP:              record.ClientIP,
			UserAgent:             record.UserAgent,
			RequestHeaders:        record.RequestHeaders,
			RequestBody:           record.RequestBody,
			RequestBodyTruncated:  record.RequestBodyTruncated,
			UserID:                record.UserID,
			ClientType:            record.ClientType,
			StatusCode:            record.StatusCode,
			ResponseBody:          record.ResponseBody,
			ResponseBodyTruncated: record.ResponseBodyTruncated,
			BizCode:               record.BizCode,
			BizMsg:                record.BizMsg,
		})
		if err != nil {
			if errors.Is(err, httprecoreusecase.ErrInvalidRecord) {
				logger.Warn(ctx, "HTTP 访问记录不合法，已丢弃",
					"traceId", record.TraceID,
					"method", record.Method,
					"path", record.Path,
				)
				return httprecord.HandleDrop()
			}
			logger.Warn(ctx, "HTTP 访问记录落库失败",
				"error", err,
				"traceId", record.TraceID,
				"method", record.Method,
				"path", record.Path,
			)
			return httprecord.HandleRetry()
		}
		return httprecord.HandleOK()
	}
}

// Package bootstrap 编排层装配入口
package bootstrap

import (
	"context"
	httprecordinput "yomi-admin-backend/domains/httprecord/pkg/shared/input"
	httprecordcontract "yomi-admin-backend/domains/httprecord/service_contract"

	"gitee.com/hq_project/yomi-boot/component/common/httprecord"
	"gitee.com/hq_project/yomi-boot/component/core/errors"
	"gitee.com/hq_project/yomi-boot/component/core/logger"
	"gitee.com/hq_project/yomi-boot/plugin"
	httprecordplugin "gitee.com/hq_project/yomi-boot/plugin/builtin/httprecord"
)

// registerHttpRecordHandler 向 httprecord 插件注册访问记录消费处理函数
// 须在插件 OnReady 启动消费者之前调用，wire 阶段即满足该时序
func registerHttpRecordHandler(slot *plugin.PluginSlot, records httprecordcontract.RecordService) {
	recordPlugin := httprecordplugin.From(slot)
	if recordPlugin == nil {
		return
	}
	recordPlugin.MustRegisterHandler(buildRecordHandler(records))
}

// buildRecordHandler 构造访问记录消费处理函数
func buildRecordHandler(records httprecordcontract.RecordService) httprecord.RecordHandler {
	return func(ctx context.Context, record *httprecord.HttpRecord) httprecord.RecordHandleResult {
		if record == nil {
			return httprecord.HandleDrop()
		}

		err := records.Save(ctx, &httprecordinput.SaveRecordInput{
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
			if errors.Is(err, httprecordcontract.ErrInvalidInput) {
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

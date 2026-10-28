// Package bootstrap 编排层装配入口
package bootstrap

import (
	"context"
	"fmt"
	"os"
	httprecordinput "yomi-admin-backend/domains/httprecord/pkg/shared/input"
	httprecordcontract "yomi-admin-backend/domains/httprecord/service_contract"

	"gitee.com/hq_project/yomi-boot/component/common/httprecord"
	"gitee.com/hq_project/yomi-boot/component/core/config"
	"gitee.com/hq_project/yomi-boot/component/core/errors"
	"gitee.com/hq_project/yomi-boot/component/core/logger"
	"gitee.com/hq_project/yomi-boot/plugin"
	httprecordplugin "gitee.com/hq_project/yomi-boot/plugin/builtin/httprecord"
	kafkaplugin "gitee.com/hq_project/yomi-boot/plugin/builtin/kafka"
	redisstreamplugin "gitee.com/hq_project/yomi-boot/plugin/builtin/redisstream"
)

// httpRecordConsumerName 是业务侧访问记录消费者插件名
const httpRecordConsumerName = "http-record-consumer"

// HttpRecordConsumerPlugin 启动访问记录消费者并把记录写入数据库
type HttpRecordConsumerPlugin struct {
	plugin.Base
}

var _ plugin.Plugin = (*HttpRecordConsumerPlugin)(nil)

// NewHttpRecordConsumerPlugin 创建访问记录消费者插件
func NewHttpRecordConsumerPlugin() *HttpRecordConsumerPlugin {
	return &HttpRecordConsumerPlugin{Base: plugin.NewBase(httpRecordConsumerName)}
}

// Init 按 http_record.provider 启动对应消费者；消费逻辑落在本可部署单元
func (p *HttpRecordConsumerPlugin) Init(_ context.Context, slot *plugin.PluginSlot) error {
	bootCfg := config.MustSection[httprecordplugin.BootConfig](slot.Config, httprecordplugin.SectionKey)
	if !bootCfg.EnabledOrDefault() {
		return plugin.ErrDisabled
	}

	records := load(slot).records.Records
	if records == nil {
		return errors.New("http_record: 访问记录服务未装配")
	}
	handler := buildRecordHandler(records)
	cfg := bootCfg.ToConfig()

	switch cfg.ProviderOrDefault() {
	case httprecord.ProviderMemory:
		// 内存队列的消费端点需要具体服务实例，由 httprecord 插件按 provider 构造
		httpRecordPlugin := httprecordplugin.From(slot)
		if httpRecordPlugin == nil {
			return errors.New("http_record: 未注册 httprecord 插件")
		}
		memory, ok := httpRecordPlugin.Service().(*httprecord.MemoryHttpRecordService)
		if !ok || memory == nil {
			return errors.New("http_record: provider=memory 时未创建内存队列服务")
		}
		return httprecord.StartMemoryRecordConsumer(
			slot.RuntimeContext(),
			memory,
			httprecord.MemoryRecordConsumerConfig{Workers: cfg.Memory.WorkersOrDefault()},
			handler,
		)
	case httprecord.ProviderKafka:
		if !cfg.Consumer.EnabledOrDefault() {
			return plugin.ErrDisabled
		}
		consumerPlugin := kafkaplugin.FromConsumer(slot)
		if consumerPlugin == nil || consumerPlugin.Service() == nil {
			return errors.New("http_record: provider=kafka 需要 kafka.consumer_enabled=true")
		}
		return httprecord.StartKafkaRecordConsumer(
			slot.RuntimeContext(),
			consumerPlugin.Service(),
			httprecord.KafkaConsumerID,
			httprecord.KafkaRecordConsumerConfig{
				Topics:  []string{cfg.KafkaTopicOrDefault()},
				GroupID: cfg.Consumer.GroupIDOrDefault(),
			},
			handler,
		)
	case httprecord.ProviderRedisStream:
		if !cfg.Consumer.EnabledOrDefault() {
			return plugin.ErrDisabled
		}
		consumerPlugin := redisstreamplugin.From(slot)
		if consumerPlugin == nil || consumerPlugin.Consumer() == nil {
			return errors.New("http_record: provider=redis_stream 需要 redis_stream.enabled=true")
		}
		return httprecord.StartRedisStreamRecordConsumer(
			slot.RuntimeContext(),
			consumerPlugin.Consumer(),
			httprecord.RedisStreamConsumerID,
			httprecord.RedisStreamRecordConsumerConfig{
				Stream:   cfg.StreamOrDefault(),
				Group:    cfg.Consumer.GroupIDOrDefault(),
				Consumer: fmt.Sprintf("%s-%d", load(slot).appCtx.Name(), os.Getpid()),
			},
			handler,
		)
	default:
		return errors.Newf("http_record: 不支持的 provider: %s", cfg.ProviderOrDefault())
	}
}

// Close 消费者随插件插槽运行期上下文停止，无需额外释放
func (p *HttpRecordConsumerPlugin) Close(context.Context) error {
	return nil
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

// Package handler 入口适配层：HTTP 路由与插件消费回调
package handler

import (
	"context"
	auditlogusecase "yomi-admin-backend/apps/server/internal/usecase/auditlog"
	auditlogenums "yomi-admin-backend/domains/auditlog/pkg/shared/enums"

	"gitee.com/hq_project/yomi-boot/component/core/errors"
	corehandler "gitee.com/hq_project/yomi-boot/component/core/handler"
	"gitee.com/hq_project/yomi-boot/component/core/shared/dto"
	sharederr "gitee.com/hq_project/yomi-boot/component/core/shared/error"

	"github.com/gin-gonic/gin"
)

// AuditLogHandler 操作审计日志路由
type AuditLogHandler struct {
	pageUC *auditlogusecase.PageLogsUsecase
}

// AuditLogHandlerDeps 审计日志 Handler 依赖
type AuditLogHandlerDeps struct {
	Page *auditlogusecase.PageLogsUsecase
}

// NewAuditLogHandler 创建审计日志 Handler
func NewAuditLogHandler(deps AuditLogHandlerDeps) *AuditLogHandler {
	return &AuditLogHandler{pageUC: deps.Page}
}

// Routes 注册审计日志路由
func (h *AuditLogHandler) Routes(router *gin.RouterGroup) {
	router.GET("/api/system/audit-log", h.page)
}

// page 分页查询审计日志
func (h *AuditLogHandler) page(c *gin.Context) {
	result, err := auditlogenums.ParseResult(c.Query("result"))
	if err != nil {
		corehandler.WriteBusinessError(c, errors.Sentinel(sharederr.ErrInvalidInput))
		return
	}
	phase, err := auditlogenums.ParsePhase(c.Query("phase"))
	if err != nil {
		corehandler.WriteBusinessError(c, errors.Sentinel(sharederr.ErrInvalidInput))
		return
	}
	in := &auditlogusecase.PageLogsInput{
		PageInput: dto.PageInput{Page: queryInt(c, "page", 1), PageSize: queryInt(c, "pageSize", 10)},
		Keyword:   c.Query("keyword"),
		EventCode: c.Query("eventCode"),
		Phase:     phase,
		Result:    result,
		AuditID:   queryInt64(c, "auditId", 0),
		StartTime: queryInt64(c, "startTime", 0),
		EndTime:   queryInt64(c, "endTime", 0),
	}
	handle(c, func(ctx context.Context) (dto.PageOutput[auditlogusecase.LogItem], error) {
		return h.pageUC.Execute(ctx, in)
	})
}

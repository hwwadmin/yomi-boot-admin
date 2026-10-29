// Package handler HTTP 适配层
package handler

import (
	"context"
	httprecoreusecase "yomi-admin-backend/apps/server/internal/usecase/httprecord"

	corehandler "gitee.com/hq_project/yomi-boot/component/core/handler"
	"gitee.com/hq_project/yomi-boot/component/core/shared/dto"

	"github.com/gin-gonic/gin"
)

// HttpRecordHandler HTTP 访问记录路由
type HttpRecordHandler struct {
	pageUC *httprecoreusecase.PageRecordsUsecase
	getUC  *httprecoreusecase.GetRecordUsecase
}

// NewHttpRecordHandler 创建 HTTP 访问记录 Handler
func NewHttpRecordHandler(
	pageUC *httprecoreusecase.PageRecordsUsecase,
	getUC *httprecoreusecase.GetRecordUsecase,
) *HttpRecordHandler {
	return &HttpRecordHandler{pageUC: pageUC, getUC: getUC}
}

// Routes 注册 HTTP 访问记录路由
func (h *HttpRecordHandler) Routes(router *gin.RouterGroup) {
	router.GET("/api/system/http-record", h.page)
	router.GET("/api/system/http-record/detail", h.get)
}

func (h *HttpRecordHandler) page(c *gin.Context) {
	in := &httprecoreusecase.PageRecordsInput{
		PageInput:   dto.PageInput{Page: queryInt(c, "page", 1), PageSize: queryInt(c, "pageSize", 10)},
		Keyword:     c.Query("keyword"),
		Method:      c.Query("method"),
		StatusClass: c.Query("statusClass"),
		UserID:      queryInt64(c, "userId", 0),
		StartTime:   queryInt64(c, "startTime", 0),
		EndTime:     queryInt64(c, "endTime", 0),
	}
	corehandler.HandlePage(c, func(ctx context.Context) (*dto.PageOutput[httprecoreusecase.RecordItem], error) {
		return h.pageUC.Execute(ctx, in)
	})
}

func (h *HttpRecordHandler) get(c *gin.Context) {
	id, ok := queryID(c, "id")
	if !ok {
		return
	}
	corehandler.Handle(c, func(ctx context.Context) (*httprecoreusecase.RecordDetail, error) {
		return h.getUC.Execute(ctx, id)
	})
}

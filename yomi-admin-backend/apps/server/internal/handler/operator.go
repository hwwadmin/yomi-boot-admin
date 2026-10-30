// Package handler 入口适配层：HTTP 路由与插件消费回调
package handler

import (
	"strconv"

	"gitee.com/hq_project/yomi-boot/component/common/token"
	"gitee.com/hq_project/yomi-boot/component/core/errors"
	corehandler "gitee.com/hq_project/yomi-boot/component/core/handler"
	sharederr "gitee.com/hq_project/yomi-boot/component/core/shared/error"

	"github.com/gin-gonic/gin"
)

// operatorID 从令牌会话读取当前操作者 ID
func operatorID(c *gin.Context) (int64, bool) {
	id, err := token.OperatorID(c.Request.Context())
	if err != nil {
		corehandler.WriteBusinessError(c, err)
		return 0, false
	}
	return id, true
}

// queryID 解析查询参数中的雪花 ID
func queryID(c *gin.Context, name string) (int64, bool) {
	id, err := strconv.ParseInt(c.Query(name), 10, 64)
	if err != nil || id <= 0 {
		corehandler.WriteBusinessError(c, errors.Sentinel(sharederr.ErrInvalidInput))
		return 0, false
	}
	return id, true
}

// queryInt 读取整数查询参数，非法或缺省时使用默认值
func queryInt(c *gin.Context, key string, defaultValue int) int {
	value, err := strconv.Atoi(c.DefaultQuery(key, strconv.Itoa(defaultValue)))
	if err != nil {
		return defaultValue
	}
	return value
}

// queryInt64 读取 64 位整数查询参数，非法或缺省时使用默认值
func queryInt64(c *gin.Context, key string, defaultValue int64) int64 {
	value, err := strconv.ParseInt(c.DefaultQuery(key, strconv.FormatInt(defaultValue, 10)), 10, 64)
	if err != nil {
		return defaultValue
	}
	return value
}

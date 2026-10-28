// Package repo HTTP 访问记录仓储实现
package repo

import (
	"time"

	"gitee.com/hq_project/yomi-boot/component/common/db/orm"
)

// tableName 是 HTTP 访问记录表名
const tableName = "t_http_record"

// httpRecordModel 是 t_http_record 的持久化模型
type httpRecordModel struct {
	ServiceType    string    `gorm:"column:service_type;size:64"`
	ServiceVersion string    `gorm:"column:service_version;size:32"`
	TraceID        string    `gorm:"column:trace_id;size:64;uniqueIndex:uk_trace_request,priority:1"`
	RequestAt      time.Time `gorm:"column:request_at;type:datetime(3);uniqueIndex:uk_trace_request,priority:2;index:idx_request_at"`
	DurationMS     int64     `gorm:"column:duration_ms"`

	Method    string `gorm:"column:method;size:16"`
	Path      string `gorm:"column:path;size:512"`
	Route     string `gorm:"column:route;size:512"`
	Query     string `gorm:"column:query;type:text"`
	ClientIP  string `gorm:"column:client_ip;size:64"`
	UserAgent string `gorm:"column:user_agent;size:1024"`

	RequestHeaders       string `gorm:"column:request_headers;type:text"`
	RequestBody          string `gorm:"column:request_body;type:mediumtext"`
	RequestBodyTruncated bool   `gorm:"column:request_body_truncated"`

	UserID     int64  `gorm:"column:user_id;index:idx_user_id"`
	ClientType string `gorm:"column:client_type;size:32"`

	StatusCode            int    `gorm:"column:status_code"`
	ResponseBody          string `gorm:"column:response_body;type:mediumtext"`
	ResponseBodyTruncated bool   `gorm:"column:response_body_truncated"`
	BizCode               string `gorm:"column:biz_code;size:64"`
	BizMsg                string `gorm:"column:biz_msg;size:1024"`

	orm.BaseModel
}

// TableName 返回表名
func (httpRecordModel) TableName() string {
	return tableName
}

// Package entity HTTP 访问记录实体
package entity

// Record 表示一条 HTTP 访问记录
type Record struct {
	ID             int64
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

	RequestHeaders       string // 请求头 JSON（已脱敏）
	RequestBody          string // 请求体（已脱敏）
	RequestBodyTruncated bool

	UserID     int64
	ClientType string

	StatusCode            int
	ResponseBody          string // 响应体（已脱敏）
	ResponseBodyTruncated bool
	BizCode               string
	BizMsg                string

	CreateTime int64 // 记录落库时间，Unix 毫秒时间戳
}

// Package input HTTP 访问记录域入参
package input

// SaveRecordInput 写入访问记录入参
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

// PageRecordsInput 访问记录分页查询入参
type PageRecordsInput struct {
	Keyword     string
	Method      string
	StatusClass string
	UserID      int64
	StartTime   int64 // 起始时间，Unix 毫秒时间戳
	EndTime     int64 // 结束时间，Unix 毫秒时间戳
	Page        int
	PageSize    int
}

// Package enums HTTP 访问记录枚举
package enums

// StatusClass 表示 HTTP 状态码分段
type StatusClass string

const (
	// StatusClass2xx 表示 2xx 成功响应
	StatusClass2xx StatusClass = "2xx"
	// StatusClass3xx 表示 3xx 重定向响应
	StatusClass3xx StatusClass = "3xx"
	// StatusClass4xx 表示 4xx 客户端错误
	StatusClass4xx StatusClass = "4xx"
	// StatusClass5xx 表示 5xx 服务端错误
	StatusClass5xx StatusClass = "5xx"
)

// IsValidStatusClass 判断状态码分段是否合法
func IsValidStatusClass(value StatusClass) bool {
	_, _, ok := StatusCodeRange(value)
	return ok
}

// StatusCodeRange 返回状态码分段对应的下界与上界（左闭右开）
func StatusCodeRange(value StatusClass) (min, max int, ok bool) {
	switch value {
	case StatusClass2xx:
		return 200, 300, true
	case StatusClass3xx:
		return 300, 400, true
	case StatusClass4xx:
		return 400, 500, true
	case StatusClass5xx:
		return 500, 600, true
	default:
		return 0, 0, false
	}
}

// httpMethods 是可查询的 HTTP 方法集合
var httpMethods = map[string]struct{}{
	"GET":     {},
	"POST":    {},
	"PUT":     {},
	"DELETE":  {},
	"PATCH":   {},
	"HEAD":    {},
	"OPTIONS": {},
}

// IsValidMethod 判断 HTTP 方法是否在可查询集合内
func IsValidMethod(method string) bool {
	_, ok := httpMethods[method]
	return ok
}

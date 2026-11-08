package orm

import "strings"

// likeEscape 转义 LIKE 模式中的 \、% 和 _。数据库默认用 \ 作为转义字符。
func likeEscape(value string) string {
	if value == "" {
		return ""
	}

	var builder strings.Builder
	builder.Grow(len(value))

	for _, r := range value {
		switch r {
		case '\\':
			builder.WriteString("\\\\")
		case '%':
			builder.WriteString("\\%")
		case '_':
			builder.WriteString("\\_")
		default:
			builder.WriteRune(r)
		}
	}

	return builder.String()
}

// LikeFull 拼接全模糊 LIKE 模式：%value%。
// 返回值已转义通配符，可直接作为参数化查询的值。
func LikeFull(value string) string {
	return "%" + likeEscape(value) + "%"
}

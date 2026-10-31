package support

import (
	"gitee.com/hq_project/yomi-boot/component/core/shared/dto"
)

// NonNilIDs 将 nil 切片规范化为空切片。
// 领域层「无数据」返回 nil 切片是 Go 惯例，但 JSON 序列化会得到 null；
// 前端按数组直接消费（如 includes / v-model），null 会引发渲染异常。
// 各用例的 HTTP 详情输出统一用它保证 ID 数组字段永远是 []。
func NonNilIDs(values dto.JSONInt64Slice) dto.JSONInt64Slice {
	if values == nil {
		return dto.JSONInt64Slice{}
	}
	return values
}

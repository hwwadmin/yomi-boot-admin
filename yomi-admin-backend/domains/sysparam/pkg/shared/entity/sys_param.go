// Package entity 系统参数实体
package entity

import "yomi-admin-backend/domains/sysparam/pkg/shared/enums"

// Param 表示一条系统参数：DB 行是业务默认值之上的"覆盖值"，
// 参数被删除、禁用或不存在时，读取方使用自己代码内的默认值。
type Param struct {
	ID         int64
	Key        string // 参数键，全局唯一，创建后不可变
	Value      string // 统一字符串形态存储，按 ValueType 解释
	ValueType  enums.ParamValueType
	Name       string // 展示名
	Remark     string
	Status     enums.Status // DISABLED 表示覆盖值失效，读取方回退默认值
	CreateTime int64        // Unix 毫秒时间戳
	UpdateTime int64        // Unix 毫秒时间戳
}

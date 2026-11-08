package orm

import (
	"github.com/hwwadmin/yomi-boot/component/core/shared/dto"

	"gorm.io/gorm"
)

// ApplyPage 应用分页与结构化排序。
// page 从 1 开始；page <= 0 时默认为 1。pageSize <= 0 时默认为 10，超过上限时截断。
func ApplyPage(query *gorm.DB, page, pageSize int, orders ...OrderClause) (*gorm.DB, error) {
	if page <= 0 {
		page = dto.DefaultPage
	}
	if pageSize <= 0 {
		pageSize = dto.DefaultPageSize
	}
	if pageSize > dto.MaxPageSize {
		pageSize = dto.MaxPageSize
	}
	offset := (page - 1) * pageSize

	paged, err := applyOrders(query, orders...)
	if err != nil {
		return nil, err
	}
	return paged.Limit(pageSize).Offset(offset), nil
}

// Package output 系统参数域服务出参
package output

import "yomi-admin-backend/domains/sysparam/pkg/shared/entity"

// PageParamsOutput 参数分页出参
type PageParamsOutput struct {
	Items []*entity.Param
	Total int64
}

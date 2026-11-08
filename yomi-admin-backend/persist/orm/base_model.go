package orm

import (
	"time"

	"gorm.io/gorm"
)

// BaseModel 是管理后台表共用的主键、审计字段与软删除列。
type BaseModel struct {
	ID         int64          `gorm:"column:id;primaryKey" json:"id,string"`
	CreateTime time.Time      `gorm:"column:create_time;autoCreateTime" json:"createTime"`
	CreateBy   *int64         `gorm:"column:create_by" json:"createBy,omitempty"`
	UpdateTime time.Time      `gorm:"column:update_time;autoUpdateTime" json:"updateTime"`
	UpdateBy   *int64         `gorm:"column:update_by" json:"updateBy,omitempty"`
	DeleteTime gorm.DeletedAt `gorm:"column:delete_time;index" json:"deleteTime,omitempty"`
	DeleteBy   *int64         `gorm:"column:delete_by" json:"deleteBy,omitempty"`
}

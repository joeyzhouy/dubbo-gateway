package entry

import "time"

type Base struct {
	ID         int64      `gorm:"primary_key" json:"id,omitempty"`
	CreateTime *time.Time `gorm:"column:create_time;default:CURRENT_TIMESTAMP" json:"createTime,omitempty"`
	ModifyTime *time.Time `gorm:"column:modify_time;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP" json:"modifyTime,omitempty"`
	IsDelete   int        `gorm:"column:is_delete" json:"isDelete,omitempty"`
}

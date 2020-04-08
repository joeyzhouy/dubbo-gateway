package entry

import "time"

type Base struct {
	ID         int64     `gorm:"primary_key" json:"id"`
	CreateTime time.Time `gorm:"column:create_time" json:"createTime"`
	ModifyTime time.Time `gorm:"column:modify_time" json:"modifyTime"`
	IsDelete   int       `gorm:"column:is_delete" json:"isDelete"`
}

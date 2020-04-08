package entry

type Registry struct {
	Name     string `gorm:"column:name" json:"name"`
	Protocol string `gorm:"column:protocol" json:"protocol"`
	Address  string `gorm:"column:address" json:"address"`
	UserName string `gorm:"column:user_name" json:"userName"`
	Password string `gorm:"column:password" json:"password"`
	UserId   int64  `gorm:"column:user_id" json:"userId"`
	Base
}

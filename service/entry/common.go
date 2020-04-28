package entry

type User struct {
	Base
	Name     string `gorm:"column:name" json:"userName,omitempty"`
	Email    string `gorm:"column:email" json:"email,omitempty"`
	Password string `gorm:"column:password" json:"password,omitempty"`
}

func (User) TableName() string {
	return "d_user"
}


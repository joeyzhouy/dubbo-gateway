package entry

type User struct {
	Base
	Name     string `gorm:"column:name" json:"name"`
	Email    string `gorm:"column:email" json:"email"`
	Password string `gorm:"column:password" json:"password"`
}

func (User) TableName() string {
	return "d_user"
}


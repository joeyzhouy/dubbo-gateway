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

type ApiConfig struct {
	Base
	Uri     string `gorm:"column:uri" json:"uri"`
	UriHash string `gorm:"column:uri_hash" json:"uriHash"`
	UserId  int64  `gorm:"column:user_id" json:"userId"`
}

func (ApiConfig) TableName() string {
	return "d_api_config"
}

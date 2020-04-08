package entry

type ApiConfig struct {
	Base
	UserId   int64  `gorm:"column:user_id" json:"userId"`
	Desc     string `gorm:"column:desc" json:"desc"`
	Uri      string `gorm:"column:uri" json:"uri"`
	UriHash  string `gorm:"column:uri_hash" json:"uriHash"`
	MethodId int64  `gorm:"column:method_id" json:"methodId"`
}

func (ApiConfig) TableName() string {
	return "d_api_config"
}

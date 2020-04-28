package entry

type ApiConfig struct {
	Base
	UserId   int64  `gorm:"column:user_id" json:"userId"`
	Desc     string `gorm:"column:desc" json:"desc"`
	Uri      string `gorm:"column:uri" json:"uri"`
	UriHash  string `gorm:"column:uri_hash" json:"uriHash"`
	FilterId int64  `gorm:"column:filter_id" json:"filterId"`
}

func (ApiConfig) TableName() string {
	return "d_api_config"
}

type ApiFilter struct {
	Base
	ReferenceId int64  `gorm:"column:reference_id" json:"referenceId"`
	//MethodName  string `gorm:"column:method_name" json:"methodName"`
	MethodId    int64  `gorm:"column:method_id" json:"methodId"`
}

func (ApiFilter) TableName() string {
	return "d_api_filter"
}

type ApiChain struct {
	Base
	ApiId       int64 `gorm:"column:api_id" json:"apiId"`
	ReferenceId int64 `gorm:"column:reference_id" json:"referenceId"`
	MethodId    int64 `gorm:"column:method_id" json:"methodId"`
	Seq         int   `gorm:"column:seq" json:"seq"`
}

func (ApiChain) TableName() string {
	return "d_api_chain"
}

type ApiResultRule struct {
	Base
	ApiId     int64  `gorm:"column:api_id" json:"apiId"`
	ChainId   int64  `gorm:"column:chain_id" json:"chainId"`
	JavaClass string `gorm:"column:java_class" json:"javaClass"`
	Index     int    `gorm:"column:index" json:"index"`
	Rule      string `gorm:"column:rule" json:"rule"`
}

func (ApiResultRule) TableName() string {
	return "d_api_result_rule"
}

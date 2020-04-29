package entry

const (
	ChainFilter = 1
	ChainMethod = 2

	ParamMapping  = 1
	ResultMapping = 2

	Available   = 1
	UnAvailable = 2
)

type ApiConfig struct {
	Base
	UserId int64  `gorm:"column:user_id" json:"userId,omitempty"`
	Desc   string `gorm:"column:desc" json:"desc,omitempty"`
	Method string `gorm:"column:method" json:"method,omitempty"`
	Status int    `gorm:"column:status" json:"status,omitempty"`
	//UriHash  string `gorm:"column:uri_hash" json:"uriHash"`
}

func (ApiConfig) TableName() string {
	return "d_api_config"
}

type ApiChain struct {
	Base
	ApiId       int64 `gorm:"column:api_id" json:"apiId,omitempty"`
	TypeId      int   `gorm:"column:type_id" json:"typeId,omitempty"`
	ReferenceId int64 `gorm:"column:reference_id" json:"referenceId,omitempty"`
	MethodId    int64 `gorm:"column:method_id" json:"methodId,omitempty"`
	Seq         int   `gorm:"column:seq" json:"seq,omitempty"`
}

func (ApiChain) TableName() string {
	return "d_api_chain"
}

type ApiParamMapping struct {
	Base
	ApiId   int64 `gorm:"column:api_id" json:"apiId"`
	ChainId int64 `gorm:"column:chain_id" json:"chainId"`
	TypeId  int   `gorm:"column:type_id" json:"typeId,omitempty"`
	ParamId int64 `gorm:"column:param_id" json:"paramId"`
	//Path    string `gorm:"column:path" json:"path"`
	Explain string `gorm:"column:explain" json:"explain"`
}

func (ApiParamMapping) TableName() string {
	return "d_api_param_mapping"
}

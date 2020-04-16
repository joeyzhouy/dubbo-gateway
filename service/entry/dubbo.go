package entry

type Registry struct {
	Name     string `gorm:"column:name" json:"name"`
	Timeout  string `gorm:"column:time_out" json:"timeout"`
	Protocol string `gorm:"column:protocol" json:"protocol"`
	Address  string `gorm:"column:address" json:"address"`
	UserName string `gorm:"column:user_name" json:"userName"`
	Password string `gorm:"column:password" json:"password"`
	UserId   int64  `gorm:"column:user_id" json:"userId"`
	Base
}

func (Registry) TableName() string {
	return "d_registry"
}

type Reference struct {
	Base
	RegistryId    int64  `gorm:"column:registry_id" json:"registryId"`
	Protocol      string `gorm:"column:protocol" json:"protocol"`
	InterfaceName string `gorm:"column:interface_name" json:"interfaceName"`
	Cluster       string `gorm:"column:cluster" json:"cluster"`
}

func (Reference) TableName() string {
	return "d_reference"
}

type Method struct {
	Base
	ReferenceId int64  `gorm:"column:reference_id" json:"referenceId"`
	MethodName  string `gorm:"column:method_name" json:"methodName"`
}

func (Method) TableName() string {
	return "d_method"
}

type MethodParam struct {
	Base
	MethodId  int64  `gorm:"column:method_id" json:"methodId"`
	Label     string `gorm:"column:label" json:"label"`
	JavaClass string `gorm:"column:java_class" json:"javaClass"`
	Seq       int    `gorm:"column:seq" json:"seq"`
}

func (MethodParam) TableName() string {
	return "d_method_param"
}

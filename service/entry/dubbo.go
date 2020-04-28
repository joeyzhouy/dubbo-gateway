package entry

const (
	Self        = -1
	BaseType    = 1
	ComplexType = 99
	//CollectionType = 10

	MethodEntryResult = 1
	MethodEntryParam  = 2
)

type Registry struct {
	Name     string `gorm:"column:name" json:"name,omitempty"`
	Timeout  string `gorm:"column:time_out" json:"timeout,omitempty"`
	Protocol string `gorm:"column:protocol" json:"protocol,omitempty"`
	Address  string `gorm:"column:address" json:"address,omitempty"`
	UserName string `gorm:"column:user_name" json:"userName,omitempty"`
	Password string `gorm:"column:password" json:"password,omitempty"`
	UserId   int64  `gorm:"column:user_id" json:"userId,omitempty"`
	Base
}

func (Registry) TableName() string {
	return "d_registry"
}

type Reference struct {
	Base
	RegistryId    int64  `gorm:"column:registry_id" json:"registryId,omitempty"`
	Protocol      string `gorm:"column:protocol" json:"protocol,omitempty"`
	InterfaceName string `gorm:"column:interface_name" json:"interfaceName,omitempty"`
	Cluster       string `gorm:"column:cluster" json:"cluster,omitempty"`
}

func (Reference) TableName() string {
	return "d_reference"
}

type Method struct {
	Base
	ReferenceId int64  `gorm:"column:reference_id" json:"referenceId,omitempty"`
	MethodName  string `gorm:"column:method_name" json:"methodName,omitempty"`
}

func (Method) TableName() string {
	return "d_method"
}

type MethodParam struct {
	Base
	TypeId         int    `gorm:"column:type_id" json:"typeId,omitempty"`
	GenericsValues string `gorm:"column:generics_values" json:"genericsValues,omitempty"`
	MethodId       int64  `gorm:"column:method_id" json:"methodId,omitempty"`
	EntryId        int64  `gorm:"column:entry_id" json:"id,omitempty"`
	Seq            int    `gorm:"column:seq" json:"seq,omitempty"`
}

func (MethodParam) TableName() string {
	return "d_method_param"
}

//type MethodEntry struct {
//	Base
//	TypeId   int   `gorm:"column:type_id" json:"typeId"`
//	MethodId int64 `gorm:"column:method_id" json:"methodId"`
//	EntryId  int64 `gorm:"column:entry_id" json:"entryId"`
//}
//
//func (MethodEntry) TableName() string {
//	return "d_method_entry"
//}
//

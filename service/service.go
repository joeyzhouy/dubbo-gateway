package service

import (
	"dubbo-gateway/service/entry"
	"dubbo-gateway/service/vo"
)

type CommonService interface {
	GetUser(userName, password string) (*entry.User, error)
	CreateUser(user *entry.User) error
	UpdatePassword(user *entry.User, oldPassword string) error
}

type RegisterService interface {
	AddRegistryConfig(config entry.Registry) error
	DeleteRegistryConfig(registryId, userId int64) error
	ListRegistryByUser(userId int64) ([]entry.Registry, error)
	RegisterDetail(userId, registerId int64) (*entry.Registry, error)
	ListAll() ([]entry.Registry, error)
	GetRegistryByName(name string) ([]entry.Registry, error)
	GetByRegistryId(registryId int64) (*vo.Registry, error)
}

type ReferenceService interface {
	AddReference(reference entry.Reference) error
	DeleteReference(id int64) error
	ListAll() ([]entry.Reference, error)
	ListByUser(userId int64) ([]entry.Reference, error)
	GetByIds(ids []int64) ([]entry.Reference, error)
	GetReferenceById(id int64) (*vo.Reference, error)
	GetReferenceEntryById(id int64) (*entry.Reference, error)
	GetByRegistryIdAndName(registryId int64, name string) ([]entry.Reference, error)
	GetReferenceByApiId(apiId int64) ([]entry.Reference, error)
}

type MethodService interface {
	AddMethod(method *vo.Method) error
	GetMethodDetail(methodId int64) (*vo.Method, error)
	GetMethodDetailByMethod(method entry.Method) (*vo.Method, error)
	GetMethodDetailByIds(methodIds []int64) ([]*vo.Method, error)
	DeleteMethod(methodId int64) error
	GetMethodsByReferenceId(referenceId int64) ([]entry.Method, error)
	GetMethodInfoByReferenceId(referenceId int64) (*vo.ReferenceMethodInfo, error)
	SearchMethods(registryId, referenceId int64, methodName string) ([]*vo.Method, error)
	GetAllMethodDeclaration() (map[int64]*vo.MethodDeclaration, error)
	GetMethodDeclaration(methodId int64) (*vo.MethodDeclaration, error)
	GetMethodDeclarationByApiId(apiId int64) (map[int64]*vo.MethodDeclaration, error)
}

type EntryService interface {
	SaveEntry(entry *entry.EntryStructure) error
	UpdateEntry(entry *entry.EntryStructure) error
	DeleteEntry(id int64) error
	DeleteEntriesByIdsIgnoreError(ids []int64)
	GetEntry(id int64) (*entry.EntryStructure, error)
	GetEntries(ids []int64) ([]*entry.EntryStructure, error)
	GetByType(typeId int) ([]entry.Entry, error)
	SearchEntries(name string, pageSize int) ([]*entry.EntryStructure, error)
	ListAll() ([]*entry.EntryStructure, error)
}

type RouterService interface {
	AddConfig(apiConfig *vo.ApiConfigInfo) error
	UpdateConfig(apiConfig *vo.ApiConfigInfo) error
	ModifyConfigStatus(configId int64, status int) error
	DeleteConfig(configId int64) error
	GetByConfigId(configId int64) (*vo.ApiConfigInfo, error)
	ListAllAvailable() ([]*vo.ApiConfigInfo, error)
	ListAllAvailableEntry() ([]*entry.ApiConfig, error)
	GetApiMethodNamesByReferenceId(referenceId int64) ([]string, error)
	GetConfigById(configId int64) (*entry.ApiConfig, error)
	GetApiIdsByMethodId(methodId int64) ([]int64, error)
	SearchByMethodName(methodName string) ([]entry.ApiConfig, error)

	AddFilter(filter *vo.ApiFilterInfo) error
	ModifyFilter(filter *vo.ApiFilterInfo) error
	DeleterFilter(filterId int64) error
	ListFilters() ([]entry.ApiFilter, error)
	GetFilter(filterId int64) (*vo.ApiFilterInfo, error)
}

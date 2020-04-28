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
	GetByRegistryIdAndName(registryId int64, name string) ([]entry.Reference, error)
}

type MethodService interface {
	AddMethod(method *vo.Method) error
	GetMethodDetail(methodId int64) (*vo.Method, error)
	GetMethodDetailByIds(methodIds []int64) ([]*vo.Method, error)
	GetMethodDetailByMethod(method entry.Method) (*vo.Method, error)
	DeleteMethod(methodId int64) error
	GetMethodsByReferenceId(referenceId int64) ([]entry.Method, error)
	GetMethodInfoByReferenceId(referenceId int64) (*vo.ReferenceMethodInfo, error)
}

type EntryService interface {
	SaveEntry(entry *entry.EntryStructure) error
	UpdateEntry(entry *entry.EntryStructure) error

	//UpdateEntry(entry *vo.Entry) error
	DeleteEntry(id int64) error
	DeleteEntriesByIdsIgnoreError(ids []int64)
	GetEntry(id int64) (*vo.Entry, error)
	GetEntries(ids []int64) ([]*vo.Entry, error)
	GetByType(typeId int) ([]entry.Entry, error)
	SearchEntries(name string, pageSize int) ([]*vo.Entry, error)
}

type RouterService interface {
	AddRouter(api *entry.ApiConfig) error
	AddApiConfig(api *vo.ApiConfigInfo) error
	DeleteRouter(apiId int64) error
	ListRouterByUserId(userId int64) ([]entry.ApiConfig, error)
	ListAll() ([]*vo.ApiConfigInfo, error)
	GetByApiId(api int64) (*vo.ApiConfigInfo, error)
}

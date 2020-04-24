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
}

type ReferenceService interface {
	AddReference(reference entry.Reference) error
	DeleteReference(id int64) error
	ListAll() ([]entry.Reference, error)
	ListByUser(userId int64) ([]entry.Reference, error)
	GetByIds(ids []int64) ([]entry.Reference, error)
	GetReferenceById(id int64) (*vo.Reference, error)
}

type MethodService interface {
	AddMethod(method *vo.Method) error
	GetMethodDetail(methodId int64) (*vo.Method, error)
	DeleteMethod(methodId int64) error
	GetMethodsByReferenceId(referenceId int64) ([]entry.Method, error)
	//ListByUserIdAndMethodName(userId int64, methodName string) ([]vo.MethodDesc, error)
}

type EntryService interface {
	SaveEntry(entry *vo.Entry) error
	UpdateEntry(entry *vo.Entry) error
	DeleteEntry(id int64) error
	GetEntry(id int64) (*vo.Entry, error)
	GetEntries(ids []int64) ([]*vo.Entry, error)
	GetByType(typeId int) ([]entry.Entry, error)
}

type RouterService interface {
	AddRouter(api *entry.ApiConfig) error
	AddApiConfig(api *vo.ApiConfigInfo) error
	DeleteRouter(apiId int64) error
	ListRouterByUserId(userId int64) ([]entry.ApiConfig, error)

	ListAll() ([]*vo.ApiConfigInfo, error)
	GetByApiId(api int64) (*vo.ApiConfigInfo, error)
	//GetByUri(uri string) (*entry.ApiConfig, error)
}

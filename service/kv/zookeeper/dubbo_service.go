package zookeeper

import (
	"dubbo-gateway/service"
	"dubbo-gateway/service/entry"
	"dubbo-gateway/service/vo"
	"github.com/dubbogo/go-zookeeper/zk"
)

type registerService struct {
	conn *zk.Conn
}

func NewRegisterService(conn *zk.Conn) service.RegisterService {
	return &registerService{conn}
}

func (*registerService) AddRegistryConfig(config entry.Registry) error {
	panic("implement me")
}

func (*registerService) DeleteRegistryConfig(registryId, userId int64) error {
	panic("implement me")
}

func (*registerService) ListRegistryByUser(userId int64) ([]entry.Registry, error) {
	panic("implement me")
}

func (*registerService) RegisterDetail(userId, registerId int64) (*entry.Registry, error) {
	panic("implement me")
}

func (*registerService) ListAll() ([]entry.Registry, error) {
	panic("implement me")
}

type referenceService struct {
	conn *zk.Conn
}

func NewReferenceService(conn *zk.Conn) service.ReferenceService {
	return &referenceService{conn}
}

func (*referenceService) AddReference(reference entry.Reference) error {
	panic("implement me")
}

func (*referenceService) DeleteReference(id int64) error {
	panic("implement me")
}

func (*referenceService) ListAll() ([]entry.Reference, error) {
	panic("implement me")
}

func (*referenceService) ListByUser(userId int64) ([]entry.Reference, error) {
	panic("implement me")
}

func (*referenceService) GetByIds(ids []int64) ([]entry.Reference, error) {
	panic("implement me")
}

type methodService struct {
	conn *zk.Conn
}

func NewMethodService(conn *zk.Conn) service.MethodService {
	return &methodService{conn}
}

func (*methodService) AddMethod(method *vo.Method) error {
	panic("implement me")
}

func (*methodService) GetMethodDetail(methodId int64) (*vo.Method, error) {
	panic("implement me")
}

func (*methodService) DeleteMethod(methodId int64) error {
	panic("implement me")
}

func (*methodService) GetMethodsByReferenceId(referenceId int64) ([]entry.Method, error) {
	panic("implement me")
}

func (*methodService) ListByUserIdAndMethodName(userId int64, methodName string) ([]vo.MethodDesc, error) {
	panic("implement me")
}






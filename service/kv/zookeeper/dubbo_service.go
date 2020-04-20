package zookeeper

import (
	"dubbo-gateway/common/constant"
	"dubbo-gateway/service"
	"dubbo-gateway/service/entry"
	"dubbo-gateway/service/vo"
	"encoding/json"
	"fmt"
	"github.com/apache/dubbo-go/common/logger"
	"github.com/dubbogo/go-zookeeper/zk"
)

type registerService struct {
	conn *zk.Conn
}

func NewRegisterService(conn *zk.Conn, event <-chan zk.Event) service.RegisterService {
	rs := &registerService{conn}
	//TODO
	go func() {
		for {
			e := <-event
			logger.Infof("idgenerator: %v", e)
		}
	}()
	return rs
}

func (r *registerService) AddRegistryConfig(config entry.Registry) error {
	var err error
	config.ID, err = next()
	if err != nil {
		return err
	}
	nodePath := fmt.Sprintf(constant.RegistryInfoPath, config.UserId, config.ID)
	bs, err := json.Marshal(&config)
	if err != nil {
		return err
	}
	_, err = r.conn.Multi(&zk.CreateRequest{
		Path: nodePath,
		Data: bs,
		Flags:0,
		Acl: zk.WorldACL(zk.PermAll),
	},&zk.CreateRequest{
		Path: fmt.Sprintf(constant.RegistrySearch, config.ID),
		Data: []byte(nodePath),
		Flags:0,
		Acl: zk.WorldACL(zk.PermAll),
	})
	//_, err = r.conn.Create(nodePath, bs, 0, zk.WorldACL(zk.PermAll))
	return err
}

func (r *registerService) DeleteRegistryConfig(registryId, userId int64) error {
	nodePath := fmt.Sprintf(constant.RegistryInfoPath, userId, registryId)
	return DeleteAll(nodePath, r.conn)
}

func (r *registerService) ListRegistryByUser(userId int64) ([]entry.Registry, error) {
	userPath := fmt.Sprintf(constant.RegistryPath, userId)
	children, _, err := r.conn.Children(userPath)
	if err != nil {
		return nil, err
	}
	result := make([]entry.Registry, 0)
	for _, child := range children {
		bs, _, err := r.conn.Get(child)
		if err != nil {
			return nil, err
		}
		registry := new(entry.Registry)
		err = json.Unmarshal(bs, registry)
		if err != nil {
			return nil, err
		}
		result = append(result, *registry)
	}
	return result, nil
}

func (r *registerService) RegisterDetail(userId, registerId int64) (*entry.Registry, error) {
	nodePath := fmt.Sprintf(constant.RegistryInfoPath, userId, registerId)
	bs, _, err := r.conn.Get(nodePath)
	if err != nil {
		return nil, err
	}
	result := new(entry.Registry)
	err = json.Unmarshal(bs, result)
	return result, err
}

func (r *registerService) ListAll() ([]entry.Registry, error) {
	children, _, err := r.conn.Children(constant.RegistrySearch)
	if err != nil {
		return nil, err
	}
	result := make([]entry.Registry, 0)
	for _, child := range children {
		bs, _, err := r.conn.Get(child)
		if err != nil {
			return nil, err
		}
		registry := new(entry.Registry)
		err = json.Unmarshal(bs, registry)
		if err != nil {
			return nil, err
		}
		result = append(result, *registry)
	}
	return result, err
}

type referenceService struct {
	conn *zk.Conn
}

func NewReferenceService(conn *zk.Conn, event <-chan zk.Event) service.ReferenceService {
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

func NewMethodService(conn *zk.Conn, event <-chan zk.Event) service.MethodService {
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

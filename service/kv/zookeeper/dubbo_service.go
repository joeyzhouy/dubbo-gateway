package zookeeper

import (
	"dubbo-gateway/common/constant"
	"dubbo-gateway/service"
	"dubbo-gateway/service/entry"
	"dubbo-gateway/service/vo"
	"encoding/json"
	"fmt"
	"github.com/dubbogo/go-zookeeper/zk"
	perrors "github.com/pkg/errors"
	"regexp"
	"strconv"
	"time"
)

var registerReg = regexp.MustCompile(constant.UserPath + `/\d/` + constant.Registries + `/(\d+)`)

type registerService struct {
	conn *zk.Conn
}

func NewRegisterService(conn *zk.Conn, event <-chan zk.Event) service.RegisterService {
	rs := &registerService{conn}
	if err := CreateBasePath(constant.RegistrySearchRoot, rs.conn); err != nil {
		panic("register service init error: " + err.Error())
	}
	initIdGenerator(conn, event)
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
		Path:  nodePath,
		Data:  bs,
		Flags: 0,
		Acl:   zk.WorldACL(zk.PermAll),
	}, &zk.CreateRequest{
		Path:  fmt.Sprintf(constant.RegistrySearch, config.ID),
		Data:  []byte(nodePath),
		Flags: 0,
		Acl:   zk.WorldACL(zk.PermAll),
	}, zk.CreateRequest{
		Path:  nodePath + constant.Registries,
		Data:  emptyValue,
		Flags: 0,
		Acl:   zk.WorldACL(zk.PermAll),
	})
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

func (r *referenceService) GetReferenceById(id int64) (*vo.Reference, error) {
	return nil, nil
}

func NewReferenceService(conn *zk.Conn, event <-chan zk.Event) service.ReferenceService {
	reference := &referenceService{conn}
	if err := CreateBasePath(constant.ReferenceSearchRoot, reference.conn); err != nil {
		panic("reference service init error: " + err.Error())
	}
	return reference
}

func (r *referenceService) AddReference(reference entry.Reference) error {
	bs, _, err := r.conn.Get(fmt.Sprintf(constant.RegistrySearch, reference.RegistryId))
	if err != nil {
		return err
	}
	temps := userPathReg.FindStringSubmatch(string(bs))
	if len(temps) == 0 {
		return perrors.Errorf("not find userId in path: %s", string(bs))
	}
	id, err := next()
	if err != nil {
		return err
	}
	userId, err := strconv.ParseInt(temps[0], 10, 64)
	if err != nil {
		return err
	}
	reference.ID = id
	nodePath := fmt.Sprintf(constant.ReferenceInfoPath, userId, reference.RegistryId, id)
	bs, err = json.Marshal(&reference)
	if err != nil {
		return err
	}
	_, err = r.conn.Multi(&zk.CreateRequest{
		Path:  nodePath,
		Data:  bs,
		Flags: 0,
		Acl:   zk.WorldACL(zk.PermAll),
	}, &zk.CreateRequest{
		Path:  fmt.Sprintf(constant.RegistrySearch, reference.ID),
		Data:  []byte(nodePath),
		Flags: 0,
		Acl:   zk.WorldACL(zk.PermAll),
	}, &zk.CreateRequest{
		Path:  nodePath + constant.Methods,
		Data:  emptyValue,
		Flags: 0,
		Acl:   zk.WorldACL(zk.PermAll),
	})
	return err
}

func (r *referenceService) DeleteReference(id int64) error {
	bs, _, err := r.conn.Get(fmt.Sprintf(constant.ReferenceSearch, id))
	if err != nil {
		return err
	}
	registryPath := string(bs)
	opts, err := deleteOperation(registryPath, r.conn)
	if err != nil {
		return err
	}
	opts = append(opts, &zk.DeleteRequest{
		Path:    registryPath,
		Version: -1,
	})
	_, err = r.conn.Multi(opts...)
	return err
}

func (r *referenceService) ListAll() ([]entry.Reference, error) {
	children, _, err := r.conn.Children(constant.ReferenceSearch)
	if err != nil {
		return nil, err
	}
	result := make([]entry.Reference, 0)
	for _, child := range children {
		bs, _, err := r.conn.Get(child)
		if err != nil {
			return nil, err
		}
		reference := new(entry.Reference)
		err = json.Unmarshal(bs, reference)
		if err != nil {
			return nil, err
		}
		result = append(result, *reference)
	}
	return result, nil
}

func (r *referenceService) ListByUser(userId int64) ([]entry.Reference, error) {
	references, _, err := r.conn.Children(constant.ReferenceSearchRoot)
	if err != nil {
		return nil, err
	}
	prefix := fmt.Sprintf(constant.UserInfoPath, userId) + "/"
	prefixLength := len(prefix)
	result := make([]entry.Reference, 0)
	for _, referencePath := range references {
		if prefixLength > len(referencePath) {
			if prefix == string([]rune(referencePath)[:prefixLength]) {
				bs, _, err := r.conn.Get(referencePath)
				if err != nil {
					return nil, err
				}
				reference := new(entry.Reference)
				err = json.Unmarshal(bs, reference)
				if err != nil {
					return nil, err
				}
				result = append(result, *reference)
			}
		}
	}
	return result, nil
}

func (r *referenceService) GetByIds(ids []int64) ([]entry.Reference, error) {
	result := make([]entry.Reference, 0, len(ids))
	for _, id := range ids {
		searchPath := fmt.Sprintf(constant.ReferenceSearch, id)
		bs, _, err := r.conn.Get(searchPath)
		if err != nil {
			return nil, err
		}
		realPath := string(bs)
		bs, _, err = r.conn.Get(realPath)
		if err != nil {
			return nil, err
		}
		reference := new(entry.Reference)
		err = json.Unmarshal(bs, reference)
		if err != nil {
			return nil, err
		}
		result = append(result, *reference)
	}
	return result, nil
}

type methodService struct {
	conn *zk.Conn
}

func NewMethodService(conn *zk.Conn, event <-chan zk.Event) service.MethodService {
	ms := &methodService{conn}
	if err := CreateBasePath(constant.MethodSearchRoot, ms.conn); err != nil {
		panic("method service init error: " + err.Error())
	}
	return ms
}

func (m *methodService) AddMethod(method *vo.Method) error {
	registrySearchPath := fmt.Sprintf(constant.ReferenceSearch, method.ReferenceId)
	bs, _, err := m.conn.Get(registrySearchPath)
	if err != nil {
		return err
	}
	registryInfoPath := string(bs)
	current := time.Now()
	method.CreateTime = current
	method.ModifyTime = current
	method.ID, err = next()
	if err != nil {
		return err
	}
	methodPath := fmt.Sprintf(registryInfoPath+constant.Methods+"/%d", method.ID)
	bs, err = json.Marshal(method)
	if err != nil {
		return err
	}
	_, err = m.conn.Multi(&zk.CreateRequest{
		Path:  methodPath,
		Flags: 0,
		Data:  bs,
		Acl:   zk.WorldACL(zk.PermAll),
	}, &zk.CreateRequest{
		Path:  fmt.Sprintf(constant.MethodSearch, method.ID),
		Data:  []byte(methodPath),
		Flags: 0,
		Acl:   zk.WorldACL(zk.PermAll),
	})
	return err
}

func (m *methodService) GetMethodDetail(methodId int64) (*vo.Method, error) {
	methodSearchPath := fmt.Sprintf(constant.MethodSearch, methodId)
	bs, _, err := m.conn.Get(methodSearchPath)
	if err != nil {
		return nil, err
	}
	methodInfoPath := string(bs)
	bs, _, err = m.conn.Get(methodInfoPath)
	if err != nil {
		return nil, err
	}
	result := new(vo.Method)
	err = json.Unmarshal(bs, result)
	return result, err
}

func (m *methodService) DeleteMethod(methodId int64) error {
	methodSearchPath := fmt.Sprintf(constant.MethodSearch, methodId)
	bs, _, err := m.conn.Get(methodSearchPath)
	if err != nil {
		return err
	}
	methodInfoPath := string(bs)
	_, err = m.conn.Multi(&zk.DeleteRequest{
		Path:    methodInfoPath,
		Version: -1,
	}, &zk.DeleteRequest{
		Path:    methodInfoPath,
		Version: -1,
	})
	return err
}

func (m *methodService) GetMethodsByReferenceId(referenceId int64) ([]entry.Method, error) {
	referenceSearchPath := fmt.Sprintf(constant.ReferenceSearch, referenceId)
	bs, _, err := m.conn.Get(referenceSearchPath)
	if err != nil {
		return nil, err
	}
	referenceInfoPath := string(bs)
	methods := referenceInfoPath + constant.Methods
	methodPaths, _, err := m.conn.Children(methods)
	if err != nil {
		return nil, err
	}
	result := make([]entry.Method, 0, len(methodPaths))
	for _, methodPath := range methodPaths {
		bs, _, err = m.conn.Get(methodPath)
		if err != nil {
			return nil, err
		}
		method := new(entry.Method)
		err = json.Unmarshal(bs, method)
		if err != nil {
			return nil, err
		}
		result = append(result, *method)
	}
	return result, nil
}

//func (*methodService) ListByUserIdAndMethodName(userId int64, methodName string) ([]vo.MethodDesc, error) {
//	panic("implement me")
//}

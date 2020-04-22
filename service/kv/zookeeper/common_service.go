package zookeeper

import (
	"dubbo-gateway/common/constant"
	"dubbo-gateway/common/utils"
	"dubbo-gateway/service"
	"dubbo-gateway/service/entry"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"github.com/apache/dubbo-go/common/logger"
	"github.com/dubbogo/go-zookeeper/zk"
	perrors "github.com/pkg/errors"
	"path"
	"regexp"
	"strconv"
	"strings"
)

var idValue = []byte("id")
var idg *idGenerator
var userPathReg = regexp.MustCompile(constant.UserPath + `/(\d+)`)

type commonService struct {
	conn *zk.Conn
}

func (c *commonService) CreateUser(user *entry.User) error {
	var err error
	user.ID, err = next()
	if err != nil {
		return err
	}
	user.Password = utils.Sha256(user.Password, user.Name)
	bs, err := json.Marshal(user)
	if err != nil {
		return err
	}
	_, err = c.conn.Multi(&zk.CreateRequest{
		Path:  fmt.Sprintf(constant.UserInfoPath, user.ID),
		Data:  bs,
		Flags: 0,
		Acl:   zk.WorldACL(zk.PermAll),
	}, &zk.CreateRequest{
		Path:  fmt.Sprintf(constant.UserNamePath, user.Name),
		Data:  Int64ToBytes(user.ID),
		Flags: 0,
		Acl:   zk.WorldACL(zk.PermAll),
	})
	return err
}

func NewCommonService(conn *zk.Conn, event <-chan zk.Event) service.CommonService {
	cs := &commonService{conn}
	if err := CreateBasePath(constant.UserPath, cs.conn); err != nil {
		panic("common service init error: " + err.Error())
	}
	initIdGenerator(conn, event)
	return cs
}

func (c *commonService) GetUser(userName, password string) (*entry.User, error) {

	if userName != constant.DefaultUserName {
		return nil, constant.UserOrPasswordError
	}
	user, _, err := c.getUser(userName, password)
	return user, err
}

func (c *commonService) getUser(userName, password string) (*entry.User, *zk.Stat, error) {
	bs, stat, err := c.conn.Get(fmt.Sprintf(constant.UserNamePath, userName))
	if err != nil {
		return nil, nil, err
	}
	user := new(entry.User)
	err = json.Unmarshal(bs, user)
	if err != nil {
		return nil, nil, err
	}
	if utils.Sha256(password, userName) != user.Password {
		return nil, nil, constant.UserOrPasswordError
	}
	return user, stat, nil
}

func (c *commonService) UpdatePassword(user *entry.User, oldPassword string) error {
	if user.Name != constant.DefaultUserName {
		return constant.UserOrPasswordError
	}
	user, _, err := c.getUser(user.Name, oldPassword)
	if err != nil {
		return err
	}
	user.Password = utils.Sha256(user.Password, user.Name)
	bs, err := json.Marshal(user)
	if err != nil {
		return err
	}
	_, err = c.conn.Set(constant.UserInfoPath, bs, -1)
	return err
}

func initIdGenerator(conn *zk.Conn, event <-chan zk.Event) {
	idg = &idGenerator{conn}
	ok, _, err := idg.conn.Exists(constant.IdPath)
	if err != nil {
		logger.Errorf("exist path: %s , error: %v", constant.IdPath, err)
	}
	if !ok {
		_, err := idg.conn.Create(constant.IdPath, idValue, 0, zk.WorldACL(zk.PermAll))
		if err != nil {
			panic(fmt.Sprintf("init zookeeper path: %s, error: %v", constant.IdPath, err))
		}
	}
	go func() {
		for {
			e := <-event
			logger.Infof("idgenerator: %v", e)
		}
	}()
}

type idGenerator struct {
	conn *zk.Conn
}

func next() (int64, error) {
	stat, err := idg.conn.Set(constant.IdPath, idValue, -1)
	if err != nil {
		return -1, err
	}
	return int64(stat.Version), nil
}

func nextN(n int) ([]int64, error) {
	requests := make([]interface{}, 0, n)
	for i := 0; i < n; i++ {
		requests = append(requests, &zk.SetDataRequest{
			Data:    idValue,
			Path:    constant.IdPath,
			Version: -1,
		})
	}
	results, err := idg.conn.Multi(requests...)
	if err != nil {
		return nil, err
	}
	ids := make([]int64, 0, n)
	for _, result := range results {
		ids = append(ids, int64(result.Stat.Version))
	}
	return ids, nil
}

func CreateBasePath(basePath string, conn *zk.Conn) error {
	var temp string
	for _, subPath := range strings.Split(basePath, "/") {
		temp = path.Join(temp, "/", subPath)
		_, err := conn.Create(temp, []byte(""), 0, zk.WorldACL(zk.PermAll))
		if err != nil {
			if err == zk.ErrNodeExists {
				logger.Infof("zk.create(\"%s\") exists\n", temp)
			} else {
				logger.Errorf("zk.create(\"%s\") error(%v)\n", temp, perrors.WithStack(err))
				return perrors.WithMessagef(err, "zk.Create(path:%s)", basePath)
			}
		}
	}
	return nil
}

func DeleteAll(basePath string, conn *zk.Conn) error {
	ops, err := deleteOperation(basePath, conn)
	if err != nil {
		return err
	}
	_, err = conn.Multi(ops)
	return err
}

var searchMap = map[string]string{
	string([]rune(constant.Registries)[1:]): constant.RegistrySearch,
	string([]rune(constant.References)[1:]): constant.ReferenceSearch,
}

func deleteOperation(basePath string, conn *zk.Conn) ([]interface{}, error) {
	ops := make([]interface{}, 0)
	children, _, err := conn.Children(basePath)
	if err != nil {
		return nil, err
	}
	if len(children) == 0 {
		ops = append(ops, &zk.DeleteRequest{
			Path:    basePath,
			Version: -1,
		})
		arrs := strings.Split(basePath, "/")
		length := len(arrs)
		if length <= 2 {
			return ops, nil
		}
		searchPath, ok := searchMap[arrs[length-2]]
		if !ok {
			return ops, nil
		}
		index, err := strconv.ParseInt(arrs[length-1], 10, 64)
		if err != nil {
			return ops, nil
		}
		ops = append(ops, &zk.DeleteRequest{
			Path:    fmt.Sprintf(searchPath, index),
			Version: -1,
		})
		return ops, nil
	}
	for _, child := range children {
		if temp, err := deleteOperation(child, conn); err != nil {
			return nil, err
		} else {
			ops = append(ops, temp...)
		}
	}
	return ops, nil
}

func Int64ToBytes(i int64) []byte {
	var buf = make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(i))
	return buf
}

func BytesToInt64(buf []byte) int64 {
	return int64(binary.BigEndian.Uint64(buf))
}

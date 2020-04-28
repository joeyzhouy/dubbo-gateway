package zookeeper

import (
	"dubbo-gateway/common/extension"
	"dubbo-gateway/meta"
	"dubbo-gateway/service"
	"dubbo-gateway/service/kv/zookeeper"
	"github.com/apache/dubbo-go/common/logger"
	"github.com/dubbogo/go-zookeeper/zk"
	"gopkg.in/yaml.v2"
	"strings"
	"sync"
	"time"
)

const (
	zkey = "zookeeper"
)

var zkMeta *metaZookeeper
var initError error
var once sync.Once

func init() {
	extension.SetMeta(zkey, NewZookeeperMeta)
}

type zookeeperConfig struct {
	Addresses string        `yaml:"addresses"`
	UserName  string        `yaml:"username"`
	Password  string        `yaml:"password"`
	TimeOut   time.Duration `yaml:"timeout"`
}

type metaZookeeper struct {
	//Conn   *zk.Conn
	Config *zookeeperConfig
}

func (z *metaZookeeper) NewEntryService() service.EntryService {
	panic("implement me")
}

func (z *metaZookeeper) createZkConn() (*zk.Conn, <-chan zk.Event, error) {
	conn, ch, err := zk.Connect(strings.Split(z.Config.Addresses, ","), z.Config.TimeOut)
	return conn, ch, err
}

func (z *metaZookeeper) NewCommonService() service.CommonService {
	if conn, ch, err := z.createZkConn(); err != nil {
		logger.Errorf("create meta[zookeeper] error: %v", err)
		return nil
	} else {
		return zookeeper.NewCommonService(conn, ch)
	}
}

func (z *metaZookeeper) NewRouterService() service.RouterService {
	if conn, ch, err := z.createZkConn(); err != nil {
		logger.Errorf("create meta[zookeeper] error: %v", err)
		return nil
	} else {
		return zookeeper.NewRouterService(conn, ch)
	}
}

func (z *metaZookeeper) NewReferenceService() service.ReferenceService {
	if conn, ch, err := z.createZkConn(); err != nil {
		logger.Errorf("create meta[zookeeper] error: %v", err)
		return nil
	} else {
		return zookeeper.NewReferenceService(conn, ch)
	}
}

func (z *metaZookeeper) NewRegisterService() service.RegisterService {
	if conn, ch, err := z.createZkConn(); err != nil {
		logger.Errorf("create meta[zookeeper] error: %v", err)
		return nil
	} else {
		return zookeeper.NewRegisterService(conn, ch)
	}
}

func (z *metaZookeeper) NewMethodService() service.MethodService {
	if conn, ch, err := z.createZkConn(); err != nil {
		logger.Errorf("create meta[zookeeper] error: %v", err)
		return nil
	} else {
		return zookeeper.NewMethodService(conn, ch)
	}
}

func NewZookeeperMeta(configString string) (meta.Meta, error) {
	once.Do(func() {
		config := new(zookeeperConfig)
		initError = yaml.Unmarshal([]byte(configString), config)

	})
	return zkMeta, initError
}

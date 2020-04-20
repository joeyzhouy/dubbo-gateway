package zookeeper

import (
	"dubbo-gateway/common/extension"
	"dubbo-gateway/meta"
	"dubbo-gateway/service"
	"dubbo-gateway/service/kv/zookeeper"
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
	Conn   *zk.Conn
	Config *zookeeperConfig
}

func (z *metaZookeeper) NewCommonService() service.CommonService {
	return zookeeper.NewCommonService(z.Conn)
}

func (z *metaZookeeper) NewRouterService() service.RouterService {
	return zookeeper.NewRouterService(z.Conn)
}

func (z *metaZookeeper) NewReferenceService() service.ReferenceService {
	return zookeeper.NewReferenceService(z.Conn)
}

func (z *metaZookeeper) NewRegisterService() service.RegisterService {
	return zookeeper.NewRegisterService(z.Conn)
}

func (z *metaZookeeper) NewMethodService() service.MethodService {
	return zookeeper.NewMethodService(z.Conn)
}

func NewZookeeperMeta(configString string) (meta.Meta, error) {
	once.Do(func() {
		config := new(zookeeperConfig)
		initError = yaml.Unmarshal([]byte(configString), config)
		zkMeta = &metaZookeeper{Config: config}
		zkMeta.Conn, _, initError = zk.Connect(strings.Split(config.Addresses, ","), config.TimeOut)

	})
	return zkMeta, initError
}

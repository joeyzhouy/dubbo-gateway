package extension

import (
	"dubbo-gateway/common/config"
	"github.com/apache/dubbo-go/common/logger"
	"github.com/go-errors/errors"
	"sync"
)

var discover = make(map[string]func(config config.DiscoverConfig) (Discover, error))
var discoverManager = make(map[string]Discover)
var mutex sync.Mutex

type Node struct {
	FullPath string
	BasePath string
	SubPath  string
}

type Discover interface {
	GetChildrenMethod(interfaceName string) ([]Node, error)
	GetChildrenInterface() ([]Node, error)
	Close()
}

func SetDiscover(protocol string, f func(config config.DiscoverConfig) (Discover, error)) {
	discover[protocol] = f
}

func RemoveDisCovert(conf config.DiscoverConfig) error {
	mutex.Lock()
	defer mutex.Unlock()
	key := conf.GetKey()
	if dis, ok := discoverManager[key]; !ok {
		return nil
	} else {
		delete(discoverManager, key)
		logger.Info("close discover: => " + key)
		dis.Close()
		return nil
	}
}

func GetDiscover(conf config.DiscoverConfig) (Discover, error) {
	mutex.Lock()
	defer mutex.Unlock()
	key := conf.GetKey()
	if dis, ok := discoverManager[key]; ok {
		return dis, nil
	}
	f, ok := discover[conf.Protocol]
	if !ok {
		return nil, errors.New("not support protocol")
	}
	dis, err := f(conf)
	if err == nil {
		discoverManager[key] = dis
	}
	return dis, err
}

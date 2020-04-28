package zookeeper

import (
	"dubbo-gateway/common"
	"dubbo-gateway/common/config"
	"dubbo-gateway/common/constant"
	"dubbo-gateway/common/extension"
	"dubbo-gateway/registry"
	"fmt"
	constants "github.com/apache/dubbo-go/common/constant"
	"github.com/apache/dubbo-go/common/logger"
	"github.com/dubbogo/go-zookeeper/zk"
	perrors "github.com/pkg/errors"
	"strconv"
	"strings"
	"sync"
	"time"
)

func init() {
	extension.SetRegistry(constant.ProtocolZookeeper, newZkRegistry)
	extension.SetDiscover(constant.ProtocolZookeeper, NewZkDiscover)
}

type zkRegistry struct {
	cli   *zkClient
	zLock sync.Mutex
}

func (z *zkRegistry) Close() {
	if z.cli != nil {
		z.cli.Close()
	}
}

func (z *zkRegistry) ListNodeByPath(path string) ([]common.Node, error) {
	children, _, err := z.cli.Conn.Children(path)
	if err != nil {
		return nil, err
	}
	result := make([]common.Node, 0, len(children))
	for index, childPath := range children {
		address := strings.Split(string([]rune(childPath)[strings.LastIndex(childPath, "/")+1:]), ":")
		port, err := strconv.Atoi(address[1])
		if err != nil {
			return nil, err
		}
		result[index] = common.Node{
			IP:   address[0],
			Port: port,
		}
	}
	return result, err
}

func (z *zkRegistry) RegisterTempNode(node common.Node) error {
	z.zLock.Lock()
	defer z.zLock.Unlock()
	err := z.cli.CreateBasePath(constant.NodePath)
	if err != nil {
		return err
	}
	_, err = z.cli.Conn.CreateProtectedEphemeralSequential(constant.NodePath+"/"+node.String(),
		[]byte(""), zk.WorldACL(zk.PermAll))
	if err != nil {
		return perrors.Errorf("create temp node, Path: %s, error: %v", constant.NodePath+"/"+node.String(), err)
	}
	return nil
}

func (z *zkRegistry) Subscribe(path string, listener registry.NotifyListener) error {
	ch := make(chan struct{})
	z.cli.RegisterEvent(path, &ch)
	go func(c *chan struct{}, listener registry.NotifyListener) {
		for {
			<-*c
			listener(&registry.Event{Path: path, Action: registry.EventChildrenChange})
		}
	}(&ch, listener)
	return nil
}

func newZkRegistry(deploy config.Deploy) (registry.Registry, error) {
	config := deploy.Config.Multiple.Coordination
	zkAddresses := make([]string, 0)
	for _, str := range strings.Split(config.Address, ",") {
		zkAddresses = append(zkAddresses, strings.TrimSpace(str))
	}
	cli, err := newZkClient(zkAddresses, config.Timeout)
	if err != nil {
		return nil, perrors.WithMessagef(err, "zk.Connect(zkAddrs:%+v)", zkAddresses)
	}
	ry := &zkRegistry{cli: cli}
	return ry, nil
}

func NewZkDiscover(conf config.DiscoverConfig) (extension.Discover, error) {
	zkAddresses := make([]string, 0)
	for _, str := range strings.Split(conf.Address, ",") {
		zkAddresses = append(zkAddresses, strings.TrimSpace(str))
	}
	d, err := time.ParseDuration(conf.Timeout)
	if err != nil {
		return nil, perrors.WithMessagef(err, "parse duration error: %s", conf.Timeout)
	}
	cli, event, err := newZkClientWithEvent(zkAddresses, d)
	if err != nil {
		return nil, perrors.WithMessagef(err, "zk.Connect(zkAddrs:%+v)", zkAddresses)
	}
	ry := &zkRegistry{cli: cli}
	go func() {
		times := 5
		index := times
	LOOP:
		for {
			e := <-event
			switch e.State {
			case zk.StateDisconnected:
				if err := extension.RemoveDisCovert(conf); err != nil {
					logger.Errorf("remove zk{addr:%s}, error :%v", zkAddresses, err)
				}
				break LOOP
			case zk.StateConnecting:
				if index > 0 {
					logger.Warnf("try to connect zk{addr:%s)", zkAddresses)
					index--
				} else {
					logger.Warnf("try [%d] times, to connect zk{addr:%s) failed", times, zkAddresses)
					if err := extension.RemoveDisCovert(conf); err != nil {
						logger.Errorf("remove zk{addr:%s}, error :%v", zkAddresses, err)
					}
					break LOOP
				}
			case zk.StateConnected:
				index = times
				logger.Infof("zk{addr:%s} already success connected", zkAddresses)
			}
		}
	}()

	return ry, nil
}

func (z *zkRegistry) GetChildNode(basePath string) ([]extension.Node, error) {
	fmt.Println(basePath)
	children, _, err := z.cli.Conn.Children(basePath)
	if err != nil {
		return nil, err
	}
	result := make([]extension.Node, 0, len(children))
	for _, path := range children {
		result = append(
			result,
			extension.Node{FullPath: basePath + "/" + path, BasePath: basePath, SubPath: path},
		)
	}
	return result, nil
}

func (z *zkRegistry) GetChildrenMethod(interfaceName string) ([]extension.Node, error) {
	return z.GetChildNode("/" + constants.DUBBO + "/" + interfaceName + "/" + constants.PROVIDER_CATEGORY)
}

func (z *zkRegistry) GetChildrenInterface() ([]extension.Node, error) {
	return z.GetChildNode("/" + constants.DUBBO)
}

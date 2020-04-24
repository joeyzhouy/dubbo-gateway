package zookeeper

import (
	"dubbo-gateway/common"
	"dubbo-gateway/common/config"
	"dubbo-gateway/common/constant"
	"dubbo-gateway/common/extension"
	"dubbo-gateway/registry"
	"github.com/dubbogo/go-zookeeper/zk"
	perrors "github.com/pkg/errors"
	"strconv"
	"strings"
	"sync"
)


func init() {
	extension.SetRegistry(constant.ProtocolZookeeper, newZkRegistry)
}

type zkRegistry struct {
	cli   *zkClient
	zLock sync.Mutex
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

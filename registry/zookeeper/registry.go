package zookeeper

import (
	"dubbo-gateway/common/extension"
	"dubbo-gateway/registry"
	perrors "github.com/pkg/errors"
	"strings"
)

const Zookeeper = "zookeeper"

func init() {
	extension.SetRegistry(Zookeeper, newZkRegistry)
}

type zkRegistry struct {
	Cli *zkClient
}

func newZkRegistry(node *extension.Node) (registry.Registry, error) {
	config := extension.GetDeployConfig().Config.Multiple.Coordination
	zkAddresses := make([]string, 0)
	for _, str := range strings.Split(config.Address, ",") {
		zkAddresses = append(zkAddresses, strings.TrimSpace(str))
	}
	cli, err := newZkClient(zkAddresses, config.Timeout)
	if err != nil {
		return nil, perrors.WithMessagef(err, "zk.Connect(zkAddrs:%+v)", zkAddresses)
	}
	ry := &zkRegistry{Cli: cli}
	return ry, nil
}

func (*zkRegistry) Registry(extension.Node) error {
	panic("implement me")
}

func (*zkRegistry) Subscribe(extension.Node, registry.NotifyListener) {
	panic("implement me")
}

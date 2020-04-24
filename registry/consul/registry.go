package consul

import (
	"dubbo-gateway/common"
	"dubbo-gateway/common/config"
	//"dubbo-gateway/common/extension"
	"dubbo-gateway/registry"
	//"github.com/hashicorp/consul/agent/consul"
	consul "github.com/hashicorp/consul/api"
)

const Consul = "consul"

func init() {
	//extension.SetRegistry(Consul, NewConsulRegistry)
}

func NewConsulRegistry(deploy config.Deploy) (registry.Registry, error) {
	consulConfig := &consul.Config{Address: deploy.Config.Multiple.Coordination.Address}
	client, err := consul.NewClient(consulConfig)
	if err != nil {
		return nil, err
	}
	return &consulRegistry{
		client: client,
	}, nil
}

type consulRegistry struct {
	client *consul.Client
}

func (c *consulRegistry) RegisterTempNode(node common.Node) error {
	panic("implement me")
}

func (c *consulRegistry) Subscribe(path string, listener registry.NotifyListener) error {
	panic("implement me")
}

func (c *consulRegistry) ListNodeByPath(path string) ([]common.Node, error) {
	panic("implement me")
}

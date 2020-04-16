package registry

import (
	"dubbo-gateway/common"
)

type Registry interface {
	RegisterTempNode(node common.Node) error

	Subscribe(path string, listener NotifyListener) error

	ListNodeByPath(path string) ([]common.Node, error)
}

type NotifyListener func(*Event)


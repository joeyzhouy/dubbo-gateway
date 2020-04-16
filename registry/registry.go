package registry

import "dubbo-gateway/common/extension"

type Registry interface {
	RegisterTempNode(node extension.Node) error

	Subscribe(path string, listener NotifyListener) error

	ListNodeByPath(path string) ([]extension.Node, error)
}

type NotifyListener func(*Event)


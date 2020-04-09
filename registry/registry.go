package registry

import "dubbo-gateway/common/extension"

type Registry interface {

	Registry(extension.Node) error

	Subscribe(extension.Node, NotifyListener)
}

type NotifyListener interface {
	Notify(event *Event)
}


package extension

import "dubbo-gateway/registry"

var registrys = make(map[string]func(node *Node) (registry.Registry, error))

func SetRegistry(name string, v func(node *Node) (registry.Registry, error)) {
	registrys[name] = v
}

func GetRegistry(name string, node *Node) (registry.Registry, error) {
	if registrys[name] == nil {
		panic("registry for " + name + " is not existing, make sure you have import the package.")
	}
	return registrys[name](node)
}

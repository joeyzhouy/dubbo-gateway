package extension

import "dubbo-gateway/registry"

var registrys = make(map[string]func(deploy Deploy) (registry.Registry, error))

func SetRegistry(name string, v func(deploy Deploy) (registry.Registry, error)) {
	registrys[name] = v
}

func GetRegistry(name string) (registry.Registry, error) {
	if registrys[name] == nil {
		panic("registry for " + name + " is not existing, make sure you have import the package.")
	}
	return registrys[name](*GetDeployConfig())
}

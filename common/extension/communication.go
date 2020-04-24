package extension

import "dubbo-gateway/common/config"

type Mode interface {
	Start()
	Add(apiId int64) error
	Remove(apiId int64) error
	Refresh() error
	Close()
}

var modes = make(map[string]func(deploy *config.Deploy) (Mode, error))

func GetMode(mode string) (Mode, error) {
	if modes[mode] == nil {
		panic("mode for " + mode + " is not existing, make sure you have import the package.")
	}
	return modes[mode](config.GetDeployConfig())
}

func SetMode(mode string, v func(deploy *config.Deploy) (Mode, error)) {
	modes[mode] = v
}

func GetConfigMode() (Mode, error) {
	return GetMode(config.GetDeployConfig().Config.Model)
}

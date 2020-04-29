package extension

import "dubbo-gateway/common/config"

const (
	Add EventType = iota
	Modify
	Delete
)

const (
	Entry Domain = iota
	Method
	Api
)

type Mode interface {
	Start()
	Init() error
	//Add(apiId int64) error
	//Remove(apiId int64) error
	//Refresh() error
	Notify(event ModeEvent)
	Close()
}

type EventType int
type Domain int

type ModeEvent struct {
	Type   EventType
	Domain Domain
	Key    int64
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

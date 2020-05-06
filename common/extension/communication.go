package extension

import (
	"dubbo-gateway/common"
	"dubbo-gateway/common/config"
	"sync"
)

const (
	AllType EventType = iota
	Add
	Modify
	Delete
)

const (
	Method Domain = iota
	Api
	Registry
	Reference
)

type Mode interface {
	common.GatewayCache
	Start()
	Notify(event ModeEvent)
	SubscribeEvent(domain Domain, eventType EventType, identify string, f func(event ModeEvent)) error
	UnsubscribeEvent(domain Domain, eventType EventType, identify string)
	Close()
}

type EventType int
type Domain int

type ModeEvent struct {
	Type        EventType
	Domain      Domain
	Key         int64
	Attachments map[string]string
}

var modes = make(map[string]func(deploy *config.Deploy) (Mode, error))
var onceMode sync.Once
var configMode Mode

func GetMode(mode string) Mode {
	if modes[mode] == nil {
		panic("mode for " + mode + " is not existing, make sure you have import the package.")
	}
	m, err := modes[mode](config.GetDeployConfig())
	if err != nil {
		panic("create mode[" + mode + "] error: " + err.Error())
	}
	return m
}

func SetMode(mode string, v func(deploy *config.Deploy) (Mode, error)) {
	modes[mode] = v
}

func GetConfigMode() Mode {
	onceMode.Do(func() {
		configMode = GetMode(config.GetDeployConfig().Config.Model)
	})
	return configMode
}

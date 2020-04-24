package extension

import (
	"dubbo-gateway/common/config"
	"dubbo-gateway/meta"
	"gopkg.in/yaml.v2"
	"sync"
)

var metaDataMap = make(map[string]func(configString string) (meta.Meta, error))
var once sync.Once
var m meta.Meta

func SetMeta(key string, f func(configString string) (meta.Meta, error)) {
	metaDataMap[key] = f
}

func GetMeta() meta.Meta {
	once.Do(func() {
		metaConfig := config.GetMetaConfig()
		if len(metaConfig.Config) == 0 {
			panic("miss meta config")
		} else if len(metaConfig.Config) > 1 {
			panic("multiple meta config")
		}
		var key string
		var value interface{}
		for k, v := range metaConfig.Config {
			key, value = k, v
		}
		f, ok := metaDataMap[key]
		if !ok {
			panic("meta for " + key + " is not existing, make sure you have import the package.")
		}
		bs, err := yaml.Marshal(value)
		if err != nil {
			panic("sub meta config marshall error: " + err.Error())
		}
		m, err = f(string(bs))
		if err != nil {
			panic("create meta[" + key + "] error: " + err.Error())
		}
	})
	return m
}

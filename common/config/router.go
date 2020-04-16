package config

import (
	"dubbo-gateway/common/constant"
	"dubbo-gateway/conf"
	"github.com/apache/dubbo-go/common/logger"
	perrors "github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

type RouterConfig struct {
	Config struct {
		Port   int    `yaml:"port"`
		Prefix string `yaml:"prefix"`
	} `yaml:"router"`
}

var routerConfig *RouterConfig

func init() {
	configStr, err := conf.GetConfig(constant.ConfGatewayFilePath, constant.DefaultGatewayFilePath)
	routerConfig = new(RouterConfig)
	err = yaml.Unmarshal([]byte(configStr), routerConfig)
	if err != nil {
		logger.Errorf("yaml.Unmarshal() = error:%v", perrors.WithStack(err))
	}
}

func GetRouterConfig() *RouterConfig {
	return routerConfig
}

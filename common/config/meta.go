package config

import (
	"dubbo-gateway/common/constant"
	"dubbo-gateway/conf"
	"github.com/apache/dubbo-go/common/logger"
	perrors "github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

type MetaConfig struct {
	Config map[string]interface{} `yaml:"meta"`
}

var metaConfig *MetaConfig

func init() {
	metaConfig = new(MetaConfig)
	configStr, err := conf.GetConfig(constant.ConfGatewayFilePath, constant.DefaultGatewayFilePath)
	deployConfig = new(Deploy)
	err = yaml.Unmarshal([]byte(configStr), deployConfig)
	if err != nil {
		logger.Errorf("yaml.Unmarshal() = error:%v", perrors.WithStack(err))
	}
}

func GetMetaConfig() MetaConfig {
	return *metaConfig
}

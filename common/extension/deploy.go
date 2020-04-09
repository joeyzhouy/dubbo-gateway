package extension

import (
	"dubbo-gateway/common/constant"
	"dubbo-gateway/conf"
	"github.com/apache/dubbo-go/common/logger"
	perrors "github.com/pkg/errors"
	"gopkg.in/yaml.v2"
	"time"
)

type Deploy struct {
	Config struct {
		Model    string `yaml:"model"`
		Multiple struct {
			Coordination struct {
				Protocol string        `yaml:"protocol"`
				Timeout  time.Duration `yaml:"timeout"`
				Address  string        `yaml:"address"`
				UserName string        `yaml:"username"`
				Password string        `yaml:"password"`
			} `yaml:"coordination"`
		} `yaml:"multiple"`
	} `yaml:"deploy"`
}

var deployConfig *Deploy

func init() {
	configStr, err := conf.GetConfig(constant.CONF_GATEWAY_FILE_PATH, "meta/gateway.xml")
	deployConfig = new(Deploy)
	err = yaml.Unmarshal([]byte(configStr), deployConfig)
	if err != nil {
		logger.Errorf("yaml.Unmarshal() = error:%v", perrors.WithStack(err))
	}
}

func GetDeployConfig() *Deploy {
	return deployConfig
}

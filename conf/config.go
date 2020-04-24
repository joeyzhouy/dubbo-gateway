package conf

import (
	perrors "github.com/pkg/errors"
	"io/ioutil"
	"os"
	"path"
	"sync"
)

var confMap sync.Map

func GetConfig(envName, defaultPath string) (string, error) {
	if v, ok := confMap.Load(envName); ok {
		return v.(string), nil
	}
	filePath := os.Getenv(envName)
	if filePath == "" {
		filePath = defaultPath
	}
	if filePath == "" {
		return "", perrors.Errorf("invalid config path: empty")
	}
	if path.Ext(filePath) != ".yml" {
		return "", perrors.Errorf("application configure file name{%v} suffix must be .yml", filePath)
	}
	confFileStream, err := ioutil.ReadFile(filePath)
	if err != nil {
		return "", perrors.Errorf("ioUtil.ReadFile(file:%s) = error:%v", filePath, perrors.WithStack(err))
	}
	result, _ := confMap.LoadOrStore(envName, string(confFileStream))
	return result.(string), nil
}


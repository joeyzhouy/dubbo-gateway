package single

import (
	"dubbo-gateway/common/config"
	"dubbo-gateway/common/extension"
	"dubbo-gateway/router/cache"
	"github.com/apache/dubbo-go/common/logger"
)

const SingleMode = "single"

type singleMode struct {
}

func (s *singleMode) Close() {
	logger.Info("single mode close")
}

func (s *singleMode) Add(apiId int64) error {
	return cache.Add(apiId)
}

func (s *singleMode) Remove(apiId int64) error {
	return cache.Remove(apiId)
}

func (s *singleMode) Refresh() error {
	return cache.Refresh()
}

func (s *singleMode) Start() {
	logger.Info("start single mode")
}

func init() {
	extension.SetMode(SingleMode, newSingleMode)
}

func newSingleMode(deploy *config.Deploy) (extension.Mode, error) {
	return &singleMode{}, nil
}

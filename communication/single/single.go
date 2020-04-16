package single

import (
	"dubbo-gateway/common/extension"
	"dubbo-gateway/router/cache"
	"github.com/apache/dubbo-go/common/logger"
)

const SingleMode = "single"

type singleMode struct {
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

func (s *singleMode) Start() error {
	logger.Info("start single mode")
	return nil
}

func init() {
	extension.SetMode(SingleMode, newSingleMode)
}

func newSingleMode(deploy *extension.Deploy) (extension.Mode, error) {
	return &singleMode{}, nil
}

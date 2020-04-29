package single

import (
	"dubbo-gateway/common/config"
	"dubbo-gateway/common/extension"
	"github.com/apache/dubbo-go/common/logger"
)

const SingleMode = "single"

type singleMode struct {
}

func (s *singleMode) Notify(event extension.ModeEvent) {
	panic("implement me")
}

func (s *singleMode) Close() {
	logger.Info("single mode close")
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

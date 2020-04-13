package single

import (
	"dubbo-gateway/common/extension"
	"github.com/apache/dubbo-go/common/logger"
)

const SingleMode = "single"

type singleMode struct {
}

func (*singleMode) Start() error {
	logger.Info("start single mode")
	return nil
}

func init() {
	extension.SetMode(SingleMode, newSingleMode)
}

func newSingleMode(deploy *extension.Deploy) (extension.Mode, error) {
	return &singleMode{}, nil
}

package communication

import (
	"dubbo-gateway/common/extension"
	_ "dubbo-gateway/communication/multiple"
	_ "dubbo-gateway/communication/single"
	"github.com/apache/dubbo-go/common/logger"
	perrors "github.com/pkg/errors"
)

func init() {
	mode, err := extension.GetConfigMode()
	if err != nil {
		logger.Errorf("get config error: %v", perrors.WithStack(err))
		return
	}
	if err := mode.Start(); err != nil {
		logger.Errorf("model start error: %v", perrors.WithStack(err))
	}
}

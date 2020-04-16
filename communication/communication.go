package communication

import (
	"dubbo-gateway/common/extension"
	_ "dubbo-gateway/communication/multiple"
	_ "dubbo-gateway/communication/single"
	"github.com/apache/dubbo-go/common/logger"
)

func init() {
	c := new(communication)
	var err error
	c.mode, err = extension.GetConfigMode()
	if err != nil {
		logger.Errorf("get config mode error: %v", err)
		return
	}
	extension.SetOrigin(extension.Communication, c)
}

type communication struct {
	mode extension.Mode
}

func (c *communication) Start() {
	c.mode.Start()
}

func (c *communication) Close() {
	c.mode.Close()
}

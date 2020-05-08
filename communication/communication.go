package communication

import (
	"dubbo-gateway/common/extension"
	_ "dubbo-gateway/communication/multiple"
	_ "dubbo-gateway/communication/single"
)

func init() {
	c := new(communication)
	c.mode  = extension.GetConfigMode()
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

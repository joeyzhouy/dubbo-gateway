package logger

import (
	"dubbo-gateway/common/extension"
	dlogger "github.com/apache/dubbo-go/common/logger"
	"log"
)

func init() {
	extension.AddInit(&loggerInit{})
}

type loggerInit struct{}

func (*loggerInit) Init() {
	if dlogger.GetLogger() == nil {
		if err := dlogger.InitLog("logger/log.yaml"); err != nil {
			log.Printf("[InitLog] warn: %v", err)
		}
	}
}

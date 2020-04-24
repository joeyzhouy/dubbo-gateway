package main

import (
	"dubbo-gateway/common/extension"
	_ "dubbo-gateway/communication"
	_ "dubbo-gateway/meta/kv/zookeeper"
	_ "dubbo-gateway/meta/relation/mysql"
	//_ "dubbo-gateway/router"
	_ "dubbo-gateway/web"
	_ "dubbo-gateway/web/console"
	"github.com/apache/dubbo-go/common/logger"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	//gin.SetMode(gin.ReleaseMode)
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, os.Kill, syscall.SIGHUP,
		syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	extension.Start()
	sig := <-signals
	logger.Infof("get signal %s", sig.String())
	extension.Close()
	time.AfterFunc(time.Duration(5*time.Second), func() {
		logger.Warnf("app exit now by force...")
		os.Exit(1)
	})
}

package main

import (
	"dubbo-gateway/common/extension"
	_ "dubbo-gateway/communication"
	_ "dubbo-gateway/registry/zookeeper"
	_ "dubbo-gateway/router"
	_ "dubbo-gateway/web"
	_ "dubbo-gateway/web/console"
	_ "github.com/apache/dubbo-go/cluster/cluster_impl"
	_ "github.com/apache/dubbo-go/cluster/loadbalance"
	"github.com/apache/dubbo-go/common/logger"
	_ "github.com/apache/dubbo-go/filter/filter_impl"
	_ "github.com/apache/dubbo-go/registry/protocol"
	_ "github.com/apache/dubbo-go/registry/zookeeper"
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

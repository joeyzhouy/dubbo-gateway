package router

import (
	"dubbo-gateway/common/config"
	"dubbo-gateway/common/extension"
	"dubbo-gateway/common/utils"
	"dubbo-gateway/router/cache"
	"fmt"
	"github.com/gin-gonic/gin"
)

var r *gin.Engine
var routerConfig *config.RouterConfig

func init() {
	router := new(routerOrigin)
	router.routerConfig = config.GetRouterConfig()
	router.r = gin.New()
	router.r.Use(utils.LoggerWithWriter(), gin.Recovery())
	router.r.Any(routerConfig.Config.Prefix, cache.Operate)
	extension.SetOrigin(extension.Router, router)
}

type routerOrigin struct {
	r            *gin.Engine
	routerConfig *config.RouterConfig
}

func (r *routerOrigin) Start() {
	go r.r.Run(fmt.Sprintf(":%d", routerConfig.Config.Port))
}

func (r *routerOrigin) Close() {
	cache.Close()
}

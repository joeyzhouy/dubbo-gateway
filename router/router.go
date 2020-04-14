package router

import (
	"dubbo-gateway/common/extension"
	"dubbo-gateway/router/cache"
	"fmt"
	"github.com/gin-gonic/gin"
)

var r *gin.Engine
//var authGroup *gin.RouterGroup
var routerConfig *extension.RouterConfig

func init() {
	r = gin.New()
	r.Use(extension.LoggerWithWriter(), gin.Recovery())
	routerConfig = extension.GetRouterConfig()
	r.Any(routerConfig.Config.Prefix, cache.Operate)
}

func Run() error {
	return r.Run(fmt.Sprintf(":%d", routerConfig.Config.Port))
}

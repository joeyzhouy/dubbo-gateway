package router

import (
	"dubbo-gateway/common/extension"
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
	r.GET(routerConfig.Config.Prefix, operate)
}

func Run() error {
	return r.Run(fmt.Sprintf(":%d", routerConfig.Config.Port))
}

func operate(ctx *gin.Context) {

}

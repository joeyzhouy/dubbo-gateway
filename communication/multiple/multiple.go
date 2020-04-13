package multiple

import (
	"dubbo-gateway/common/extension"
	"dubbo-gateway/registry"
	"fmt"
	"github.com/gin-gonic/gin"
)

const MultipleMode = "multiple"

type multipleMode struct {
	r         *gin.Engine
	authGroup *gin.RouterGroup
	ipList    []string
	reg       *registry.Registry
	port      int
}

func (m *multipleMode) Start() error {
	return m.r.Run(fmt.Sprintf(":%d", m.port))
}

func init() {
	extension.SetMode(MultipleMode, newMultipleMode)
}

func newMultipleMode(deploy *extension.Deploy) (extension.Mode, error) {
	mode := new(multipleMode)
	mode.r = gin.New()
	mode.r.Use(extension.LoggerWithWriter(), gin.Recovery())
	mode.authGroup = mode.r.Group("/", auth(mode))
	mConfig := deploy.Config.Multiple
	mode.port = mConfig.Port
	return nil, nil
}

func auth(mode *multipleMode) gin.HandlerFunc {
	result := make(map[string]interface{})
	result["code"] = 403
	return func(ctx *gin.Context) {
		if len(mode.ipList) == 0 {
			ctx.AbortWithStatusJSON(200, &result)
			return
		}
		remoteAddress := ctx.Request.RemoteAddr
		for _, ip := range mode.ipList {
			if ip == remoteAddress {
				ctx.Next()
				return
			}
		}
		ctx.AbortWithStatusJSON(200, &result)
	}
}

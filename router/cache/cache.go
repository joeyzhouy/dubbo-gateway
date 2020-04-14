package cache

import (
	"dubbo-gateway/common/utils"
	"dubbo-gateway/service"
	"dubbo-gateway/service/vo"
	"github.com/apache/dubbo-go/common"
	//_ "github.com/apache/dubbo-go/common/proxy/proxy_factory"
	"github.com/apache/dubbo-go/config"
	"github.com/gin-gonic/gin"
	perrors "github.com/pkg/errors"
	"strings"
	"sync"
)

const APPLICATION_JSON = "application/json"

var rCache *routerCache
var requestParamOperate map[string]func(ctx *gin.Context) (header, params map[string]interface{}, err error)

func init() {
	rCache = new(routerCache)
	requestParamOperate[utils.GET] = func(ctx *gin.Context) (header, params map[string]interface{}, err error) {

	}
	requestParamOperate[utils.POST] = func(ctx *gin.Context) (header, params map[string]interface{}, err error) {

	}
}

type routerCache struct {
	sync.RWMutex
	uri  map[string]*apiConfig
	rMap map[int64]*config.ReferenceConfig
	rUri map[int64][]string
	service.RouterService
	service.
}

type apiConfigCache struct {
	vo.ApiConfigInfo

}

func (r *routerCache) get(key string) {

}

func (r *routerCache) refresh() {

}

func (r *routerCache) getApiConfig() error {
	apiConfigs, err := r.ListAll()
	if err != nil {
		return err
	}

}

func Operate(ctx *gin.Context) {
	if APPLICATION_JSON != ctx.ContentType() {
		ctx.AbortWithStatusJSON(200,
			&utils.Response{Code: utils.Fail, Message: "just support: application/json"})
		return
	}
	f, ok := requestParamOperate[strings.ToLower(ctx.Request.Method)]
	if !ok {
		ctx.AbortWithStatusJSON(200,
			&utils.Response{Code: utils.Fail, Message: "not support request method: " + ctx.Request.Method})
		return
	}

	header, params, err := f(ctx)
	if err != nil {
		ctx.AbortWithStatusJSON(200,
			&utils.Response{Code: utils.Fail, Message: "system error"})
		return
	}
}

func Refresh() error {

	return nil
}

func Get(key string) (common.RPCService, error) {
	rCache.RLock()
	defer rCache.Unlock()

	apiConfig, ok := rCache.uri[key]
	if !ok {
		return nil, perrors.Errorf("not found service with uri: %s", key)
	}
	rConfig, ok := rCache.rMap[apiConfig.config.ReferenceId]

}

func Remove(key string) {

}



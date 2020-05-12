package router

import (
	"dubbo-gateway/common/config"
	"dubbo-gateway/common/constant"
	"dubbo-gateway/common/extension"
	"dubbo-gateway/common/utils"
	"fmt"
	"github.com/gin-gonic/gin"
	jsoniter "github.com/json-iterator/go"
	"strings"
)

var justSupportPostResponse = utils.Response{Code: 400, Message: "just support POST"}
var missMethodParamResponse = utils.Response{Code: 400, Message: "miss method key"}
var systemErrorReponse = utils.Response{Code: 500, Message: "system error"}
var json = jsoniter.ConfigCompatibleWithStandardLibrary

func init() {
	router := new(routerOrigin)
	router.routerConfig = config.GetRouterConfig()
	router.r = gin.New()
	router.r.Use(utils.LoggerWithWriter(), gin.Recovery())
	router.r.Any(router.routerConfig.Config.Prefix, router.operate)
	router.mode = extension.GetConfigMode()
	extension.SetOrigin(extension.Router, router)
}

type routerOrigin struct {
	r            *gin.Engine
	routerConfig *config.RouterConfig
	mode         extension.Mode
}

func (r *routerOrigin) Start() {
	go r.r.Run(fmt.Sprintf(":%d", r.routerConfig.Config.Port))
}

func (r *routerOrigin) Close() {
	//cache.Close()
}

func (r *routerOrigin) operate(ctx *gin.Context) {
	if utils.POST != strings.ToLower(ctx.Request.Method) {
		ctx.AbortWithStatusJSON(200, &justSupportPostResponse)
		return
	}
	rs := new(RequestStructure)
	if utils.IsErrorEmpty(ctx.ShouldBindJSON(rs), ctx) {
		if rs.Method == "" {
			ctx.AbortWithStatusJSON(200, &missMethodParamResponse)
			return
		}
		paramMap, err := r.getContextParam(ctx)
		if err != nil {
			ctx.AbortWithStatusJSON(200, &systemErrorReponse)
			return
		}
		paramMap[constant.RouterBodyKey] = rs.Content
		result, err := r.mode.Invoke(rs.Method, paramMap)
		if err != nil {
			utils.OperateResponse(nil, err, ctx)
			return
		}
		bs, err := json.Marshal(result)
		_, _ = ctx.Writer.Write(bs)
	}

}

func (r *routerOrigin) getContextParam(ctx *gin.Context) (map[string]interface{}, error) {
	paramMap := make(map[string]interface{})
	paramMap[constant.RouterHeaderKey] = ctx.Request.Header
	paramMap[constant.RouterBodyKey] = ctx.Request.Header
	paramMap[constant.RouterQueryKey] = ctx.Request.URL.Query()
	paramMap[constant.CustomKey] = make(map[string]interface{})
	return paramMap, nil
}

type RequestStructure struct {
	Method  string      `json:"method"`
	Content interface{} `json:"content"`
}

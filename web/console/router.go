package console

import (
	"dubbo-gateway/common/utils"
	"dubbo-gateway/service"
	"dubbo-gateway/service/entry"
	"dubbo-gateway/web"
	"github.com/gin-gonic/gin"
	"strconv"
)

func init() {
	r := &routerController{service.NewRouterService()}
	rGroup := web.AuthGroup().Group("/r")
	rGroup.POST("/", r.CreateRouter)
	rGroup.GET("/list", r.ListByUser)
	rGroup.DELETE("/", r.DeleteRouter)
}

type routerController struct {
	service.RouterService
}

func (r *routerController) CreateRouter(ctx *gin.Context) {
	api := new(entry.ApiConfig)
	if utils.IsErrorEmpty(ctx.ShouldBindJSON(api), ctx) {
		if api.Uri == "" || api.MethodId == 0 {
			utils.ParamMissResponseOperation(ctx)
			return
		}
		if user, err := web.GetSessionUser(ctx); utils.IsErrorEmpty(err, ctx) {
			api.UserId = user.ID
			utils.OperateResponse(nil, r.RouterService.AddRouter(api), ctx)
		}
	}
}

func (r *routerController) DeleteRouter(ctx *gin.Context) {
	if idStr, ok := ctx.GetQuery("id"); ok {
		if id, err := strconv.ParseInt(idStr, 10, 64); utils.IsErrorEmpty(err, ctx) {
			utils.OperateResponse(nil, r.RouterService.DeleteRouter(id), ctx)
		}
	} else {
		utils.ParamMissResponseOperation(ctx)
	}
}

func (r *routerController) ListByUser(ctx *gin.Context) {
	if user, err := web.GetSessionUser(ctx); utils.IsErrorEmpty(err, ctx) {
		result, err := r.RouterService.ListRouterByUserId(user.ID)
		utils.OperateResponse(result, err, ctx)
	}
}

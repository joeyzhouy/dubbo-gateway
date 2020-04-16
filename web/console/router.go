package console

import (
	"dubbo-gateway/common/extension"
	"dubbo-gateway/common/utils"
	"dubbo-gateway/service"
	"dubbo-gateway/service/entry"
	"dubbo-gateway/web"
	"github.com/apache/dubbo-go/common/logger"
	"github.com/gin-gonic/gin"
	perrors "github.com/pkg/errors"
	"strconv"
)

func init() {
	mode, err := extension.GetConfigMode()
	if err != nil {
		logger.Errorf("get config mode error: %v", perrors.WithStack(err))
		return
	}
	r := &routerController{service.NewRouterService(), mode}
	rGroup := web.AuthGroup().Group("/r")
	rGroup.POST("/", r.CreateRouter)
	rGroup.GET("/list", r.ListByUser)
	rGroup.DELETE("/", r.DeleteRouter)
}

type routerController struct {
	service.RouterService
	extension.Mode
}

func (r *routerController) CreateRouter(ctx *gin.Context) {
	api := new(entry.ApiConfig)
	if utils.IsErrorEmpty(ctx.ShouldBindJSON(api), ctx) {
		if api.Uri == "" {
			utils.ParamMissResponseOperation(ctx)
			return
		}
		if user, err := web.GetSessionUser(ctx); utils.IsErrorEmpty(err, ctx) {
			api.UserId = user.ID
			if utils.IsErrorEmpty(r.RouterService.AddRouter(api), ctx) {
				go r.Mode.Add(api.ID)
			}
			//utils.OperateResponse(nil, r.RouterService.AddRouter(api), ctx)
		}
	}
}

func (r *routerController) DeleteRouter(ctx *gin.Context) {
	if idStr, ok := ctx.GetQuery("id"); ok {
		if id, err := strconv.ParseInt(idStr, 10, 64); utils.IsErrorEmpty(err, ctx) {
			if utils.IsErrorEmpty(r.RouterService.DeleteRouter(id), ctx) {
				go r.Mode.Remove(id)
			}
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

package console

import (
	"dubbo-gateway/common/extension"
	"dubbo-gateway/common/utils"
	"dubbo-gateway/service"
	"dubbo-gateway/service/vo"
	"dubbo-gateway/web"
	"github.com/gin-gonic/gin"
	"strconv"
)

func init() {
	r := &routerController{extension.GetMeta().NewRouterService(), extension.GetConfigMode()}
	rGroup := web.AuthGroup().Group("/ro")
	rGroup.POST("/", r.CreateRouter)
	rGroup.GET("/list", r.ListAllAvailable)
	rGroup.DELETE("/", r.DeleteRouter)
	rGroup.GET("/", r.)

	fGroup := web.AuthGroup().Group("/f")
	fGroup.POST("/", r.CreateFilter)
	fGroup.PUT("/", )
	fGroup.DELETE("/")
	fGroup.GET("/list")
	fGroup.GET("/", r)
}

type routerController struct {
	service.RouterService
	extension.Mode
}

func (r *routerController) FilterDetail(ctx *gin.Context) {
	if idStr, ok := ctx.GetQuery("id"); ok {
		if id, err := strconv.ParseInt(idStr, 10, 64); utils.IsErrorEmpty(err, ctx) {
			if utils.IsErrorEmpty(r.RouterService.DeleterFilter(id), ctx) {
				utils.OperateSuccessResponse(nil, ctx)
			}
		}
	} else {
		utils.ParamMissResponseOperation(ctx)
	}
}

func (r *routerController) CreateFilter(ctx *gin.Context) {
	filter := new(vo.ApiFilterInfo)
	if utils.IsErrorEmpty(ctx.ShouldBindJSON(filter), ctx) {
		if filter.Filter.MethodId == 0 || filter.Filter.ReferenceId == 0 ||
			filter.Filter.Name == "" {
			utils.ParamMissResponseOperation(ctx)
			return
		}
		utils.OperateResponse(nil, r.RouterService.AddFilter(filter), ctx)
	}
}

func (r *routerController) ModifyFilter(ctx *gin.Context) {
	filter := new(vo.ApiFilterInfo)
	if utils.IsErrorEmpty(ctx.ShouldBindJSON(filter), ctx) {
		if filter.Filter.MethodId == 0 || filter.Filter.ReferenceId == 0 ||
			filter.Filter.ID == 0 || filter.Filter.Name == "" {
			utils.ParamMissResponseOperation(ctx)
			return
		}
		utils.OperateResponse(nil, r.RouterService.ModifyFilter(filter), ctx)
	}
}

func (r *routerController) DeleteFilter(ctx *gin.Context) {
	if idStr, ok := ctx.GetQuery("id"); ok {
		if id, err := strconv.ParseInt(idStr, 10, 64); utils.IsErrorEmpty(err, ctx) {
			if utils.IsErrorEmpty(r.RouterService.DeleterFilter(id), ctx) {
				utils.OperateSuccessResponse(nil, ctx)
			}
		}
	} else {
		utils.ParamMissResponseOperation(ctx)
	}
}

func (r *routerController) ListFilters(ctx *gin.Context) {
	result, err := r.RouterService.ListFilters()
	utils.OperateResponse(result, err, ctx)
}

func (r *routerController) ListAllAvailable(ctx *gin.Context) {
	if methodNameLike, ok := ctx.GetQuery("methodName"); ok {
		result, err := r.RouterService.SearchByMethodName(methodNameLike)
		utils.OperateResponse(result, err, ctx)
		return
	}
	utils.ParamMissResponseOperation(ctx)
}

func (r *routerController) CreateRouter(ctx *gin.Context) {
	api := new(vo.ApiConfigInfo)
	if utils.IsErrorEmpty(ctx.ShouldBindJSON(api), ctx) {
		if api.ApiConfig.Method == "" {
			utils.ParamMissResponseOperation(ctx)
			return
		}
		if user, err := web.GetSessionUser(ctx); utils.IsErrorEmpty(err, ctx) {
			api.ApiConfig.UserId = user.ID
			if utils.IsErrorEmpty(r.RouterService.AddConfig(api), ctx) {
				utils.OperateSuccessResponse(nil, ctx)
			}
		}
	}
}

func (r *routerController) DeleteRouter(ctx *gin.Context) {
	if idStr, ok := ctx.GetQuery("id"); ok {
		if id, err := strconv.ParseInt(idStr, 10, 64); utils.IsErrorEmpty(err, ctx) {
			if utils.IsErrorEmpty(r.RouterService.DeleteConfig(id), ctx) {
				utils.OperateSuccessResponse(nil, ctx)
			}
		}
	} else {
		utils.ParamMissResponseOperation(ctx)
	}
}

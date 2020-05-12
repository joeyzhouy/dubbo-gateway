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
	r := &routerController{extension.GetMeta().NewRouterService()}
	rGroup := web.AuthGroup().Group("/ro")
	rGroup.POST("/", r.CreateRouter)
	rGroup.GET("/list", r.ListAllAvailable)
	rGroup.DELETE("/", r.DeleteRouter)
	rGroup.GET("/", r.GetApiDetail)

	fGroup := web.AuthGroup().Group("/f")
	fGroup.POST("/", r.CreateFilter)
	fGroup.PUT("/", r.ModifyFilter)
	fGroup.DELETE("/", r.DeleteFilter)
	fGroup.GET("/list", r.ListFilters)
	fGroup.GET("/", r.FilterDetail)
}

type routerController struct {
	service.RouterService
}

func (r *routerController) GetApiDetail(ctx *gin.Context) {
	if idStr, ok := ctx.GetQuery("id"); ok {
		if id, err := strconv.ParseInt(idStr, 10, 64); utils.IsErrorEmpty(err, ctx) {
			result, err := r.RouterService.GetByConfigId(id)
			utils.OperateResponse(result, err, ctx)
		}
	} else {
		utils.ParamMissResponseOperation(ctx)
	}
}

func (r *routerController) FilterDetail(ctx *gin.Context) {
	if idStr, ok := ctx.GetQuery("id"); ok {
		if id, err := strconv.ParseInt(idStr, 10, 64); utils.IsErrorEmpty(err, ctx) {
			result, err := r.RouterService.GetFilter(id)
			utils.OperateResponse(result, err, ctx)
		}
	} else {
		utils.ParamMissResponseOperation(ctx)
	}
}

func (r *routerController) CreateFilter(ctx *gin.Context) {
	filter := new(vo.ApiFilterInfo)
	if utils.IsErrorEmpty(ctx.ShouldBindJSON(filter), ctx) {
		if filter.ApiFilter.MethodId == 0 || filter.ApiFilter.ReferenceId == 0 ||
			filter.ApiFilter.Name == "" {
			utils.ParamMissResponseOperation(ctx)
			return
		}
		utils.OperateResponse(nil, r.RouterService.AddFilter(filter), ctx)
	}
}

func (r *routerController) ModifyFilter(ctx *gin.Context) {
	filter := new(vo.ApiFilterInfo)
	if utils.IsErrorEmpty(ctx.ShouldBindJSON(filter), ctx) {
		if filter.ApiFilter.MethodId == 0 || filter.ApiFilter.ReferenceId == 0 ||
			filter.ApiFilter.ID == 0 || filter.ApiFilter.Name == "" {
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

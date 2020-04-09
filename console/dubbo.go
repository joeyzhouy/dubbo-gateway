package console

import (
	"dubbo-gateway/service"
	"dubbo-gateway/service/entry"
	"dubbo-gateway/service/vo"
	"dubbo-gateway/web"
	"github.com/gin-gonic/gin"
	"strconv"
)

func init() {
	d := &dubboController{service.NewRegistryService(),
		service.NewMethodService()}
	rGroup := web.AuthGroup().Group("/reg")
	rGroup.GET("/detail", d.GetRegisterDetail)
	rGroup.GET("/list", d.ListByUser)
	rGroup.POST("/", d.CreateRegister)
	rGroup.DELETE("/", d.DeleteRegister)

	mGroup := web.AuthGroup().Group("/m")
	mGroup.POST("/", d.AddMethod)
	mGroup.GET("/", d.GetMethodDetail)
	mGroup.DELETE("/", d.DeleteMethod)
	mGroup.GET("/r", d.GetMethodsByReference)
	mGroup.GET("/u", d.ListByUserIdAndMethodName)
}

type dubboController struct {
	service.RegisterService
	service.MethodService
}

func (d *dubboController) GetRegisterDetail(ctx *gin.Context) {
	if idStr, ok := ctx.GetQuery("id"); ok {
		if id, err := strconv.ParseInt(idStr, 10, 64); isErrorEmpty(err, ctx) {
			if user, err := web.GetSessionUser(ctx); isErrorEmpty(err, ctx) {
				result, err := d.RegisterService.RegisterDetail(user.ID, id)
				operateResponse(result, err, ctx)
			}
		}
	} else {
		ParamMissResponseOperation(ctx)
	}
}

func (d *dubboController) ListByUser(ctx *gin.Context) {
	if user, err := web.GetSessionUser(ctx); isErrorEmpty(err, ctx) {
		result, err := d.RegisterService.ListRegistryByUser(user.ID)
		operateResponse(result, err, ctx)
	}
}

func (d *dubboController) CreateRegister(ctx *gin.Context) {
	reg := new(entry.Registry)
	if isErrorEmpty(ctx.ShouldBindJSON(reg), ctx) {
		if reg.Name == "" || reg.Address == "" || reg.Protocol == "" {
			ParamMissResponseOperation(ctx)
			return
		}
		if user, err := web.GetSessionUser(ctx); isErrorEmpty(err, ctx) {
			reg.UserId = user.ID
			operateResponse(nil, d.AddRegistryConfig(*reg), ctx)
		}
	}
}

func (d *dubboController) DeleteRegister(ctx *gin.Context) {
	if idStr, ok := ctx.GetQuery("id"); ok {
		if id, err := strconv.ParseInt(idStr, 10, 64); isErrorEmpty(err, ctx) {
			if user, err := web.GetSessionUser(ctx); isErrorEmpty(err, ctx) {
				operateResponse(nil, d.RegisterService.DeleteRegistryConfig(id, user.ID), ctx)
			}
		}
	} else {
		ParamMissResponseOperation(ctx)
	}
}

func (d *dubboController) AddMethod(ctx *gin.Context) {
	method := new(vo.Method)
	if isErrorEmpty(ctx.ShouldBindJSON(method), ctx) {
		if method.MethodName == "" || method.ReferenceId == 0 ||
			len(method.Params) == 0 {
			ParamMissResponseOperation(ctx)
			return
		}
		operateResponse(nil, d.MethodService.AddMethod(method), ctx)
	}
}

func (d *dubboController) GetMethodDetail(ctx *gin.Context) {
	if idStr, ok := ctx.GetQuery("id"); ok {
		if id, err := strconv.ParseInt(idStr, 10, 64); isErrorEmpty(err, ctx) {
			result, err := d.MethodService.GetMethodDetail(id)
			operateResponse(result, err, ctx)
		}
	} else {
		ParamMissResponseOperation(ctx)
	}
}

func (d *dubboController) DeleteMethod(ctx *gin.Context) {
	if idStr, ok := ctx.GetQuery("id"); ok {
		if id, err := strconv.ParseInt(idStr, 10, 64); isErrorEmpty(err, ctx) {
			operateResponse(nil, d.MethodService.DeleteMethod(id), ctx)
		}
	} else {
		ParamMissResponseOperation(ctx)
	}
}

func (d *dubboController) GetMethodsByReference(ctx *gin.Context) {
	if idStr, ok := ctx.GetQuery("id"); ok {
		if id, err := strconv.ParseInt(idStr, 10, 64); isErrorEmpty(err, ctx) {
			result, err := d.MethodService.GetMethodsByReferenceId(id);
			operateResponse(result, err, ctx)
		}
	} else {
		ParamMissResponseOperation(ctx)
	}
}

func (d *dubboController) ListByUserIdAndMethodName(ctx *gin.Context) {
	if methodName, ok := ctx.GetQuery("methodName"); ok {
		if user, err := web.GetSessionUser(ctx); isErrorEmpty(err, ctx) {
			result, err := d.MethodService.ListByUserIdAndMethodName(user.ID, methodName)
			operateResponse(result, err, ctx)
		}
	} else {
		ParamMissResponseOperation(ctx)
	}
}

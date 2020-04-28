package console

import (
	"dubbo-gateway/common/constant"
	"dubbo-gateway/common/extension"
	"dubbo-gateway/common/utils"
	"dubbo-gateway/service"
	"dubbo-gateway/service/entry"
	"dubbo-gateway/service/vo"
	"dubbo-gateway/web"
	"github.com/gin-gonic/gin"
	"strconv"
)

func init() {
	metaDate := extension.GetMeta()
	d := &dubboController{metaDate.NewRegisterService(),
		metaDate.NewMethodService(),
		metaDate.NewReferenceService(),
		metaDate.NewEntryService()}
	rGroup := web.AuthGroup().Group("/reg")
	rGroup.GET("/protocol", d.Protocol)
	rGroup.GET("/detail", d.GetRegisterDetail)
	rGroup.GET("/dv", d.GetRegistryDetailWithReference)
	rGroup.GET("/list", d.ListByUser)
	rGroup.POST("/", d.CreateRegister)
	rGroup.DELETE("/", d.DeleteRegister)
	rGroup.POST("/search", d.SearchRegistry)

	mGroup := web.AuthGroup().Group("/m")
	mGroup.POST("/", d.AddMethod)
	mGroup.GET("/", d.GetMethodDetail)
	mGroup.DELETE("/", d.DeleteMethod)
	mGroup.GET("/r", d.GetMethodsByReference)
	mGroup.POST("/search", d.SearchMethods)
	//mGroup.GET("/u", d.ListByUserIdAndMethodName)

	reGroup := web.AuthGroup().Group("/r")
	reGroup.GET("/c", d.GetClusters)
	reGroup.GET("/protocol", d.ReferenceProtocol)
	reGroup.GET("/list", d.ListReference)
	reGroup.GET("/detail", d.ReferenceDetail)
	reGroup.GET("/u", d.ListReferenceByUser)
	reGroup.POST("/", d.CreateReference)
	reGroup.DELETE("/", d.DeleteReference)
	reGroup.POST("/search", d.SearchReference)

	eGroup := web.AuthGroup().Group("/e")
	eGroup.POST("/search", d.SearchEntries)
	eGroup.POST("/", d.AddEntry)
	eGroup.GET("/list", d.GetAllEntries)

}

type dubboController struct {
	service.RegisterService
	service.MethodService
	service.ReferenceService
	service.EntryService
}

type SearchEntryParam struct {
	Name     string `json:"name"`
	PageSize int    `json:"pageSize"`
}

type SearchMethodParam struct {
	Name        string `json:"methodName"`
	RegistryId  int64  `json:"registryId"`
	ReferenceId int64  `json:"referenceId"`
}

func (d *dubboController) SearchMethods(ctx *gin.Context) {
	param := new(SearchMethodParam)
	if utils.IsErrorEmpty(ctx.ShouldBindJSON(param), ctx) {
		result, err := d.MethodService.SearchMethods(param.RegistryId, param.RegistryId, param.Name)
		utils.OperateResponse(result, err, ctx)
	}
}

func (d *dubboController) GetAllEntries(ctx *gin.Context) {
	result, err := d.EntryService.ListAll()
	utils.OperateResponse(result, err, ctx)
}

func (d *dubboController) AddEntry(ctx *gin.Context) {
	e := new(entry.EntryStructure)
	if utils.IsErrorEmpty(ctx.ShouldBindJSON(e), ctx) {
		if e.Name == "" || e.Key == "" {
			utils.ParamMissResponseOperation(ctx)
			return
		}
		utils.OperateResponse(nil, d.EntryService.SaveEntry(e), ctx)
	}
}

func (d *dubboController) SearchEntries(ctx *gin.Context) {
	param := new(SearchEntryParam)
	if utils.IsErrorEmpty(ctx.ShouldBindJSON(param), ctx) {
		size := param.PageSize
		if size == 0 {
			size = 10
		}
		result, err := d.EntryService.SearchEntries(param.Name, size)
		utils.OperateResponse(result, err, ctx)
	}
}

type SearchRegistryParam struct {
	Name string `json:"name"`
}

func (d *dubboController) GetClusters(ctx *gin.Context) {
	utils.OperateSuccessResponse([]string{
		"failover", "failback", "failfast", "forking",
	}, ctx)
}

func (d *dubboController) GetRegistryDetailWithReference(ctx *gin.Context) {
	if idStr, ok := ctx.GetQuery("id"); ok {
		if id, err := strconv.ParseInt(idStr, 10, 64); utils.IsErrorEmpty(err, ctx) {
			result, err := d.RegisterService.GetByRegistryId(id)
			utils.OperateResponse(result, err, ctx)
		}
	} else {
		utils.ParamMissResponseOperation(ctx)
	}
}

func (d *dubboController) SearchRegistry(ctx *gin.Context) {
	param := new(SearchRegistryParam)
	if utils.IsErrorEmpty(ctx.ShouldBindJSON(param), ctx) {
		result, err := d.RegisterService.GetRegistryByName(param.Name)
		utils.OperateResponse(result, err, ctx)
	}
}

type SearchReferenceParam struct {
	RegistryId int64  `json:"registryId"`
	Name       string `json:"name"`
}

func (d *dubboController) SearchReference(ctx *gin.Context) {
	param := new(SearchReferenceParam)
	if utils.IsErrorEmpty(ctx.ShouldBindJSON(param), ctx) {
		result, err := d.ReferenceService.GetByRegistryIdAndName(param.RegistryId, param.Name)
		utils.OperateResponse(result, err, ctx)
	}
}

func (d *dubboController) ReferenceProtocol(ctx *gin.Context) {
	utils.OperateSuccessResponse([]string{constant.ProtocolDubbo}, ctx)
}

func (d *dubboController) ListReference(ctx *gin.Context) {
	result, err := d.ReferenceService.ListAll()
	utils.OperateResponse(result, err, ctx)
}

func (d *dubboController) ReferenceDetail(ctx *gin.Context) {
	if idStr, ok := ctx.GetQuery("id"); ok {
		if id, err := strconv.ParseInt(idStr, 10, 64); utils.IsErrorEmpty(err, ctx) {
			result, err := d.ReferenceService.GetReferenceById(id)
			utils.OperateResponse(result, err, ctx)
		}
	} else {
		utils.ParamMissResponseOperation(ctx)
	}
}

func (d *dubboController) ListReferenceByUser(ctx *gin.Context) {
	if user, err := web.GetSessionUser(ctx); utils.IsErrorEmpty(err, ctx) {
		result, err := d.ReferenceService.ListByUser(user.ID)
		utils.OperateResponse(result, err, ctx)
	}
}

func (d *dubboController) CreateReference(ctx *gin.Context) {
	reference := new(entry.Reference)
	if err := ctx.ShouldBindJSON(reference); utils.IsErrorEmpty(err, ctx) {
		if reference.RegistryId == 0 || reference.Protocol == "" ||
			reference.InterfaceName == "" || reference.Cluster == "" {
			utils.ParamMissResponseOperation(ctx)
		} else {
			utils.OperateResponse(nil, d.ReferenceService.AddReference(*reference), ctx)
		}
	}
}

func (d *dubboController) DeleteReference(ctx *gin.Context) {
	if idStr, ok := ctx.GetQuery("id"); ok {
		if id, err := strconv.ParseInt(idStr, 10, 64); utils.IsErrorEmpty(err, ctx) {
			utils.OperateResponse(nil, d.ReferenceService.DeleteReference(id), ctx)
		}
	} else {
		utils.ParamMissResponseOperation(ctx)
	}

}

func (d *dubboController) Protocol(ctx *gin.Context) {
	protocols := []string{constant.ProtocolZookeeper}
	utils.OperateSuccessResponse(protocols, ctx)
}

func (d *dubboController) GetRegisterDetail(ctx *gin.Context) {
	if idStr, ok := ctx.GetQuery("id"); ok {
		if id, err := strconv.ParseInt(idStr, 10, 64); utils.IsErrorEmpty(err, ctx) {
			if user, err := web.GetSessionUser(ctx); utils.IsErrorEmpty(err, ctx) {
				result, err := d.RegisterService.RegisterDetail(user.ID, id)
				utils.OperateResponse(result, err, ctx)
			}
		}
	} else {
		utils.ParamMissResponseOperation(ctx)
	}
}

func (d *dubboController) ListByUser(ctx *gin.Context) {
	if user, err := web.GetSessionUser(ctx); utils.IsErrorEmpty(err, ctx) {
		result, err := d.RegisterService.ListRegistryByUser(user.ID)
		utils.OperateResponse(result, err, ctx)
	}
}

func (d *dubboController) CreateRegister(ctx *gin.Context) {
	reg := new(entry.Registry)
	if utils.IsErrorEmpty(ctx.ShouldBindJSON(reg), ctx) {
		if reg.Name == "" || reg.Address == "" || reg.Protocol == "" {
			utils.ParamMissResponseOperation(ctx)
			return
		}
		if user, err := web.GetSessionUser(ctx); utils.IsErrorEmpty(err, ctx) {
			reg.UserId = user.ID
			utils.OperateResponse(nil, d.AddRegistryConfig(*reg), ctx)
		}
	}
}

func (d *dubboController) DeleteRegister(ctx *gin.Context) {
	if idStr, ok := ctx.GetQuery("id"); ok {
		if id, err := strconv.ParseInt(idStr, 10, 64); utils.IsErrorEmpty(err, ctx) {
			if user, err := web.GetSessionUser(ctx); utils.IsErrorEmpty(err, ctx) {
				utils.OperateResponse(nil, d.RegisterService.DeleteRegistryConfig(id, user.ID), ctx)
			}
		}
	} else {
		utils.ParamMissResponseOperation(ctx)
	}
}

func (d *dubboController) AddMethod(ctx *gin.Context) {
	method := new(vo.Method)
	if utils.IsErrorEmpty(ctx.ShouldBindJSON(method), ctx) {
		if method.MethodName == "" || method.ReferenceId == 0 ||
			len(method.Params) == 0 {
			utils.ParamMissResponseOperation(ctx)
			return
		}
		utils.OperateResponse(nil, d.MethodService.AddMethod(method), ctx)
	}
}

func (d *dubboController) GetMethodDetail(ctx *gin.Context) {
	if idStr, ok := ctx.GetQuery("id"); ok {
		if id, err := strconv.ParseInt(idStr, 10, 64); utils.IsErrorEmpty(err, ctx) {
			result, err := d.MethodService.GetMethodDetail(id)
			utils.OperateResponse(result, err, ctx)
		}
	} else {
		utils.ParamMissResponseOperation(ctx)
	}
}

func (d *dubboController) DeleteMethod(ctx *gin.Context) {
	if idStr, ok := ctx.GetQuery("id"); ok {
		if id, err := strconv.ParseInt(idStr, 10, 64); utils.IsErrorEmpty(err, ctx) {
			utils.OperateResponse(nil, d.MethodService.DeleteMethod(id), ctx)
		}
	} else {
		utils.ParamMissResponseOperation(ctx)
	}
}

func (d *dubboController) GetMethodsByReference(ctx *gin.Context) {
	if idStr, ok := ctx.GetQuery("id"); ok {
		if id, err := strconv.ParseInt(idStr, 10, 64); utils.IsErrorEmpty(err, ctx) {
			result, err := d.MethodService.GetMethodInfoByReferenceId(id)
			utils.OperateResponse(result, err, ctx)
		}
	} else {
		utils.ParamMissResponseOperation(ctx)
	}
}

//func (d *dubboController) ListByUserIdAndMethodName(ctx *gin.Context) {
//	if methodName, ok := ctx.GetQuery("methodName"); ok {
//		if user, err := web.GetSessionUser(ctx); utils.IsErrorEmpty(err, ctx) {
//			result, err := d.MethodService.ListByUserIdAndMethodName(user.ID, methodName)
//			utils.OperateResponse(result, err, ctx)
//		}
//	} else {
//		utils.ParamMissResponseOperation(ctx)
//	}
//}

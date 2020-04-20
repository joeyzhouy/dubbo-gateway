package console

import (
	"dubbo-gateway/common/extension"
	"dubbo-gateway/common/utils"
	"dubbo-gateway/service"
	"dubbo-gateway/service/entry"
	"dubbo-gateway/web"
	"github.com/gin-gonic/gin"
)

func init() {
	metaData, err := extension.GetMeta()
	if err != nil {
		panic("get meta error")
	}
	u := &userController{metaData.NewCommonService()}
	uGroup := web.AuthGroup().Group("u")
	uGroup.POST("/login", u.Login)
	uGroup.POST("/", u.CreateUser)
	uGroup.PUT("/", u.UpdatePassword)
}

type userController struct {
	service.CommonService
}

func (u *userController) Login(ctx *gin.Context) {
	user := new(entry.User)
	if utils.IsErrorEmpty(ctx.ShouldBindJSON(user), ctx) {
		if user.Name == "" || user.Password == "" {
			utils.ParamMissResponseOperation(ctx)
			return
		}
		if dbUser, err := u.GetUser(user.Name, user.Password);
			utils.IsErrorEmpty(err, ctx) {
			utils.OperateResponse(nil, web.SaveUser(dbUser, ctx), ctx)
		}
	}
}

func (u *userController) CreateUser(ctx *gin.Context) {
	user := new(entry.User)
	if utils.IsErrorEmpty(ctx.ShouldBindJSON(user), ctx) {
		if user.Name == "" || user.Password == "" || user.Email == "" {
			utils.ParamMissResponseOperation(ctx)
			return
		}
		utils.OperateResponse(nil, u.CommonService.CreateUser(user), ctx)
	}
}

type UpdatePassword struct {
	UserName    string `json:"userName"`
	Password    string `json:"password"`
	OldPassword string `json:"oldPassword"`
}

func (u *userController) UpdatePassword(ctx *gin.Context) {
	updatePassword := new(UpdatePassword)
	if utils.IsErrorEmpty(ctx.ShouldBindJSON(updatePassword), ctx) {
		if updatePassword.UserName == "" || updatePassword.Password == "" ||
			updatePassword.OldPassword == "" {
			utils.ParamMissResponseOperation(ctx)
			return
		}
		utils.OperateResponse(nil, u.CommonService.UpdatePassword(&entry.User{Name: updatePassword.UserName,
			Password: updatePassword.Password}, updatePassword.OldPassword), ctx)
	}
}

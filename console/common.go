package console

import (
	"dubbo-gateway/service"
	"dubbo-gateway/service/entry"
	"dubbo-gateway/web"
	"github.com/gin-gonic/gin"
)

func init() {
	u := &userController{service.NewCommonService()}
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
	if isErrorEmpty(ctx.ShouldBindJSON(user), ctx) {
		if user.Name == "" || user.Password == "" {
			ParamMissResponseOperation(ctx)
			return
		}
		if dbUser, err := u.GetUser(user.Name, user.Password);
			isErrorEmpty(err, ctx) {
			operateResponse(nil, web.SaveUser(dbUser, ctx), ctx)
		}
	}
}

func (u *userController) CreateUser(ctx *gin.Context) {
	user := new(entry.User)
	if isErrorEmpty(ctx.ShouldBindJSON(user), ctx) {
		if user.Name == "" || user.Password == "" || user.Email == "" {
			ParamMissResponseOperation(ctx)
			return
		}
		operateResponse(nil, u.CommonService.CreateUser(user), ctx)
	}
}

type UpdatePassword struct {
	UserName    string `json:"userName"`
	Password    string `json:"password"`
	OldPassword string `json:"oldPassword"`
}

func (u *userController) UpdatePassword(ctx *gin.Context) {
	updatePassword := new(UpdatePassword)
	if isErrorEmpty(ctx.ShouldBindJSON(updatePassword), ctx) {
		if updatePassword.UserName == "" || updatePassword.Password == "" ||
			updatePassword.OldPassword == "" {
			ParamMissResponseOperation(ctx)
			return
		}
		operateResponse(nil, u.CommonService.UpdatePassword(&entry.User{Name: updatePassword.UserName,
			Password: updatePassword.Password}, updatePassword.OldPassword), ctx)
	}
}

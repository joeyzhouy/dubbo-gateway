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
		if user.Name == "" || user.Password == "" || user.Email == "" {
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

}

func (u *userController) UpdatePassword(ctx *gin.Context) {

}

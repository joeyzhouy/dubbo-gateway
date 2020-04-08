package console

import (
	"github.com/apache/dubbo-go/common/logger"
	"github.com/gin-gonic/gin"
)

const (
	Success             = 200
	Fail                = 500
	NotFound            = 501
	UnknowOperation     = 502
	JsonError           = 503
	ParamMiss           = 599
	InvalidArgu         = 600
	NoHeaderInfo        = 601
	UserOrPasswordError = 701
	DbError             = 702
	ServiceDeployError  = 703
	Forbit              = 403
	LogoutError         = 704
)

type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func SuccessResponse() *Response {
	return &Response{Code: Success}
}

func FailResponseOperation(ctx *gin.Context) {
	ctx.JSON(200, &Response{Code: Fail, Message: "system error"})
}

func ParamMissResponseOperation(ctx *gin.Context) {
	ctx.JSON(200, &Response{Code: ParamMiss, Message: "param miss"})
}

func isErrorEmpty(err error, ctx *gin.Context) bool {
	if err != nil {
		logger.Error(err.Error(), err)
		ctx.JSON(200, &Response{Code: Fail, Message: err.Error()})
		return false
	}
	return true
}

func operateResponse(data interface{}, err error, ctx *gin.Context) {
	if err != nil {
		logger.Error(err.Error(), err)
		ctx.JSON(200, &Response{Code: Fail, Message: err.Error()})
		return
	}
	var result *Response
	if data == nil {
		result = SuccessResponse()
	} else {
		result = &Response{Code: Success, Data: data}
	}
	ctx.JSON(200, result)
}

func successResponse(data interface{}, ctx *gin.Context) {
	operateResponse(data, nil, ctx)
}

func successResponseWithoutData(ctx *gin.Context) {
	operateResponse(nil, nil, ctx)
}

func failResponse(err error, ctx *gin.Context) {
	operateResponse(nil, err, ctx)
}

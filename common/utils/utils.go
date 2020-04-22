package utils

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"github.com/apache/dubbo-go/common/logger"
	"github.com/gin-gonic/gin"
	"net"
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

	GET    = "get"
	POST   = "post"
	PUT    = "put"
	DELETE = "delete"
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

func IsErrorEmpty(err error, ctx *gin.Context) bool {
	if err != nil {
		logger.Error(err.Error(), err)
		ctx.JSON(200, &Response{Code: Fail, Message: err.Error()})
		return false
	}
	return true
}

func OperateResponse(data interface{}, err error, ctx *gin.Context) {
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

func OperateSuccessResponse(data interface{}, ctx *gin.Context) {
	OperateResponse(data, nil, ctx)
}

func SuccessResponseWithoutData(ctx *gin.Context) {
	OperateResponse(nil, nil, ctx)
}

func FailResponse(err error, ctx *gin.Context) {
	OperateResponse(nil, err, ctx)
}

func GetLocalIp() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}
	for _, address := range addrs {
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String(), nil
			}
		}
	}
	return "", err
}

func Sha256(message string, secret string) string {
	key := []byte(secret)
	h := hmac.New(sha256.New, key)
	h.Write([]byte(message))
	sha := hex.EncodeToString(h.Sum(nil))
	return base64.StdEncoding.EncodeToString([]byte(sha))
}

func Hash(str string) string {
	h := md5.New()
	h.Write([]byte(str))
	return hex.EncodeToString(h.Sum(nil))
}
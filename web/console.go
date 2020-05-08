package web

import (
	"dubbo-gateway/common/config"
	"dubbo-gateway/common/constant"
	"dubbo-gateway/common/extension"
	"dubbo-gateway/common/utils"
	"dubbo-gateway/conf"
	"dubbo-gateway/service/entry"
	"encoding/json"
	"fmt"
	"github.com/apache/dubbo-go/common/logger"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-contrib/sessions/redis"
	"github.com/gin-gonic/gin"
	perrors "github.com/pkg/errors"
	"gopkg.in/yaml.v2"
	"net/http"
	"path/filepath"
	"strings"
)

const (
	COOKIE        = "cookie"
	REDIS         = "redis"
	consoleOrigin = "console"
	SessionUser   = "_user"
)

var con *console
var ignoreUrlMap = make(map[string]string, 0)

type console struct {
	r         *gin.Engine
	authGroup *gin.RouterGroup
	webConfig *config.WebConfig
}

func (c *console) Start() {
	logger.Infof("console start port: %d", c.webConfig.Config.Port)
	err := c.r.Run(fmt.Sprintf(":%d", c.webConfig.Config.Port))
	if err != nil {
		logger.Errorf("start console error: %v", perrors.WithStack(err))
	}
	return
}

func (*console) Close() {
}

func init() {
	confStr, err := conf.GetConfig(constant.ConfGatewayFilePath, constant.DefaultGatewayFilePath)
	if err != nil {
		logger.Errorf("get config error: %v", perrors.WithStack(err))
		return
	}
	webConfig := new(config.WebConfig)
	err = yaml.Unmarshal([]byte(confStr), webConfig)
	if err != nil {
		logger.Errorf("yaml.Unmarshal() = error:%v", perrors.WithStack(err))
		return
	}
	store, err := getSessionStore(*webConfig)
	if err != nil {
		logger.Errorf("init session store error: %v", perrors.WithStack(err))
		return
	}
	con = new(console)
	con.webConfig = webConfig
	con.r = gin.New()
	con.r.Use(utils.LoggerWithWriter(), gin.Recovery())
	con.r.Use(sessions.Sessions("session", store))
	resourcesPath, err := filepath.Abs("web/resources")
	if err != nil {
		logger.Errorf("init static resource  error: %v", perrors.WithStack(err))
		return
	}
	con.r.StaticFS("/static", http.Dir(resourcesPath+"/static"))
	con.r.StaticFile("/seltek.ico", resourcesPath+"/static/seltek.ico")
	con.r.StaticFile("/index.htm", resourcesPath+"/index.htm")
	con.r.StaticFile("/", resourcesPath+"/index.html")
	con.authGroup = con.r.Group("/", Auth())
	extension.SetOrigin(consoleOrigin, con)
}

func RegisterIgnoreUri(uri, method string) {
	if methods, ok := ignoreUrlMap[uri]; ok {
		ignoreUrlMap[uri] = methods + "|" + method
	} else {
		ignoreUrlMap[uri] = method
	}
}

func AuthGroup() *gin.RouterGroup {
	return con.authGroup
}

func Auth() gin.HandlerFunc {
	result := make(map[string]interface{})
	result["code"] = 403
	return func(ctx *gin.Context) {
		uri := ctx.Request.RequestURI
		if methods, ok := ignoreUrlMap[uri]; ok && !strings.Contains(methods, ctx.Request.Method) || !ok {
			session := sessions.Default(ctx)
			user := session.Get(SessionUser)
			if user == nil {
				ctx.AbortWithStatusJSON(200, &result)
				return
			}
		}
		ctx.Next()
	}
}

func SaveUser(user *entry.User, ctx *gin.Context) error {
	bs, err := json.Marshal(user)
	if err != nil {
		return err
	}
	session := sessions.Default(ctx)
	session.Set(SessionUser, string(bs))
	return session.Save()
}

func GetSessionUser(ctx *gin.Context) (*entry.User, error) {
	session := sessions.Default(ctx)
	userStr := session.Get(SessionUser)
	user := new(entry.User)
	if err := json.Unmarshal([]byte(userStr.(string)), user); err != nil {
		return nil, err
	}
	return user, nil
}

func getSessionStore(config config.WebConfig) (sessions.Store, error) {
	switch config.Config.Session.Type {
	case COOKIE:
		return cookie.NewStore([]byte("secret")), nil
	case REDIS:
		redisConfig := config.Config.Session.Redis
		return redis.NewStore(redisConfig.DB, redisConfig.Network,
			redisConfig.Address, redisConfig.Password, []byte("secret"))
	}
	return nil, perrors.Errorf("not support session type: %s", config.Config.Session.Type)
}

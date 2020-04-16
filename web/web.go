package web

import (
	"dubbo-gateway/common/constant"
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
)

const (
	COOKIE = "cookie"
	REDIS  = "redis"

	SessionUser = "_user"
)

type WebConfig struct {
	Name   string `yaml:"name"`
	Port   int    `yaml:"port"`
	Config struct {
		Session struct {
			Type    string `yaml:"type"`
			Timeout int    `yaml:"time_out"`
			Redis   struct {
				Network  string `yaml:"network"`
				Address  string `yaml:"address"`
				Password string `yaml:"password"`
				DB       int    `yaml:"db"`
			}
		} `yaml:"session"`
	} `yaml:"web"`
}

var r *gin.Engine
var authGroup *gin.RouterGroup
var webConfig *WebConfig

func init() {
	confStr, err := conf.GetConfig(constant.ConfGatewayFilePath, constant.DefaultGatewayFilePath)
	if err != nil {
		logger.Errorf("get config error: %v", perrors.WithStack(err))
		return
	}
	webConfig := new(WebConfig)
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
	r = gin.New()
	r.Use(utils.LoggerWithWriter(), gin.Recovery())
	r.Use(sessions.Sessions("session", store))
	resourcesPath, err := filepath.Abs("web/resources")
	if err != nil {
		logger.Errorf("init static resource  error: %v", perrors.WithStack(err))
		return
	}
	r.StaticFS("/static", http.Dir(resourcesPath+"/static"))
	r.StaticFile("/seltek.ico", resourcesPath+"/static/seltek.ico")
	r.StaticFile("/index.htm", resourcesPath+"/index.htm")
	r.StaticFile("/", resourcesPath+"/index.html")
	authGroup = r.Group("/", Auth())
}

func GetEngine() *gin.Engine {
	return r
}

func AuthGroup() *gin.RouterGroup {
	return authGroup
}

func Run() error {
	return r.Run(fmt.Sprintf(":%d", webConfig.Port))
}

func Auth() gin.HandlerFunc {
	result := make(map[string]interface{})
	result["code"] = 403
	return func(ctx *gin.Context) {
		session := sessions.Default(ctx)
		user := session.Get(SessionUser)
		if user == nil {
			ctx.AbortWithStatusJSON(200, &result)
			return
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

func getSessionStore(config WebConfig) (sessions.Store, error) {
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

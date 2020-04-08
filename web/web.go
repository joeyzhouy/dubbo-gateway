package web

import (
	"bufio"
	"bytes"
	"dubbo-gateway/common/constant"
	"dubbo-gateway/conf"
	"dubbo-gateway/service/entry"
	"encoding/json"
	"github.com/apache/dubbo-go/common/logger"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-contrib/sessions/redis"
	"github.com/gin-gonic/gin"
	perrors "github.com/pkg/errors"
	"gopkg.in/yaml.v2"
	"net/http"
	"path/filepath"
	"time"
)

const (
	COOKIE = "cookie"
	REDIS  = "redis"

	SessionUser = "_user"
)

type WebConfig struct {
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

func init() {
	confStr, err := conf.GetConfig(constant.CONF_GATEWAY_FILE_PATH, "")
	if err != nil {
		logger.Errorf("get config error: %v", perrors.WithStack(err))
		return
	}
	config := new(WebConfig)
	err = yaml.Unmarshal([]byte(confStr), config)
	if err != nil {
		logger.Errorf("yaml.Unmarshal() = error:%v", perrors.WithStack(err))
		return
	}
	store, err := getSessionStore(*config)
	if err != nil {
		logger.Errorf("init session store error: %v", perrors.WithStack(err))
		return
	}
	r.Use(LoggerWithWriter(), gin.Recovery())
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

func LoggerWithWriter(notlogged ...string) gin.HandlerFunc {
	var skip map[string]struct{}
	if length := len(notlogged); length > 0 {
		skip = make(map[string]struct{}, length)
		for _, path := range notlogged {
			skip[path] = struct{}{}
		}
	}

	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery
		w := bufio.NewWriter(c.Writer)
		buff := bytes.Buffer{}
		newWriter := &bufferedWriter{c.Writer, w, buff}
		c.Writer = newWriter
		c.Next()
		if _, ok := skip[path]; !ok {
			end := time.Now()
			latency := end.Sub(start)
			clientIP := c.ClientIP()
			method := c.Request.Method
			statusCode := c.Writer.Status()
			if raw != "" {
				path = path + "?" + raw
			}
			//log.Infof(" | %d | %13v | %15s | %s | %s ", statusCode, latency, clientIP, method, path)
			logger.Infof(" | %d | %13v | %15s | %s | %s /n %s", statusCode, latency, clientIP, method, path, newWriter.Buffer.Bytes())
			_ = w.Flush()
		}
	}
}

type bufferedWriter struct {
	gin.ResponseWriter
	out    *bufio.Writer
	Buffer bytes.Buffer
}

func (g *bufferedWriter) Write(data []byte) (int, error) {
	g.Buffer.Write(data)
	return g.out.Write(data)
}

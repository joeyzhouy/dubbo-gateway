package multiple

import (
	"dubbo-gateway/common/constant"
	"dubbo-gateway/common/extension"
	"dubbo-gateway/common/utils"
	"dubbo-gateway/communication/single"
	"dubbo-gateway/registry"
	"encoding/json"
	"fmt"
	"github.com/apache/dubbo-go/common/logger"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"net/http"
	"strconv"
	"sync"
)

const (
	MultipleMode = "multiple"
	success      = "success"
	failed       = "failed"
	addUrl       = "http://%s/add?apiId=%d"
	removeUrl    = "http://%s/remove?apiId=%d"
	refreshUrl   = "http://%s/refresh"
)

type multipleMode struct {
	r         *gin.Engine
	authGroup *gin.RouterGroup
	ipList    []string
	ports     []int
	reg       registry.Registry
	loc       string
	port      int
	single    extension.Mode
	retry     int
	sync.RWMutex
}

type kv struct {
	key   string
	value string
}

func (m *multipleMode) getAddresses() []string {
	result := make([]string, 0, len(m.ipList))
	ports := make([]int, 0, len(m.ports))
	m.RLock()
	defer m.RUnlock()
	for index, value := range m.ipList {
		result[index] = fmt.Sprintf("%s:%d", value, ports[index])
	}
	return result
}

func (m *multipleMode) getIps() []string {
	result := make([]string, 0, len(m.ipList))
	m.RLock()
	defer m.RUnlock()
	for index, value := range m.ipList {
		result[index] = value
	}
	return result
}

func (m *multipleMode) Add(apiId int64) error {
	if err := m.single.Add(apiId); err != nil {
		return err
	}
	return m.notify(func(address string) string {
		return fmt.Sprintf(addUrl, address, apiId)
	}, func() error {
		return m.Remove(apiId)
	})
}

func (m *multipleMode) notify(getUrl func(string) string, callBack func() error) error {
	addresses := m.getAddresses()
	for i := m.retry; i > 0; i-- {
		ch := make(chan kv, len(addresses))
		for _, address := range addresses {
			go func(c <-chan kv, address string) {
				request := getUrl(address)
				resp, err := http.Get(request)
				if err != nil {
					logger.Errorf("requestUrl: %s, error: %v", request, err)
					ch <- kv{key: address, value: failed}
					return
				}
				defer resp.Body.Close()
				if resp.StatusCode != 200 {
					logger.Errorf("requestUrl: %s, statusCode: %d", request, resp.StatusCode)
					ch <- kv{key: address, value: failed}
					return
				}
				bs, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					logger.Errorf("requestUrl: %s, read request body error: %v", request, err)
					ch <- kv{key: address, value: failed}
					return
				}
				data := new(utils.Response)
				err = json.Unmarshal(bs, data)
				if err != nil {
					logger.Errorf("requestUrl: %s, json umarshal request body: %s error: %v", request, string(bs), err)
					ch <- kv{key: address, value: failed}
					return
				}
				if data.Code != utils.Success {
					ch <- kv{key: address, value: failed}
					return
				}
				ch <- kv{key: address, value: success}
			}(ch, address)
		}
		temp := make([]string, 0)
		for i := 0; i < len(addresses); i++ {
			if result := <-ch; result.value != success {
				temp = append(temp, result.key)
			}
		}
		close(ch)
		if len(temp) == 0 {
			break
		}
		addresses = temp
	}
	if len(addresses) == 0 {
		return nil
	}
	if callBack != nil {
		return callBack()
	}
	return nil
}

func (m *multipleMode) Remove(apiId int64) error {
	if err := m.single.Remove(apiId); err != nil {
		return err
	}
	return m.notify(func(address string) string {
		return fmt.Sprintf(removeUrl, address, apiId)
	}, nil)
}

func (m *multipleMode) Refresh() error {
	return m.notify(func(address string) string {
		return fmt.Sprintf(refreshUrl, address)
	}, nil)
}

func (m *multipleMode) Start() error {
	m.authGroup.GET("/add", func(ctx *gin.Context) {
		if str := ctx.Param("apiId"); str != "" {
			if apiId, err := strconv.ParseInt(str, 10, 64); utils.IsErrorEmpty(err, ctx) {
				utils.OperateResponse(nil, m.single.Add(apiId), ctx)
			}
			return
		}
		utils.ParamMissResponseOperation(ctx)
	})
	m.authGroup.GET("/remove", func(ctx *gin.Context) {
		if str := ctx.Param("apiId"); str != "" {
			if apiId, err := strconv.ParseInt(str, 10, 64); utils.IsErrorEmpty(err, ctx) {
				utils.OperateResponse(nil, m.single.Remove(apiId), ctx)
			}
			return
		}
		utils.ParamMissResponseOperation(ctx)
	})
	m.authGroup.GET("/refresh", func(ctx *gin.Context) {
		utils.OperateResponse(nil, m.single.Refresh(), ctx)
	})
	return m.r.Run(fmt.Sprintf(":%d", m.port))
}

func init() {
	extension.SetMode(MultipleMode, newMultipleMode)
}

func newMultipleMode(deploy *extension.Deploy) (extension.Mode, error) {
	mode := new(multipleMode)
	mode.r = gin.New()
	mode.r.Use(extension.LoggerWithWriter(), gin.Recovery())
	mode.authGroup = mode.r.Group("/", auth(mode))
	mConfig := deploy.Config.Multiple
	mode.port = mConfig.Port
	mode.retry = mConfig.Retry
	var err error
	mode.single, err = extension.GetMode(single.SingleMode)
	if err != nil {
		return nil, err
	}
	mode.reg, err = extension.GetRegistry(mConfig.Coordination.Protocol)
	if err != nil {
		return nil, err
	}
	mode.loc, err = utils.GetLocalIp()
	if err != nil {
		return nil, err
	}
	nodes, err := mode.reg.ListNodeByPath(constant.NodePath)
	if err != nil {
		return nil, err
	}
	mode.ipList = make([]string, 0, len(nodes))
	mode.ports = make([]int, 0, len(nodes))
	for index, node := range nodes {
		mode.ipList[index] = node.IP
		mode.ports[index] = node.Port
	}
	err = mode.reg.RegisterTempNode(extension.Node{
		Port: mode.port,
		IP:   mode.loc,
	})
	err = mode.reg.Subscribe(constant.NodePath, func(event *registry.Event) {
		nodes, err := mode.reg.ListNodeByPath(constant.NodePath)
		if err != nil {
			logger.Errorf("registry get children nodes, parent path: %s, error: %v", constant.NodePath, err)
			return
		}
		ips := make([]string, 0)
		ports := make([]int, 0)
		for _, node := range nodes {
			if node.IP == mode.loc && node.Port == mode.port {
				continue
			}
			ips = append(ips, node.IP)
			ports = append(ports, node.Port)
		}
		mode.Lock()
		defer mode.Unlock()
		mode.ipList = ips
		mode.ports = ports
	})
	if err != nil {
		return nil, err
	}
	return mode, err
}

func auth(mode *multipleMode) gin.HandlerFunc {
	result := make(map[string]interface{})
	result["code"] = 403
	return func(ctx *gin.Context) {
		ips := mode.getIps()
		if len(ips) == 0 {
			ctx.AbortWithStatusJSON(200, &result)
			return
		}
		remoteAddress := ctx.Request.RemoteAddr
		for _, ip := range ips {
			if ip == remoteAddress {
				ctx.Next()
				return
			}
		}
		ctx.AbortWithStatusJSON(200, &result)
	}
}

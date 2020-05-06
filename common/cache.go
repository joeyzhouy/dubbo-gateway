package common

import (
	"dubbo-gateway/common/constant"
	"github.com/go-errors/errors"
	"net/http"
	"strings"
	"sync"
)

const (
	SelfKey = "_self"
)

var parsing map[string]func(key string, param map[string]interface{}) (interface{}, error)

func init() {
	parsing[constant.RouterHeaderKey] =
		func(key string, param map[string]interface{}) (interface{}, error) {
			header := param[constant.RouterHeaderKey].(http.Header)
			result := header[key]
			if len(result) == 0 {
				return nil, nil
			}
			return result[0], nil
		}
	parsing[constant.RouterBodyKey] =
		func(key string, param map[string]interface{}) (interface{}, error) {
			if key == SelfKey {
				return param[constant.RouterBodyKey], nil
			}
			st := param[constant.RouterBodyKey]
			for _, p := range strings.Split(key, ".") {
				if st == nil {
					return nil, nil
				}
				temp := st.(map[string]interface{})
				st = temp[p]
			}
			return st, nil
		}
	parsing[constant.RouterQueryKey] =
		func(key string, param map[string]interface{}) (interface{}, error) {
			params := param[constant.RouterQueryKey].(map[string]interface{})
			return params[key], nil
		}
	parsing[constant.CustomKey] =
		func(key string, param map[string]interface{}) (i interface{}, e error) {
			params := param[constant.CustomKey]
			for _, p := range strings.Split(key, ".") {
				if params == nil {
					return nil, nil
				}
				temp := params.(map[string]interface{})
				params = temp[p]
			}
			return params, nil
		}
}

type ApiCache struct {
	sync.RWMutex
	mappings map[string]*ApiInfo
}

func (a *ApiCache) SetApiInfos(mappings map[string]*ApiInfo) {
	a.Lock()
	defer a.Unlock()
	a.mappings = mappings
}

func (a *ApiCache) SetAPiInfo(apiInfo *ApiInfo) {
	a.Lock()
	defer a.Unlock()
	if a.mappings == nil {
		a.mappings = make(map[string]*ApiInfo)
	}
	a.mappings[apiInfo.Method] = apiInfo
}

func (a *ApiCache) GetByMethodName(methodName string) *ApiInfo {
	a.RLock()
	defer a.RUnlock()
	return a.mappings[methodName]
}

func (a *ApiCache) RemoveByMethods(methodNames []string) {
	if methodNames == nil || len(methodNames) == 0 {
		return
	}
	a.Lock()
	defer a.Unlock()
	for _, key := range methodNames {
		delete(a.mappings, key)
	}
}

type ApiInfo struct {
	Method      string    `json:"method"`
	ApiId       int64     `json:"apiId"`
	FilterChain *ApiChain `json:"filter"`
	MethodChain *ApiChain `json:"method"`
}

type ApiChain struct {
	ChainId     int64             `json:"chainId"`
	ReferenceId int64             `json:"referenceId"`
	MethodName  string            `json:"methodName"`
	ParamClass  []string          `json:"paramClass"`
	ParamRule   []ApiParamExplain `json:"paramRule"`
	ResultRule  []ApiParamExplain `json:"resultRule"`
	Next        *ApiChain         `json:"next"`
}

type ApiParamExplain map[string]*FiledExpression

func (a *ApiParamExplain) Convert(params map[string]interface{}) (interface{}, error) {
	filedExpression, ok := (*a)[SelfKey]
	if ok {
		return filedExpression.parse(params)
	}
	var (
		fieldValue interface{}
		err        error
	)
	result := make(map[string]interface{})
	for key, field := range *a {
		if fieldValue, err = field.parse(params); err != nil {
			result[key] = fieldValue
		}
	}
	return result, nil
}

type FiledExpression struct {
	Static     interface{} `json:"_static"`
	Expression Expression  `json:"_expression"`
	ApiParamExplain
}

func (f *FiledExpression) parse(params map[string]interface{}) (interface{}, error) {
	if f.Static != nil {
		return f.Static, nil
	}
	if f.ApiParamExplain != nil {
		return f.ApiParamExplain.Convert(params)
	}
	return f.Expression.Get(params)
}

type Expression interface {
	Get(map[string]interface{}) (interface{}, error)
}

type SingleExpression struct {
	Prefix string
	Path   string
}

func (s *SingleExpression) Get(params map[string]interface{}) (interface{}, error) {
	return parsing[s.Prefix](s.Path, params)
}

type ArrayExpression struct {
	SingleExpression
	ApiParamExplain
}

func (a *ArrayExpression) Get(params map[string]interface{}) (interface{}, error) {
	arr, err := a.SingleExpression.Get(params)
	if err != nil {
		return nil, err
	}
	if len(a.ApiParamExplain) == 0 {
		return arr, nil
	}
	list, ok := arr.([]interface{})
	if !ok {
		return nil, errors.New("path: " + a.Path + "not array")
	} else if len(list) == 0 {
		return nil, nil
	}
	result := make([]interface{}, 0, len(list))
	for _, en := range list {
		params[constant.CustomKey] = en
		temp, err := a.ApiParamExplain.Convert(params)
		if err != nil {
			return nil, err
		}
		result = append(result, temp)
	}
	delete(params, constant.CustomKey)
	return result, nil
}

type GatewayCache interface {
	Invoke(method string, params map[string]interface{}) (interface{}, error)
}

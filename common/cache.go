package common

import (
	"dubbo-gateway/common/constant"
	"github.com/go-errors/errors"
	"net/http"
	"regexp"
	"strings"
	"sync"
)

const (
	SelfKey       = "_self"
	StaticKey     = "_static"
	ExpressionKey = "_expression"
)

var expressionReg = regexp.MustCompile(`\${(.*?)\.(.*)}`)
var parsing = make(map[string]func(key string, param map[string]interface{}) (interface{}, error))

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
	filters  map[string]*ApiFilter
}

func (a *ApiCache) Exist(filterId string) bool {
	_, ok := a.filters[filterId]
	return ok
}

func (a *ApiCache) SetFilters(filtersMap map[string]*ApiFilter) {
	a.Lock()
	defer a.Unlock()
	a.filters = filtersMap
}

func (a *ApiCache) SetFilter(filterId string, filter *ApiFilter) {
	a.Lock()
	defer a.Unlock()
	a.filters[filterId] = filter
}

func (a *ApiCache) GetFilter(filterId string) (*ApiFilter, bool) {
	a.RLock()
	defer a.RUnlock()
	result, ok := a.filters[filterId]
	return result, ok
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
	FilterId    int64     `json:"filterId"`
	MethodChain *ApiChain `json:"method"`
}

type ApiFilter struct {
	FilterId    int64              `json:"filterId"`
	ReferenceId int64              `json:"referenceId"`
	MethodName  string             `json:"methodName"`
	ParamClass  []string           `json:"paramClass"`
	ParamTypes  []int              `json:"paramTypes"`
	ParamRule   []*ApiParamExplain `json:"paramRule"`
}

type ApiChain struct {
	ChainId     int64              `json:"chainId"`
	ReferenceId int64              `json:"referenceId"`
	MethodName  string             `json:"methodName"`
	ParamClass  []string           `json:"paramClass"`
	ParamRule   []*ApiParamExplain `json:"paramRule"`
	ParamTypes  []int              `json:"paramTypes"`
	ResultRule  []*ApiParamExplain `json:"resultRule"`
	Next        *ApiChain          `json:"next"`
}

type ApiParamExplain map[string]interface{}

func (a *ApiParamExplain) Convert(params map[string]interface{}) (interface{}, error) {
	var (
		ok         bool
		expression Expression
		err        error
		value      interface{}
	)
	value, ok = (*a)[SelfKey]
	if ok {
		expression, err = CreateExpression(value.(map[string]interface{}))
	} else {
		expression, err = CreateExpression(*a)
	}
	if err != nil {
		return nil, err
	}
	return expression.Get(params)
}

func CreateExpression(expression map[string]interface{}) (Expression, error) {
	var (
		value interface{}
		ok    bool
	)
	if value, ok = expression[StaticKey]; ok {
		return &staticExpression{Value: value}, nil
	} else if value, ok = expression[ExpressionKey]; ok {
		return newSingleExpression(value.(string))
	} else {
		return newObjectExpression(value.(map[string]interface{}))
	}
}

type Expression interface {
	Get(map[string]interface{}) (interface{}, error)
}

type staticExpression struct {
	Value interface{}
}

func (s *staticExpression) Get(map[string]interface{}) (interface{}, error) {
	return s.Value, nil
}

type singleExpression struct {
	Prefix string
	Path   string
}

func newSingleExpression(expression string) (*singleExpression, error) {
	strs := expressionReg.FindStringSubmatch(expression)
	if len(strs) != 2 {
		return nil, errors.New("error expression: " + expression)
	}
	return &singleExpression{
		Prefix: strs[0],
		Path:   strs[1],
	}, nil
}

func (s *singleExpression) Get(params map[string]interface{}) (interface{}, error) {
	return parsing[s.Prefix](s.Path, params)
}

type objectExpression struct {
	mappings map[string]Expression
}

func newObjectExpression(fieldMap map[string]interface{}) (*objectExpression, error) {
	mappings := make(map[string]Expression)
	var err error
	for key, value := range fieldMap {
		mappings[key], err = CreateExpression(value.(map[string]interface{}))
		if err != nil {
			return nil, err
		}
	}
	return &objectExpression{
		mappings: mappings,
	}, nil
}

func (o *objectExpression) Get(param map[string]interface{}) (interface{}, error) {
	result := make(map[string]interface{})
	var err error
	for key, mapping := range o.mappings {
		if result[key], err = mapping.Get(param); err != nil {
			return nil, err
		}
	}
	return result, nil
}

//type ArrayExpression struct {
//	SingleExpression
//	ApiParamExplain
//}
//
//func (a *ArrayExpression) Get(params map[string]interface{}) (interface{}, error) {
//	arr, err := a.SingleExpression.Get(params)
//	if err != nil {
//		return nil, err
//	}
//	if len(a.ApiParamExplain) == 0 {
//		return arr, nil
//	}
//	list, ok := arr.([]interface{})
//	if !ok {
//		return nil, errors.New("path: " + a.Path + "not array")
//	} else if len(list) == 0 {
//		return nil, nil
//	}
//	result := make([]interface{}, 0, len(list))
//	for _, en := range list {
//		params[constant.CustomKey] = en
//		temp, err := a.ApiParamExplain.Convert(params)
//		if err != nil {
//			return nil, err
//		}
//		result = append(result, temp)
//	}
//	delete(params, constant.CustomKey)
//	return result, nil
//}

type GatewayCache interface {
	Invoke(method string, params map[string]interface{}) (interface{}, error)
}

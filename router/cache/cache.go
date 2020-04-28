package cache

import (
	"dubbo-gateway/common/extension"
	"dubbo-gateway/common/utils"
	"dubbo-gateway/service"
	"dubbo-gateway/service/entry"
	"dubbo-gateway/service/vo"
	"errors"
	"fmt"
	"github.com/apache/dubbo-go/common/logger"
	"github.com/apache/dubbo-go/config"
	"github.com/gin-gonic/gin"
	jsoniter "github.com/json-iterator/go"
	perrors "github.com/pkg/errors"
	"github.com/tidwall/gjson"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"sync"
)

const (
	APPLICATION_JSON = "application/json"
	Filter_Prefix    = "fitler-"
	Chain_prefix     = "chain-"
	requestUri       = "requestUri"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary
var rCache *routerCache
var requestParamOperate map[string]func(*gin.Context) (map[string]interface{}, error)
var cacheOnce sync.Once

func getCache() *routerCache {
	cacheOnce.Do(func() {
		rCache = new(routerCache)
		metaData := extension.GetMeta()
		rCache.RouterService = metaData.NewRouterService()
		rCache.ReferenceService = metaData.NewReferenceService()
		rCache.RegisterService = metaData.NewRegisterService()
		rCache.uris = make(map[string]*apiConfigCache, 0)
		rCache.rMap = make(map[int64]*config.ReferenceConfig, 0)
		rCache.rUri = make(map[int64][]string)
		err := rCache.refresh()
		if err != nil {
			panic(fmt.Sprintf("router cache refresh error: %v", perrors.WithStack(err)))
		}
		requestParamOperate[utils.GET] = func(ctx *gin.Context) (map[string]interface{}, error) {
			result := make(map[string]interface{})
			for _, param := range ctx.Params {
				result[param.Key] = param.Value
			}
			return result, nil
		}
		requestParamOperate[utils.POST] = func(ctx *gin.Context) (map[string]interface{}, error) {
			result := make(map[string]interface{})
			data, err := ioutil.ReadAll(ctx.Request.Body)
			if err != nil {
				return nil, err
			}
			err = json.Unmarshal(data, &result)
			return result, err
		}
	})
	return rCache
}

type routerCache struct {
	sync.RWMutex
	uris map[string]*apiConfigCache
	rMap map[int64]*config.ReferenceConfig
	rUri map[int64][]string
	service.RouterService
	service.ReferenceService
	service.RegisterService
}

type apiConfigCache struct {
	vo.ApiConfigInfo
	filterReference *referenceCache
	chainReferences *referenceChainCache
}

type requestParam struct {
	params map[string]interface{}
	header http.Header
	other  map[string]interface{}
}

func (a *apiConfigCache) doFilter(ctx *gin.Context) (bool, *requestParam, error) {
	f, ok := requestParamOperate[strings.ToLower(ctx.Request.Method)]
	if !ok {
		return false, nil, perrors.Errorf("not support request method: " + ctx.Request.Method)
	}
	header := ctx.Request.Header
	other := attachment(ctx)
	params, err := f(ctx)
	if err != nil {
		logger.Errorf("get request param error: %v", err)
		return false, nil, perrors.Errorf("system error")
	}
	reqParam := &requestParam{
		params: params,
		other:  other,
		header: header,
	}
	if a.filterReference != nil {
		return true, reqParam, nil
	}
	filterConfig := a.filterReference
	resp, err := filterConfig.referenceConfig.GetRPCService().(*config.GenericService).
		Invoke([]interface{}{filterConfig.methodName, filterConfig.paramClass, []interface{}{header, params, other}})
	if err != nil {
		logger.Errorf("invoke filter error: interfaceName: %s, methodName: %s, uri: %s", filterConfig.referenceConfig.InterfaceName,
			filterConfig.methodName, ctx.Request.RequestURI)
		return false, nil, perrors.Errorf("system error")
	}
	var result bool
	bs, err := json.Marshal(resp)
	if err != nil {
		logger.Errorf("json marshal error: interfaceName: %s, methodName: %s, uri: %s", filterConfig.referenceConfig.InterfaceName,
			filterConfig.methodName, ctx.Request.RequestURI)
		return false, nil, perrors.Errorf("system error")
	}
	err = json.Unmarshal(bs, &result)
	if err != nil {
		logger.Errorf("json Unmarshal error: interfaceName: %s, methodName: %s, uri: %s, body: %s", filterConfig.referenceConfig.InterfaceName,
			filterConfig.methodName, ctx.Request.RequestURI, string(bs))
		return false, nil, perrors.Errorf("system error")
	}
	return result, reqParam, nil
}

func (a *apiConfigCache) invoke(param *requestParam) (interface{}, error) {
	var result interface{}
	chain := a.chainReferences
	if chain == nil {
		return nil, perrors.Errorf("miss chain config, uri: %s", param.other[requestUri])
	}
	var err error
	var paramClasses []string
	var bs []byte
	var resultString string
	params := []interface{}{param.params}

	for ; chain != nil; chain = chain.next {
		if paramClasses == nil {
			paramClasses = chain.paramClass
		}
		result, err = chain.referenceConfig.GetRPCService().(*config.GenericService).
			Invoke([]interface{}{chain.methodName, paramClasses, params})
		if err != nil {
			return nil, err
		}
		if chain.rules == nil || len(chain.rules) == 0 {
			break
		}
		params = make([]interface{}, 0)
		paramClasses = make([]string, 0)
		bs, err = jsoniter.Marshal(result)
		if err != nil {
			break
		}
		resultString = string(bs)
		for _, rule := range chain.rules {
			paramClasses = append(paramClasses, rule.rule.JavaClass)
			params = append(params, gjson.Get(resultString, rule.rule.Rule).String())
		}
	}
	return result, err
}

type referenceCache struct {
	referenceConfig *config.ReferenceConfig
	methodName      string
	paramClass      []string
	keyPrefix       string
	ID              string
	ReferenceId     int64
}

func (r *referenceCache) key() string {
	return r.keyPrefix + r.ID
}

type referenceChainCache struct {
	referenceCache
	rules []resultRule
	next  *referenceChainCache
}

type resultRule struct {
	rule *entry.ApiResultRule
}

//TODO
func (r *routerCache) clear() error {
	return nil
}

func (r *routerCache) refresh() error {
	apiConfigInfos, err := r.RouterService.ListAll()
	if err != nil {
		//logger.Errorf("get api config info error: %v", perrors.WithStack(err))
		return err
	}
	uris := make(map[string]*apiConfigCache, 0)
	rUri := make(map[int64][]string)
	r.Lock()
	defer r.Unlock()
	if err := r.refreshConsumerConfig(); err != nil {
		//logger.Errorf("set consumer config error: %v", perrors.WithStack(err))
		return err
	}
	referenceMap, err := r.fillReferenceConfig(nil)
	if err != nil {
		//logger.Errorf("fill reference config error: %v", perrors.WithStack(err))
		return err
	}
	for _, apiConfigInfo := range apiConfigInfos {
		apiCache := new(apiConfigCache)
		apiCache.ApiConfigInfo = *apiConfigInfo
		if apiCache.Filter.ID != 0 {
			apiFilter := apiConfigInfo.Filter
			rConfig, ok := referenceMap[apiFilter.ReferenceId]
			if !ok {
				return perrors.Errorf("filter: can not found reference with id: %d, in apiConfig uri: %s", apiFilter.ReferenceId, apiCache.ApiConfig.Uri)
			}
			//TODO filter param consumer
			apiCache.filterReference = &referenceCache{
				ID:              strconv.FormatInt(apiFilter.ID, 10),
				ReferenceId:     apiFilter.ReferenceId,
				keyPrefix:       Filter_Prefix,
				referenceConfig: rConfig,
				methodName:      apiFilter.MethodName,
				paramClass: []string{
					"java.utils.Map",   //header
					"java.lang.String", //requestUri
					"java.lang.String", //requestBody
				},
			}
			temp, ok := rUri[apiFilter.ReferenceId]
			if !ok {
				temp = make([]string, 0)
			}
			temp = append(temp, apiCache.filterReference.key())
			rUri[apiFilter.ReferenceId] = temp
		}
		var origin *referenceChainCache
		for index := len(apiConfigInfo.Chains) - 1; index >= 0; index-- {
			old := origin
			chain := apiConfigInfo.Chains[index]
			rConfig, ok := referenceMap[chain.Chain.ReferenceId]
			if !ok {
				return perrors.Errorf("chain[first]: can not found reference with id: %d, in apiConfig uri: %s", chain.Chain.ReferenceId, apiCache.ApiConfig.Uri)
			}
			params := make([]string, 0, len(chain.Params))
			for _, param := range chain.Params {
				params = append(params, param.Key)
			}
			resultRules := make([]resultRule, 0, len(chain.Rules))
			for _, rule := range chain.Rules {
				resultRules = append(resultRules, resultRule{&rule})
			}
			origin = &referenceChainCache{
				referenceCache: referenceCache{
					ID:              strconv.FormatInt(chain.Chain.ReferenceId, 10),
					ReferenceId:     chain.Chain.ReferenceId,
					methodName:      chain.Method.MethodName,
					keyPrefix:       Chain_prefix,
					referenceConfig: rConfig,
					paramClass:      params,
				},
				next:  old,
				rules: resultRules,
			}
			temp, ok := r.rUri[chain.Chain.ReferenceId]
			if !ok {
				temp = make([]string, 0)
			}
			temp = append(temp, origin.key())
			rUri[chain.Chain.ReferenceId] = temp
		}
		apiCache.chainReferences = origin
		uris[apiCache.ApiConfig.Uri] = apiCache
		r.uris = uris
		r.rUri = rUri
		r.rMap = referenceMap
	}
	return nil
}

func (r *routerCache) fillReferenceConfig(references []entry.Reference) (map[int64]*config.ReferenceConfig, error) {
	if references == nil {
		var err error
		references, err = r.ReferenceService.ListAll()
		if err != nil {
			return nil, err
		}
	}
	if len(references) == 0 {
		return nil, errors.New("no reference info config")
	}
	result := make(map[int64]*config.ReferenceConfig)
	for _, reference := range references {
		rConfig := &config.ReferenceConfig{
			Protocol:      reference.Protocol,
			Cluster:       reference.Cluster,
			Registry:      strconv.FormatInt(reference.RegistryId, 10),
			InterfaceName: reference.InterfaceName,
			Generic:       true,
		}
		rConfig.GenericLoad(strconv.FormatInt(reference.ID, 10))
		result[reference.ID] = rConfig
	}
	return result, nil
}

func (r *routerCache) refreshConsumerConfig() error {
	registries, err := r.RegisterService.ListAll()
	if err != nil {
		return err
	}
	if len(registries) == 0 {
		return errors.New("no registry info config")
	}
	registryMap := make(map[string]*config.RegistryConfig)
	for _, registry := range registries {
		registryMap[strconv.FormatInt(registry.ID, 10)] = &config.RegistryConfig{
			Protocol:   registry.Protocol,
			TimeoutStr: registry.Timeout,
			Address:    registry.Address,
			Username:   registry.UserName,
			Password:   registry.Password,
		}
	}
	check := true
	consumerConfig := config.ConsumerConfig{
		Check:           &check,
		Request_Timeout: "3s",
		Connect_Timeout: "3s",
		ApplicationConfig: &config.ApplicationConfig{
			Name:        "gateway",
			Environment: "prod",
		},
		Registries: registryMap,
	}
	config.SetConsumerConfig(consumerConfig)
	return nil
}

func (r *routerCache) get(uri string) (*apiConfigCache, error) {
	r.RLock()
	defer r.RUnlock()
	if cache, ok := r.uris[uri]; !ok {
		return nil, perrors.Errorf("miss config with uri: %s", uri)
	} else {
		return cache, nil
	}
}

func Operate(ctx *gin.Context) {
	if APPLICATION_JSON != ctx.ContentType() {
		ctx.AbortWithStatusJSON(200,
			&utils.Response{Code: utils.Fail, Message: "just support: application/json"})
		return
	}
	requestUri := ctx.Request.RequestURI
	cache, err := getCache().get(requestUri)
	if err != nil {
		logger.Errorf("get rCache error: %v", perrors.WithStack(err))
		ctx.AbortWithStatusJSON(200,
			&utils.Response{Code: utils.Fail, Message: "no api with uri: " + requestUri})
		return
	}
	ok, param, err := cache.doFilter(ctx)
	if err != nil {
		ctx.AbortWithStatusJSON(200,
			&utils.Response{Code: utils.Fail, Message: "system error"})
		return
	} else if !ok {
		ctx.AbortWithStatusJSON(200,
			&utils.Response{Code: utils.Forbit, Message: "forbit"})
		return
	}
	result, err := cache.invoke(param)
	if err != nil {
		logger.Errorf("invoke error: uri: %s, err: %v", cache.ApiConfigInfo.ApiConfig.Uri,
			perrors.WithStack(err))
		ctx.AbortWithStatusJSON(200,
			&utils.Response{Code: utils.Fail, Message: "system error"})
		return
	}
	bs, err := json.Marshal(result)
	if err != nil {
		logger.Errorf("json Unmarshal result error: uri: %s, body: %s, error:%v",
			cache.ApiConfigInfo.ApiConfig.Uri, string(bs), perrors.WithStack(err))
		return
	}
	ctx.Writer.WriteHeader(200)
	_, _ = ctx.Writer.Write(bs)
}

func attachment(ctx *gin.Context) map[string]interface{} {
	return map[string]interface{}{
		requestUri: ctx.Request.RequestURI,
	}
}

func RemoveKey(key string) error {
	getCache().RLock()
	apiCache, ok := getCache().uris[key]
	if !ok {
		getCache().RUnlock()
		logger.Warnf("no config with uri:%s to remove", key)
		return nil
	}
	filterReference := apiCache.filterReference
	chainReference := apiCache.chainReferences
	removeMap := make(map[int64][]string)
	if filterReference != nil {
		removeMap[filterReference.ReferenceId] = []string{filterReference.key()}
	}
	chain := chainReference
	for ; chain != nil; chain = chain.next {
		temp, ok := removeMap[chain.ReferenceId]
		if !ok {
			temp = make([]string, 0)
		}
		temp = append(temp, chain.key())
		removeMap[chain.ReferenceId] = temp
	}
	getCache().Lock()
	defer getCache().Unlock()
	removeReference := make([]int64, 0)
	for key, arr := range removeMap {
		if old, ok := getCache().rUri[key]; ok {
			for _, removeItem := range arr {
				var index int
				for i, item := range old {
					if item == removeItem {
						index = i
						break
					}
				}
				old = append(old[0:index], old[index+1:]...)
			}
			if len(old) == 0 {
				delete(getCache().rUri, key)
				removeReference = append(removeReference, key)
			}
		}
	}
	if len(removeReference) > 0 {
		for _, referenceId := range removeReference {
			delete(getCache().rMap, referenceId)
			//TODO resource to release
		}
	}
	delete(getCache().uris, key)
	return nil
}

func Remove(apiId int64) error {
	apiConfigInfo, err := getCache().RouterService.GetByApiId(apiId)
	if err != nil {
		return err
	}
	return RemoveKey(apiConfigInfo.ApiConfig.Uri)
}

func Add(apiId int64) error {
	apiConfigInfo, err := getCache().RouterService.GetByApiId(apiId)
	if err != nil {
		return err
	}
	apiCache := new(apiConfigCache)
	apiCache.ApiConfigInfo = *apiConfigInfo
	// refresh registry info
	if err := getCache().refreshConsumerConfig(); err != nil {
		return err
	}
	// fill reference info
	referenceIds := make([]int64, 0)
	if apiCache.Filter.ID != 0 {
		if _, ok := getCache().rMap[apiConfigInfo.Filter.ReferenceId]; !ok {
			referenceIds = append(referenceIds, apiCache.Filter.ReferenceId)
		}
	}
	for _, chain := range apiConfigInfo.Chains {
		if _, ok := getCache().rMap[chain.Chain.ReferenceId]; !ok {
			referenceIds = append(referenceIds, chain.Chain.ReferenceId)
		}
	}
	if len(referenceIds) > 0 {
		references, err := getCache().ReferenceService.GetByIds(referenceIds)
		if err != nil {
			return err
		}
		if referenceMap, err := getCache().fillReferenceConfig(references); err != nil {
			return err
		} else if len(referenceMap) > 0 {
			getCache().Lock()
			for key, value := range referenceMap {
				getCache().rMap[key] = value
			}
			getCache().Unlock()
		}
	}
	rUri := make(map[int64][]string, 0)
	if apiCache.ApiConfig.ID != 0 {
		apiFilter := apiConfigInfo.Filter
		apiCache.filterReference = &referenceCache{
			ID:              strconv.FormatInt(apiFilter.ID, 10),
			ReferenceId:     apiFilter.ReferenceId,
			keyPrefix:       Filter_Prefix,
			referenceConfig: getCache().rMap[apiFilter.ReferenceId],
			methodName:      apiFilter.MethodName,
			paramClass: []string{
				"java.utils.Map",   //header
				"java.lang.String", //requestUri
				"java.lang.String", //requestBody
			},
		}
		temp, ok := rUri[apiFilter.ReferenceId]
		if !ok {
			temp = make([]string, 0)
		}
		temp = append(temp, apiCache.filterReference.key())
		rUri[apiFilter.ReferenceId] = temp
	}
	var origin *referenceChainCache
	for index := len(apiConfigInfo.Chains) - 1; index >= 0; index-- {
		old := origin
		chain := apiConfigInfo.Chains[index]
		params := make([]string, 0, len(chain.Params))
		for _, param := range chain.Params {
			params = append(params, param.Key)
		}
		resultRules := make([]resultRule, 0, len(chain.Rules))
		for _, rule := range chain.Rules {
			resultRules = append(resultRules, resultRule{&rule})
		}
		origin = &referenceChainCache{
			referenceCache: referenceCache{
				ID:              strconv.FormatInt(chain.Chain.ReferenceId, 10),
				ReferenceId:     chain.Chain.ReferenceId,
				methodName:      chain.Method.MethodName,
				keyPrefix:       Chain_prefix,
				referenceConfig: getCache().rMap[chain.Chain.ReferenceId],
				paramClass:      params,
			},
			next:  old,
			rules: resultRules,
		}
		temp, ok := rUri[chain.Method.ReferenceId]
		if !ok {
			temp = make([]string, 0)
		}
		temp = append(temp, origin.key())
		rUri[chain.Chain.ReferenceId] = temp
	}
	apiCache.chainReferences = origin
	getCache().Lock()
	defer getCache().Unlock()
	getCache().uris[apiCache.ApiConfig.Uri] = apiCache
	for key, arr := range rUri {
		old, ok := getCache().rUri[key]
		if ok {
			arr = append(arr, old...)
		}
		getCache().rUri[key] = arr
	}
	return nil
}

func Refresh() error {
	return getCache().refresh()
}

func Close() {
	_ = getCache().clear()
	rCache = nil
}

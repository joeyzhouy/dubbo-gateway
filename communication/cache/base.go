package cache

import (
	"dubbo-gateway/common"
	"dubbo-gateway/common/constant"
	"dubbo-gateway/common/extension"
	"dubbo-gateway/service"
	"dubbo-gateway/service/entry"
	"dubbo-gateway/service/vo"
	"fmt"
	"github.com/apache/dubbo-go/common/logger"
	dubboConfig "github.com/apache/dubbo-go/config"
	"github.com/go-errors/errors"
	perrors "github.com/pkg/errors"
	"strconv"
)

const baseKey = "base"

type base struct {
	service.RegisterService
	service.ReferenceService
	service.RouterService
	*common.ApiCache
}

func (b *base) Invoke(method string, params map[string]interface{}) (interface{}, error) {
	apiInfo := b.GetByMethodName(method)
	if apiInfo == nil {
		return nil, errors.New("no method: " + method)
	}
	var (
		result interface{}
		err    error
	)
	// doFilter
	if apiInfo.FilterChain != nil {
		if _, err = b.doChain(apiInfo.FilterChain, params, constant.ResultChainFilterPrefix); err != nil {
			return nil, err
		}
	}
	// doMethod
	if result, err = b.doChain(apiInfo.MethodChain, params, constant.ResultChainMethodPrefix); err != nil {
		return nil, err
	}
	return result, nil
}

func (b *base) doChain(chain *common.ApiChain, params map[string]interface{}, resultPrefix string /*, doResult func(result interface{})*/) (result interface{}, err error) {
	var (
		referenceConfig   *dubboConfig.ReferenceConfig
		paramClasses      []string
		requestParamValue []interface{}
		resultKey         string
	)
	index := 0
	for c := chain; c != nil; c = c.Next {
		if referenceConfig, err = getReference(strconv.FormatInt(c.ReferenceId, 10)); err != nil {
			return
		}
		paramClasses = c.ParamClass
		if paramClasses == nil || len(paramClasses) == 0 {
			paramClasses = nil
			requestParamValue = nil
		} else if requestParamValue, err = b.buildParameter(c.ParamRule, params); err != nil {
			return
		}
		if len(paramClasses) != len(requestParamValue) {
			return nil, errors.New("len(paramClasses) != len(requestParamValue)")
		}
		result, err = referenceConfig.GetRPCService().(*dubboConfig.GenericService).Invoke(
			[]interface{}{c.MethodName, paramClasses, requestParamValue})
		if err != nil {
			return
		}
		resultKey = fmt.Sprintf(resultPrefix+"%d", index)
		params[resultKey] = result
		if chain.ResultRule != nil && len(chain.ResultRule) > 0 {
			if err = b.reduceResult(chain.ResultRule, params, resultKey); err != nil {
				return
			}
			result = params[resultKey]
		}
		index++
	}
	return
}

func (b *base) reduceResult(explains []common.ApiParamExplain, params map[string]interface{}, resultKey string) error {
	if explains == nil || len(explains) == 0 {
		return nil
	}
	if len(explains) > 1 {
		return errors.New("not support")
	}
	if result, err := explains[0].Convert(params); err != nil {
		return err
	} else {
		params[resultKey] = result
	}
	return nil
}

func (b *base) buildParameter(explains []common.ApiParamExplain, params map[string]interface{}) ([]interface{}, error) {
	if explains == nil || len(explains) == 0 {
		return nil, nil
	}
	paramValues := make([]interface{}, 0, len(explains))
	for index, explain := range explains {
		if temp, err := explain.Convert(params); err != nil {
			return nil, err
		} else {
			paramValues[index] = temp
		}
	}
	return paramValues, nil
}

func (b *base) init() error {
	registries, err := b.RegisterService.ListAll()
	if err != nil {
		return perrors.Errorf("get registries error: %v", err)
	}
	for _, registry := range registries {
		addRegistry(strconv.FormatInt(registry.ID, 10), &dubboConfig.RegistryConfig{
			Protocol:   registry.Protocol,
			TimeoutStr: registry.Timeout,
			Address:    registry.Address,
			Username:   registry.UserName,
			Password:   registry.Password,
		})
	}
	RefreshConsumerConfig()
	references, err := b.ReferenceService.ListAll()
	if err != nil {
		return err
	}
	referenceMap := make(map[int64]entry.Reference)
	for _, reference := range references {
		referenceMap[reference.ID] = reference
	}
	apiConfigInfos, err := b.RouterService.ListAllAvailable()
	if err != nil {
		return err
	}
	b.ApiCache = &common.ApiCache{}
	b.ApiCache.SetApiInfos(make(map[string]*common.ApiInfo))
	for _, apiConfigInfo := range apiConfigInfos {
		info, err := apiConfigInfo.ConvertCache()
		if err != nil {
			return err
		}
		if err = b.addApiInfoInDubbo(info, referenceMap); err != nil {
			return err
		}
	}

	return nil
}

func (b *base) addApiInfoInDubboByAId(apiId int64) error {
	references, err := b.GetReferenceByApiId(apiId)
	if err != nil {
		return err
	}
	referenceMap := make(map[int64]entry.Reference)
	for _, reference := range references {
		referenceMap[reference.ID] = reference
	}
	return b.addApiInfoInDubboByApiId(apiId, referenceMap)
}

func (b *base) addApiInfoInDubboByApiId(apiId int64, referenceMap map[int64]entry.Reference) error {
	apiConfigInfo, err := b.GetByConfigId(apiId)
	if err != nil {
		return err
	}
	return b.addApiInfoInDubboByApiConfigInfo(apiConfigInfo, referenceMap)
}

func (b *base) addApiInfoInDubboByApiConfigInfo(apiConfigInfo *vo.ApiConfigInfo, referenceMap map[int64]entry.Reference) error {
	apiInfo, err := apiConfigInfo.ConvertCache()
	if err != nil {
		return err

	}
	return b.addApiInfoInDubbo(apiInfo, referenceMap)
}

func (b *base) addApiInfoInDubbo(info *common.ApiInfo, referenceMap map[int64]entry.Reference) error {
	for filter := info.FilterChain; filter != nil; filter = filter.Next {
		ref, ok := referenceMap[filter.ReferenceId]
		if !ok {
			return perrors.Errorf("not found filter reference config with Id: %d", filter.ReferenceId)
		}
		if err := addReference(strconv.FormatInt(filter.ReferenceId, 10),
			strconv.FormatInt(filter.ChainId, 10), &dubboConfig.ReferenceConfig{
				Protocol:      ref.Protocol,
				Cluster:       ref.Cluster,
				Registry:      strconv.FormatInt(ref.RegistryId, 10),
				InterfaceName: ref.InterfaceName,
				Generic:       true,
			}); err != nil {
			return err
		}
	}
	for chain := info.MethodChain; chain != nil; chain = chain.Next {
		ref, ok := referenceMap[chain.ReferenceId]
		if !ok {
			return perrors.Errorf("not found method reference config with Id: %d", chain.ReferenceId)
		}
		if err := addReference(strconv.FormatInt(chain.ReferenceId, 10),
			strconv.FormatInt(chain.ChainId, 10), &dubboConfig.ReferenceConfig{
				Protocol:      ref.Protocol,
				Cluster:       ref.Cluster,
				Registry:      strconv.FormatInt(ref.RegistryId, 10),
				InterfaceName: ref.InterfaceName,
				Generic:       true,
			}); err != nil {
			return err
		}
	}
	b.ApiCache.SetAPiInfo(info)
	return nil
}

func NewLocalCache(mode extension.Mode) (common.GatewayCache, error) {
	if mode == nil {
		return nil, errors.New("param mode is empty")
	}
	meta := extension.GetMeta()
	cache := &base{
		RegisterService:  meta.NewRegisterService(),
		ReferenceService: meta.NewReferenceService(),
		RouterService:    meta.NewRouterService(),
	}
	err := cache.init()
	if err != nil {
		panic("create local cache error:" + err.Error())
	}
	if err = mode.SubscribeEvent(extension.Registry, extension.AllType, baseKey, newRegistryOperator(cache)); err != nil {
		return nil, err
	}
	if err = mode.SubscribeEvent(extension.Reference, extension.AllType, baseKey, newReferenceOperator(cache)); err != nil {
		return nil, err
	}
	if err = mode.SubscribeEvent(extension.Api, extension.AllType, baseKey, newApiOperator(cache)); err != nil {
		return nil, err
	}
	if err = mode.SubscribeEvent(extension.Method, extension.AllType, baseKey, newMethodOperator(cache)); err != nil {
		return nil, err
	}
	return cache, nil
}

func newRegistryOperator(base *base) func(extension.ModeEvent) {
	return func(event extension.ModeEvent) {
		if event.Type != extension.Add && event.Type != extension.Modify {
			logger.Infof("receive eventType[%d], eventKey[%d], do nothing", event.Type, event.Key)
			return
		}
		registry, err := base.RegisterService.GetByRegistryId(event.Key)
		if err != nil {
			logger.Errorf("[base] registryOperator operate registry event[%d] error: %v", event.Key, err)
			return
		}
		addRegistry(strconv.FormatInt(registry.ID, 10), &dubboConfig.RegistryConfig{
			Protocol:   registry.Protocol,
			TimeoutStr: registry.Timeout,
			Address:    registry.Address,
			Username:   registry.UserName,
			Password:   registry.Password,
		})
		RefreshConsumerConfig()
	}
}

func newReferenceOperator(base *base) func(extension.ModeEvent) {
	return func(event extension.ModeEvent) {
		// only operate modify & delete
		switch event.Type {
		case extension.Modify:
			reference, err := base.ReferenceService.GetReferenceEntryById(event.Key)
			if err != nil {
				logger.Errorf("GetReferenceEntryById[%d], error: %v", event.Key, err)
				return
			}
			modifyReference(strconv.FormatInt(event.Key, 10), &dubboConfig.ReferenceConfig{
				Protocol:      reference.Protocol,
				Cluster:       reference.Cluster,
				Registry:      strconv.FormatInt(reference.RegistryId, 10),
				InterfaceName: reference.InterfaceName,
				Generic:       true,
			})
		case extension.Delete:
			chainIds := removeAllReference(strconv.FormatInt(event.Key, 10))
			if chainIds != nil && len(chainIds) > 0 {
				// remove api
				methodNames, err := base.RouterService.GetApiMethodNamesByReferenceId(event.Key)
				if err != nil {
					logger.Error("GetApiMethodNamesByReferenceId[%d], error: %v", event.Key, err)
					return
				}
				base.ApiCache.RemoveByMethods(methodNames)
			}
		default:
			logger.Infof("[base] referenceOperator receive eventType[%d], eventKey[%d], do nothing", event.Type, event.Key)
			return
		}
	}
}

func newApiOperator(base *base) func(extension.ModeEvent) {
	return func(event extension.ModeEvent) {
		switch event.Type {
		case extension.Add:
			if err := base.addApiInfoInDubboByAId(event.Key); err != nil {
				logger.Errorf("base.addApiInfoInDubboByAId[%s], error: %v", event.Key, err)
				return
			}
		case extension.Modify:
			removeApiCacheById(event.Key, base)
			if err := base.addApiInfoInDubboByAId(event.Key); err != nil {
				logger.Errorf("base.addApiInfoInDubboByAId[%s], error: %v", event.Key, err)
				return
			}
		case extension.Delete:
			removeApiCacheById(event.Key, base)
		default:
			logger.Infof("[base] apiOperator receive eventType[%d], eventKey[%d], do nothing", event.Type, event.Key)
		}
	}
}

func newMethodOperator(base *base) func(extension.ModeEvent) {
	return func(event extension.ModeEvent) {
		switch event.Type {
		//case extension.Add:
		case extension.Modify, extension.Delete:
			apiIds, err := base.GetApiIdsByMethodId(event.Key)
			if err != nil {
				logger.Errorf("[base] GetApiIdsByMethodId[%d], error: %v", event.Key, err)
				return
			}
			if apiIds != nil {
				for _, apiId := range apiIds {
					removeApiCacheById(apiId, base)
				}
			}
			if event.Type == extension.Delete {
				return
			}
			for _, apiId := range apiIds {
				if err := base.addApiInfoInDubboByAId(apiId); err != nil {
					logger.Errorf("base.addApiInfoInDubboByAId[%d], error: %v", apiId, err)
				}
			}
		default:
			logger.Infof("[base] methodOperator receive eventType[%d], eventKey[%d], do nothing", event.Type, event.Key)
		}
	}
}

func removeApiCacheById(apiId int64, base *base) {
	apiConfig, err := base.RouterService.GetConfigById(apiId)
	if err != nil {
		logger.Errorf("RouterService.GetConfigById[%d], error: %v", apiId, err)
		return
	}
	apiInfo := base.GetByMethodName(apiConfig.Method)
	if apiInfo == nil {
		logger.Infof("not found apiInfo[%s], do nothing", apiConfig.Method)
		return
	}
	base.RemoveByMethods([]string{apiConfig.Method})
	referenceMap := make(map[string][]string)
	var (
		referenceId string
		chainId     string
	)
	if apiInfo.FilterChain != nil {
		for filter := apiInfo.FilterChain; filter != nil; filter = filter.Next {
			referenceId = strconv.FormatInt(filter.ReferenceId, 10)
			chainId = strconv.FormatInt(filter.ChainId, 10)
			identifies, ok := referenceMap[referenceId]
			if !ok {
				identifies = make([]string, 0)
			}
			identifies = append(identifies, chainId)
			referenceMap[referenceId] = identifies
		}
	}
	for chain := apiInfo.MethodChain; chain != nil; chain = chain.Next {
		referenceId = strconv.FormatInt(chain.ReferenceId, 10)
		chainId = strconv.FormatInt(chain.ChainId, 10)
		identifies, ok := referenceMap[referenceId]
		if !ok {
			identifies = make([]string, 0)
		}
		identifies = append(identifies, chainId)
		referenceMap[referenceId] = identifies
	}
	removeReferences(referenceMap)
}

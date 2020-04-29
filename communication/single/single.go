package single

import (
	"dubbo-gateway/common/config"
	"dubbo-gateway/common/extension"
	"dubbo-gateway/communication/cache"
	"dubbo-gateway/service"
	"dubbo-gateway/service/entry"
	"dubbo-gateway/service/vo"
	"github.com/apache/dubbo-go/common/logger"
	dubboConfig "github.com/apache/dubbo-go/config"
	perrors "github.com/pkg/errors"
	"strconv"
)

const SingleMode = "single"

type singleMode struct {
	service.RegisterService
	service.ReferenceService
	service.RouterService
	apiCache *vo.ApiCache
}

func (s *singleMode) Init() error {
	registries, err := s.RegisterService.ListAll()
	if err != nil {
		return perrors.Errorf("get registries error: %v", err)
	}
	for _, registry := range registries {
		cache.AddRegistry(strconv.FormatInt(registry.ID, 10), &dubboConfig.RegistryConfig{
			Protocol:   registry.Protocol,
			TimeoutStr: registry.Timeout,
			Address:    registry.Address,
			Username:   registry.UserName,
			Password:   registry.Password,
		})
	}
	cache.RefreshConsumerConfig()
	references, err := s.ReferenceService.ListAll()
	if err != nil {
		return err
	}
	referenceMap := make(map[int64]entry.Reference)
	for _, reference := range references {
		referenceMap[reference.ID] = reference
	}
	apiConfigInfos, err := s.RouterService.ListAllAvailable()
	if err != nil {
		return err
	}
	apiMaps := make(map[string]*vo.ApiInfo)
	for _, apiConfigInfo := range apiConfigInfos {
		info, err := apiConfigInfo.ConvertCache()
		if err != nil {
			return err
		}
		apiMaps[info.Method] = info
		for chain := info.FilterChain; chain != nil; chain = chain.Next {
			ref, ok := referenceMap[chain.ReferenceId]
			if !ok {
				return perrors.Errorf("not found reference config with Id: %d", chain.ReferenceId)
			}
			err = cache.AddReference(strconv.FormatInt(chain.ReferenceId, 10),
				strconv.FormatInt(chain.ChainId, 10), &dubboConfig.ReferenceConfig{
					Protocol:      ref.Protocol,
					Cluster:       ref.Cluster,
					Registry:      strconv.FormatInt(ref.RegistryId, 10),
					InterfaceName: ref.InterfaceName,
					Generic:       true,
				})
			if err != nil {
				return err
			}
		}
	}
	s.apiCache = &vo.ApiCache{
		Mappings: apiMaps,
	}
	return nil
}

func (s *singleMode) Notify(event extension.ModeEvent) {
	panic("implement me")
}

func (s *singleMode) Close() {
	logger.Info("single mode close")
}

func (s *singleMode) Start() {
	logger.Info("start single mode")
}

func init() {
	extension.SetMode(SingleMode, newSingleMode)
}

func newSingleMode(deploy *config.Deploy) (extension.Mode, error) {
	meta := extension.GetMeta()
	mode := &singleMode{
		RegisterService:  meta.NewRegisterService(),
		ReferenceService: meta.NewReferenceService(),
		RouterService:    meta.NewRouterService(),
	}
	if err := mode.Init(); err != nil {
		return nil, err
	}
	return mode, nil
}

package cache

import (
	"github.com/apache/dubbo-go/config"
	_ "github.com/apache/dubbo-go/registry/protocol"
	perrors "github.com/pkg/errors"
	"sync"
)

var drCache dubboReferenceCache
var regCache dubboRegistryCache
var defaultConsumerConfig *config.ConsumerConfig
var referReference = make(map[string]map[string]string)
var referLock sync.RWMutex

func init() {
	drCache = dubboReferenceCache{}
	regCache = dubboRegistryCache{}
	check := true
	defaultConsumerConfig = &config.ConsumerConfig{
		Check:           &check,
		Request_Timeout: "3s",
		Connect_Timeout: "3s",
		ApplicationConfig: &config.ApplicationConfig{
			Name:        "gateway",
			Environment: "prod",
		},
	}
}

type dubboReferenceCache map[string]*config.ReferenceConfig
type dubboRegistryCache map[string]*config.RegistryConfig

func addRegistry(registryId string, registryConfig *config.RegistryConfig) {
	regCache[registryId] = registryConfig
}

func getRegistry(registryId string) *config.RegistryConfig {
	return regCache[registryId]
}

func addReference(referenceId, identify string, referenceConfig *config.ReferenceConfig) error {
	referLock.Lock()
	defer referLock.Unlock()
	if referenceConfig.Registry == "" {
		return perrors.Errorf("miss registry param")
	}
	if registry := getRegistry(referenceConfig.Registry); registry == nil {
		return perrors.Errorf("registry[%s] not found in cache", referenceConfig.Registry)
	}
	if _, ok := drCache[referenceId]; !ok {
		if service := config.GetConsumerService(referenceId); service == nil {
			referenceConfig.GenericLoad(referenceId)
		}
		drCache[referenceId] = referenceConfig
	}
	content, ok := referReference[referenceId]
	if !ok {
		content = make(map[string]string)
	}
	content[identify] = identify
	referReference[referenceId] = content
	return nil
}

func getReference(referenceId string) (*config.ReferenceConfig, error) {
	referLock.RLock()
	defer referLock.RUnlock()
	reference, ok := drCache[referenceId]
	if !ok {
		return nil, perrors.Errorf("reference[%s] not found in cache", referenceId)
	}
	return reference, nil
}

func modifyReference(referenceId string, referenceConfig *config.ReferenceConfig) {
	referLock.Lock()
	defer referLock.Unlock()
	referenceConfig.GenericLoad(referenceId)
	drCache[referenceId] = referenceConfig
}

func removeAllReference(referenceId string) []string {
	referLock.Lock()
	defer referLock.Unlock()
	var identifies []string
	content, ok := referReference[referenceId]
	if ok {
		identifies = make([]string, 0, len(content))
		index := 0
		for key, _ := range content {
			identifies[index] = key
			index++
		}
	}
	delete(drCache, referenceId)
	return identifies
}

func removeReferences(referenceMap map[string][]string) {
	referLock.Lock()
	defer referLock.Unlock()
	for referenceId, identifies := range referenceMap {
		cacheIdentifies, ok := referReference[referenceId]
		if ok {
			for _, key := range identifies {
				delete(cacheIdentifies, key)
			}
		}
		if cacheIdentifies == nil || len(cacheIdentifies) == 0 {
			delete(drCache, referenceId)
		}
	}
}

func RefreshConsumerConfig() {
	defaultConsumerConfig.Registries = regCache
	config.SetConsumerConfig(*defaultConsumerConfig)
}

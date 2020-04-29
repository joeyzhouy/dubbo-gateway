package cache

import (
	"github.com/apache/dubbo-go/common/logger"
	"github.com/apache/dubbo-go/config"
	perrors "github.com/pkg/errors"
	"sync"
)

var drCache dubboReferenceCache
var regCache dubboRegistryCache
var defaultConsumerConfig *config.ConsumerConfig
var referReference map[string]map[string]string
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

func AddRegistry(registryId string, registryConfig *config.RegistryConfig) {
	regCache[registryId] = registryConfig
}

func getRegistry(registryId string) *config.RegistryConfig {
	return regCache[registryId]
}

func AddReference(referenceId, identify string, referenceConfig *config.ReferenceConfig) error {
	referLock.Lock()
	defer referLock.Unlock()
	if referenceConfig.Registry == "" {
		return perrors.Errorf("miss registry param")
	}
	if registry := getRegistry(referenceConfig.Registry); registry == nil {
		return perrors.Errorf("registry[%s] not found in cache", referenceConfig.Registry)
	}
	if _, ok := drCache[referenceId]; !ok {
		referenceConfig.GenericLoad(referenceId)
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

func GetReference(referenceId string) (*config.ReferenceConfig, error) {
	referLock.RLock()
	defer referLock.RUnlock()
	reference, ok := drCache[referenceId]
	if !ok {
		return nil, perrors.Errorf("reference[%s] not found in cache", referenceId)
	}
	return reference, nil
}

func RemoveReference(referenceId, identify string) {
	referLock.Lock()
	defer referLock.Unlock()
	_, ok := drCache[referenceId]
	if !ok {
		logger.Warnf("remove reference[%s] with identify[%s], but not found reference", referenceId, identify)
		return
	}
	content, ok := referReference[referenceId]
	if !ok {
		logger.Warnf("remove reference[%s] with identify[%s], but not found identify", referenceId, identify)
		return
	}
	delete(content, identify)
	if len(content) == 0 {
		// release
		delete(drCache, referenceId)
	}
}

func RefreshConsumerConfig() {
	defaultConsumerConfig.Registries = regCache
	config.SetConsumerConfig(*defaultConsumerConfig)
}

package single

import (
	"dubbo-gateway/common"
	"dubbo-gateway/common/config"
	"dubbo-gateway/common/extension"
	"dubbo-gateway/communication/cache"
	_ "dubbo-gateway/meta/kv/zookeeper"
	_ "dubbo-gateway/meta/relation/mysql"
	"fmt"
	"github.com/apache/dubbo-go/common/logger"
	"sync"
)

const SingleMode = "single"

type singleMode struct {
	common.GatewayCache
	sync.RWMutex
	sMap map[string]map[string]func(mode extension.ModeEvent)
}

func (s *singleMode) UnsubscribeEvent(domain extension.Domain, eventType extension.EventType, identify string) {
	key := getKey(domain, eventType)
	eventMap, ok := s.sMap[key]
	if ok {
		delete(eventMap, identify)
		s.sMap[key] = eventMap
	}
}

func (s *singleMode) SubscribeEvent(domain extension.Domain, eventType extension.EventType,
	identify string, f func(event extension.ModeEvent)) error {
	key := getKey(domain, eventType)
	eventMap, ok := s.sMap[key]
	if !ok {
		eventMap = make(map[string]func(extension.ModeEvent))
	}
	eventMap[identify] = f
	return nil
}

func getKey(domain extension.Domain, eventType extension.EventType) string {
	return fmt.Sprintf("%d-%d", domain, eventType)
}

func (*singleMode) Start() {
	logger.Info("single mode start")
}

func (s *singleMode) Notify(event extension.ModeEvent) {
	go func() {
		var (
			key1 string
			key2 string
		)
		key1 = getKey(event.Domain, event.Type)
		key2 = getKey(event.Domain, extension.AllType)
		eventMap, ok := s.sMap[key1]
		if ok {
			for _, f := range eventMap {
				f(event)
			}
		}
		if key1 != key2 {
			eventMap, ok = s.sMap[key2]
			if ok {
				for _, f := range eventMap {
					f(event)
				}
			}
		}
	}()
}

func (*singleMode) Close() {
	logger.Info("single mode close")
}

func init() {
	extension.SetMode(SingleMode, newSingleMode)
}

func newSingleMode(deploy *config.Deploy) (extension.Mode, error) {
	mode := &singleMode{
		sMap: make(map[string]map[string]func(extension.ModeEvent)),
	}
	gatewayCache, err := cache.NewLocalCache(mode)
	if err != nil {
		return nil, err
	}
	mode.GatewayCache = gatewayCache
	return mode, nil
}

package extension

import (
	"fmt"
	"github.com/apache/dubbo-go/common/logger"
)

const (
	Console       = "console"
	Communication = "communication"
	Router        = "router"
)

var origins = make(map[string]Origin)

func SetOrigin(key string, origin Origin) {
	origins[key] = origin
}

func GetOrigin(key string) Origin {
	return origins[key]
}

func Start() {
	for key, origin := range origins {
		logger.Info(fmt.Sprintf("start %s...", key))
		origin.Start()
	}
}

func Close() {
	for key, origin := range origins {
		logger.Info(fmt.Sprintf("close %s...", key))
		origin.Close()
	}
}

type Origin interface {
	Start()
	Close()
}

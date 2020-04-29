package redis

import "dubbo-gateway/common/extension"

type redisMode struct {

}

func (*redisMode) Start() {
	panic("implement me")
}

func (*redisMode) Notify(event extension.ModeEvent) {
	panic("implement me")
}

func (*redisMode) Close() {
	panic("implement me")
}


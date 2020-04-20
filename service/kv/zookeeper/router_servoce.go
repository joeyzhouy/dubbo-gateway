package zookeeper

import (
	"dubbo-gateway/service"
	"dubbo-gateway/service/entry"
	"dubbo-gateway/service/vo"
	"github.com/dubbogo/go-zookeeper/zk"
)

type routerService struct {
	conn *zk.Conn
}

func NewRouterService(conn *zk.Conn) service.RouterService {
	return &routerService{conn}
}

func (*routerService) AddRouter(api *entry.ApiConfig) error {
	panic("implement me")
}

func (*routerService) AddApiConfig(api *vo.ApiConfigInfo) error {
	panic("implement me")
}

func (*routerService) DeleteRouter(apiId int64) error {
	panic("implement me")
}

func (*routerService) ListRouterByUserId(userId int64) ([]entry.ApiConfig, error) {
	panic("implement me")
}

func (*routerService) ListAll() ([]*vo.ApiConfigInfo, error) {
	panic("implement me")
}

func (*routerService) GetByApiId(api int64) (*vo.ApiConfigInfo, error) {
	panic("implement me")
}

func (*routerService) GetByUri(uri string) (*entry.ApiConfig, error) {
	panic("implement me")
}

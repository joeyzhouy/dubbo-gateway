package zookeeper

import (
	"dubbo-gateway/service"
	"dubbo-gateway/service/entry"
	"github.com/dubbogo/go-zookeeper/zk"
)

type commonService struct {
	conn *zk.Conn
}

func NewCommonService(conn *zk.Conn) service.CommonService {
	return &commonService{conn}
}

func (*commonService) GetUser(userName, password string) (*entry.User, error) {
	panic("implement me")
}

func (*commonService) CreateUser(user *entry.User) error {
	panic("implement me")
}

func (*commonService) UpdatePassword(user *entry.User, oldPassword string) error {
	panic("implement me")
}

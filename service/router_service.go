package service

import (
	"dubbo-gateway/meta"
	"dubbo-gateway/service/entry"
	"github.com/jinzhu/gorm"
)

type RouterService interface {
	AddRouter(api *entry.ApiConfig) error
	DeleteRouter(apiId int64) error
	UpdateRouter(api *entry.ApiConfig) error
	ListRouterByUserId(userId int64) ([]entry.ApiConfig, error)
}

type routerService struct {
	*gorm.DB
}

func (r *routerService) AddRouter(api *entry.ApiConfig) error {
	panic("implement me")
}

func (r *routerService) DeleteRouter(apiId int64) error {
	panic("implement me")
}

func (r *routerService) UpdateRouter(api *entry.ApiConfig) error {
	panic("implement me")
}

func (r *routerService) ListRouterByUserId(userId int64) ([]entry.ApiConfig, error) {
	panic("implement me")
}

func NewRouterService() RouterService {
	return &routerService{meta.GetDB()}
}

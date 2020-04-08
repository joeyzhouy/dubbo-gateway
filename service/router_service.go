package service

import (
	"crypto/md5"
	"dubbo-gateway/meta"
	"dubbo-gateway/service/entry"
	"encoding/hex"
	"github.com/jinzhu/gorm"
)

type RouterService interface {
	AddRouter(api *entry.ApiConfig) error
	DeleteRouter(apiId int64) error
	ListRouterByUserId(userId int64) ([]entry.ApiConfig, error)
}

type routerService struct {
	*gorm.DB
}

func (r *routerService) AddRouter(api *entry.ApiConfig) error {
	api.UriHash = hash(api.Uri)
	return r.Save(api).Error
}

func (r *routerService) DeleteRouter(apiId int64) error {
	return r.Model(&entry.ApiConfig{}).Where("id = ?").UpdateColumn("is_delete", 1).Error
}

func (r *routerService) ListRouterByUserId(userId int64) ([]entry.ApiConfig, error) {
	result := make([]entry.ApiConfig, 0)
	err := r.Where("user_id = ? AND is_delete = 0", userId).Find(&result).Error
	return result, err
}

func NewRouterService() RouterService {
	return &routerService{meta.GetDB()}
}

func hash(str string) string {
	h := md5.New()
	h.Write([]byte(str))
	return hex.EncodeToString(h.Sum(nil))
}

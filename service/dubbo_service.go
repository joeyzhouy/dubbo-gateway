package service

import (
	"dubbo-gateway/meta"
	"dubbo-gateway/service/entry"
	"errors"
	"github.com/jinzhu/gorm"
)

var NoRight = errors.New("no right")

type RegisterService interface {
	AddRegistryConfig(config entry.Registry) error
	DeleteRegistryConfig(registryId, userId int64) error
	ListRegistryByUser(userId int64) ([]entry.Registry, error)
}

func NewDubboService() RegisterService {
	return &registryService{meta.GetDB()}
}

type registryService struct {
	*gorm.DB
}

func (d *registryService) AddRegistryConfig(config entry.Registry) error {
	return d.Save(config).Error
}

func (d *registryService) DeleteRegistryConfig(registryId, userId int64) error {
	dbRegistry := new(entry.Registry)
	err := d.Where("user_id = ? and id = ?", userId, registryId).
		Find(dbRegistry).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return NoRight
		}
		return err
	}
	return d.Where("id = ?", registryId).UpdateColumn("is_delete", 1).Error
}

func (d *registryService) ListRegistryByUser(userId int64) ([]entry.Registry, error) {
	result := make([]entry.Registry, 0)
	err := d.Where("user_id = ?", userId).Find(&result).Error
	return result, err
}

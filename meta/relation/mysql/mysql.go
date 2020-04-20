package mysql

import (
	"dubbo-gateway/common/extension"
	"dubbo-gateway/meta"
	"dubbo-gateway/meta/relation"
	"dubbo-gateway/service"
	relationService "dubbo-gateway/service/relation"
	"github.com/jinzhu/gorm"
	"sync"
)

const (
	mysql = "mysql"
)

func init() {
	extension.SetMeta(mysql, NewMetaMysql)
}

var mysqlMeta *metaMysql
var initError error
var once sync.Once

type metaMysql struct {
	db *gorm.DB
}

func (m *metaMysql) NewCommonService() service.CommonService {
	return relationService.NewCommonService(m.db)
}

func (m *metaMysql) NewRouterService() service.RouterService {
	return relationService.NewRouterService(m.db)
}

func (m *metaMysql) NewReferenceService() service.ReferenceService {
	return relationService.NewReferenceService(m.db)
}

func (m *metaMysql) NewRegisterService() service.RegisterService {
	return relationService.NewRegistryService(m.db)
}

func (m *metaMysql) NewMethodService() service.MethodService {
	return relationService.NewMethodService(m.db)
}

func NewMetaMysql(configString string) (meta.Meta, error) {
	once.Do(func() {
		mysqlMeta = new(metaMysql)
		mysqlMeta.db, initError = relation.InitRelationConfig(mysql, configString)
	})
	//TODO init mysql tables
	return mysqlMeta, initError
}

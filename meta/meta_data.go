package meta

import "dubbo-gateway/service"

type Meta interface {
	NewCommonService() service.CommonService
	NewRouterService() service.RouterService
	NewReferenceService() service.ReferenceService
	NewRegisterService() service.RegisterService
	NewMethodService() service.MethodService
	NewEntryService() service.EntryService
}

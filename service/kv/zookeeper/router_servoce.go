package zookeeper

import (
	"dubbo-gateway/common/constant"
	"dubbo-gateway/common/utils"
	"dubbo-gateway/service"
	"dubbo-gateway/service/entry"
	"dubbo-gateway/service/vo"
	"encoding/json"
	"fmt"
	"github.com/dubbogo/go-zookeeper/zk"
)

type routerService struct {
	conn *zk.Conn
}

func NewRouterService(conn *zk.Conn, event <-chan zk.Event) service.RouterService {
	return &routerService{conn}
}

func (r *routerService) AddRouter(api *entry.ApiConfig) error {
	id, err := next()
	if err != nil {
		return err
	}
	api.ID = id
	nodePath := fmt.Sprintf(constant.ApiInfoPath, api.UserId, id)
	bs, _, err := r.conn.Get(nodePath)
	if err != nil {
		return err
	}
	_, err = r.conn.Multi(&zk.CreateRequest{
		Path:  nodePath,
		Data:  bs,
		Flags: 0,
		Acl:   zk.WorldACL(zk.PermAll),
	}, &zk.CreateRequest{
		Path:  fmt.Sprintf(constant.ApiSearchInfo, id),
		Data:  []byte(nodePath),
		Flags: 0,
		Acl:   zk.WorldACL(zk.PermAll),
	})
	return err
}

func (r *routerService) AddApiConfig(api *vo.ApiConfigInfo) error {
	api.ApiConfig.UriHash = utils.Hash(api.ApiConfig.Uri)
	requests := make([]interface{}, 0)
	filter := api.Filter
	api.Filter = entry.ApiFilter{}
	ids, err := nextN(2)
	filter.ID = ids[0]
	api.ApiConfig.ID = ids[1]
	api.ApiConfig.FilterId = filter.ID
	bs, err := json.Marshal(&filter)
	if err != nil {
		return err
	}
	filterPath := fmt.Sprintf(constant.FilterInfoPath, filter.ID)
	requests = append(requests, &zk.CreateRequest{
		Path:  filterPath,
		Data:  bs,
		Flags: 0,
		Acl:   zk.WorldACL(zk.PermAll),
	})
	bs, err = json.Marshal(api)
	if err != nil {
		return nil
	}
	userAPiPath := fmt.Sprintf(constant.ApiInfoPath, api.ApiConfig.UserId, api.ApiConfig.ID)
	requests = append(requests, &zk.CreateRequest{
		Path:  userAPiPath,
		Data:  bs,
		Flags: 0,
		Acl:   zk.WorldACL(zk.PermAll),
	}, &zk.CreateRequest{
		Path:  fmt.Sprintf(constant.ApiSearchInfo, api.ApiConfig.ID),
		Data:  []byte(userAPiPath),
		Flags: 0,
		Acl:   zk.WorldACL(zk.PermAll),
	})
	_, err = r.conn.Multi(requests...)
	return err

}

func (r *routerService) DeleteRouter(apiId int64) error {
	apiInfoPath, _, err := r.conn.Get(fmt.Sprintf(constant.ApiSearchInfo, apiId))
	if err != nil {
		return err
	}
	_, err = r.conn.Multi(&zk.DeleteRequest{
		Path:    string(apiInfoPath),
		Version: -1,
	}, &zk.DeleteRequest{
		Path:    fmt.Sprintf(constant.ApiSearchInfo, apiId),
		Version: -1,
	})
	return err
}

func (r *routerService) ListRouterByUserId(userId int64) ([]entry.ApiConfig, error) {
	apiBasePath := fmt.Sprintf(constant.ApiPath, userId)
	children, _, err := r.conn.Children(apiBasePath)
	if err != nil {
		return nil, err
	}
	var bs []byte
	result := make([]entry.ApiConfig, 0, len(children))
	for _, child := range children {
		bs, _, err = r.conn.Get(child)
		if err != nil {
			return nil, err
		}
		config := new(vo.ApiConfigInfo)
		err = json.Unmarshal(bs, config)
		if err != nil {
			return nil, err
		}
		result = append(result, config.ApiConfig)
	}
	return result, nil
}

func (r *routerService) getAllFilter() ([]entry.ApiFilter, error) {
	children, _, err := r.conn.Children(constant.FilterPath)
	if err != nil {
		return nil, err
	}
	result := make([]entry.ApiFilter, 0, len(children))
	var bs []byte
	for _, child := range children {
		bs, _, err = r.conn.Get(child)
		if err != nil {
			return nil, err
		}
		filter := new(entry.ApiFilter)
		err = json.Unmarshal(bs, filter)
		if err != nil {
			return nil, err
		}
		result = append(result, *filter)
	}
	return result, nil
}

func (r *routerService) getFilterById(filterId int64) (*entry.ApiFilter, error) {
	filterInfoPath := fmt.Sprintf(constant.FilterInfoPath, filterId)
	bs, _, err := r.conn.Get(filterInfoPath)
	if err != nil {
		return nil, err
	}
	result := new(entry.ApiFilter)
	err = json.Unmarshal(bs, result)
	return result, nil
}

func (r *routerService) ListAll() ([]*vo.ApiConfigInfo, error) {
	children, _, err := r.conn.Children(constant.ApiSearchRoot)
	if err != nil {
		return nil, err
	}
	filters, err := r.getAllFilter()
	if err != nil {
		return nil, err
	}
	filterMap := make(map[int64]entry.ApiFilter)
	for _, filter := range filters {
		filterMap[filter.ID] = filter
	}
	var bs []byte
	result := make([]*vo.ApiConfigInfo, 0)
	for _, child := range children {
		bs, _, err = r.conn.Get(child)
		if err != nil {
			return nil, err
		}
		configInfo := new(vo.ApiConfigInfo)
		err = json.Unmarshal(bs, configInfo)
		if err != nil {
			return nil, err
		}
		if filter, ok := filterMap[configInfo.ApiConfig.FilterId]; ok {
			configInfo.Filter = filter
		}
		result = append(result, configInfo)
	}
	return result, nil
}

func (r *routerService) GetByApiId(api int64) (*vo.ApiConfigInfo, error) {
	configSearchPath := fmt.Sprintf(constant.ApiSearchInfo, api)
	bs, _, err := r.conn.Get(configSearchPath)
	if err != nil {
		return nil, err
	}
	bs, _, err = r.conn.Get(string(bs))
	if err != nil {
		return nil, err
	}
	result := new(vo.ApiConfigInfo)
	err = json.Unmarshal(bs, result)
	if err != nil {
		return nil, err
	}
	if result.ApiConfig.FilterId > 0 {
		filter, err := r.getFilterById(result.ApiConfig.FilterId)
		if err != nil {
			return nil, err
		}
		result.Filter = *filter
	}
	return result, nil
}

//func (r *routerService) GetByUri(uri string) (*entry.ApiConfig, error) {
//	panic("implement me")
//}

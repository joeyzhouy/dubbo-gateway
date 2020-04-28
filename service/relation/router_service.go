package relation

import (
	"dubbo-gateway/common/utils"
	"dubbo-gateway/service"
	"dubbo-gateway/service/entry"
	"dubbo-gateway/service/vo"
	"fmt"
	"github.com/apache/dubbo-go/common/logger"
	"github.com/jinzhu/gorm"
	"strings"
)

type routerService struct {
	*gorm.DB
	service.MethodService
}

func (r *routerService) GetByApiId(api int64) (*vo.ApiConfigInfo, error) {
	apiConfig := new(entry.ApiConfig)
	if err := r.Where("id = ?", api).Find(&apiConfig).Error; err != nil {
		return nil, err
	}
	apiFilter := new(entry.ApiFilter)
	if err := r.Where("api_id = ? and is_delete = 0", api).Find(apiFilter).Error; err != nil {
		return nil, err
	}
	apiChains := make([]entry.ApiChain, 0)
	if err := r.Where("api_id = ? and is_delete = 0", api).Order("seq").Find(&apiChains).Error; err != nil {
		return nil, err
	}
	apiResultRules := make([]entry.ApiResultRule, 0)
	if err := r.Where("api_id = ? and is_delete = 0", api).Order("chain_id, seq").Find(&apiResultRules).Error; err != nil {
		return nil, err
	}
	return r.join([]entry.ApiConfig{*apiConfig}, []entry.ApiFilter{*apiFilter},
		apiChains, apiResultRules)[0], nil
}

func (r *routerService) join(apiConfigs []entry.ApiConfig, apiFilters []entry.ApiFilter,
	apiChains []entry.ApiChain, apiResultRules []entry.ApiResultRule) []*vo.ApiConfigInfo {
	rules := make(map[int64][]entry.ApiResultRule)
	methodIds := make([]int64, 0)
	for _, rule := range apiResultRules {
		temp, ok := rules[rule.ChainId]
		if !ok {
			temp = make([]entry.ApiResultRule, 0)
		}
		temp = append(temp, rule)
		rules[rule.ChainId] = temp
	}
	chains := make(map[int64][]vo.ApiChainInfo)
	for _,chain := range apiChains {
		methodIds = append(methodIds, chain.MethodId)
	}
	filterMap := make(map[int64]entry.ApiFilter)
	for _, filter := range apiFilters {
		filterMap[filter.ID] = filter
		methodIds = append(methodIds, filter.MethodId)
	}
	methods, err := r.MethodService.GetMethodDetailByIds(methodIds)
	if err != nil {
		logger.Errorf("find method info error")
		return nil
	}
	methodMap := make(map[int64]*vo.Method)
	for _, method := range methods {
		methodMap[method.ID] = method
	}
	for _, chain := range apiChains {
		temp, ok := chains[chain.ApiId]
		if !ok {
			temp = make([]vo.ApiChainInfo, 0)
		}
		temp = append(temp, vo.ApiChainInfo{Chain: chain, Rules: rules[chain.ID], Method: *methodMap[chain.MethodId]})
		chains[chain.ApiId] = temp
	}
	result := make([]*vo.ApiConfigInfo, 0, len(apiConfigs))
	for _, config := range apiConfigs {
		filter := filterMap[config.FilterId]
		configInfo := new(vo.ApiConfigInfo)
		configInfo.ApiConfig = config
		configInfo.Filter = vo.ApiFilter{ApiFilter: filter, Method: *methodMap[filter.MethodId]}
		configInfo.Chains = chains[config.ID]
		result = append(result, configInfo)
	}
	return result
}

func (r *routerService) ListAll() ([]*vo.ApiConfigInfo, error) {
	apiConfigs := make([]entry.ApiConfig, 0)
	if err := r.Where("is_delete = 0").Find(&apiConfigs).Error; err != nil {
		return nil, err
	}
	apiFilters := make([]entry.ApiFilter, 0)
	if err := r.Where("is_delete = 0").Find(&apiFilters).Error; err != nil {
		return nil, err
	}
	apiChains := make([]entry.ApiChain, 0)
	if err := r.Where("is_delete = 0").Order("app_id, seq").Find(&apiChains).Error; err != nil {
		return nil, err
	}
	apiResultRules := make([]entry.ApiResultRule, 0)
	if err := r.Where("is_delete = 0").Order("chain_id, seq").Find(&apiResultRules).Error; err != nil {
		return nil, err
	}
	return r.join(apiConfigs, apiFilters, apiChains, apiResultRules), nil
}

func (r *routerService) AddApiConfig(api *vo.ApiConfigInfo) error {
	api.ApiConfig.UriHash = utils.Hash(api.ApiConfig.Uri)
	tx := r.Begin()
	if err := tx.Save(&(api.ApiConfig)).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Save(&(api.Filter)).Error; err != nil {
		tx.Rollback()
		return err
	}
	for _, c := range api.Chains {
		chain := c.Chain
		if err := tx.Save(&chain).Error; err != nil {
			tx.Rollback()
			return err
		}
		length := len(c.Rules)
		if length == 0 {
			continue
		}
		valueStrings := make([]string, 0, length)
		valueArgs := make([]interface{}, 0, length)
		for _, rule := range c.Rules {
			valueStrings = append(valueStrings, fmt.Sprintf("(%d, %d, '%s', %d, '%s')", api.ApiConfig.ID,
				chain.ID, rule.JavaClass, rule.Index, rule.Rule))
		}
		smt := "INSERT INTO d_api_result_rule(api_id, chain_id, java_class, `index`, rule) VALUES " + strings.Join(valueStrings, ",")
		if err := tx.Exec(smt, valueArgs...).Error; err != nil {
			tx.Rollback()
			return err
		}
	}
	tx.Commit()
	return nil
}

func (r *routerService) AddRouter(api *entry.ApiConfig) error {
	api.UriHash = utils.Hash(api.Uri)
	return r.Save(api).Error
}

func (r *routerService) DeleteRouter(apiId int64) error {
	tx := r.Begin()
	if err := tx.Model(&entry.ApiResultRule{}).Where("api_id = ?", apiId).UpdateColumn("is_delete", 1).Error;
		err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Model(&entry.ApiChain{}).Where("api_id = ?", apiId).UpdateColumn("is_delete", 1).Error;
		err != nil {
		tx.Rollback()
		return err
	}
	if err := r.Model(&entry.ApiConfig{}).Where("id = ?").UpdateColumn("is_delete", 1).Error; err != nil {
		tx.Rollback()
		return err
	}
	tx.Commit()
	return nil
}

func (r *routerService) ListRouterByUserId(userId int64) ([]entry.ApiConfig, error) {
	result := make([]entry.ApiConfig, 0)
	err := r.Where("user_id = ? AND is_delete = 0", userId).Find(&result).Error
	return result, err
}

func NewRouterService(db *gorm.DB) service.RouterService {
	return &routerService{db, NewMethodService(db)}
}

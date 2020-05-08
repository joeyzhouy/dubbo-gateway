package relation

import (
	"dubbo-gateway/common/extension"
	"dubbo-gateway/service"
	"dubbo-gateway/service/entry"
	"dubbo-gateway/service/vo"
	"fmt"
	"github.com/apache/dubbo-go/common/logger"
	"github.com/go-errors/errors"
	"github.com/jinzhu/gorm"
	"strings"
)

type routerService struct {
	*gorm.DB
	service.MethodService
}

func (r *routerService) GetFilter(filterId int64) (*vo.ApiFilterInfo, error) {
	result := new(vo.ApiFilterInfo)
	err := r.Where("id = ?", filterId).Find(&(result.Filter)).Error
	if err != nil {
		return nil, err
	}
	var mappings []entry.ApiParamMapping
	if err = r.Where("chain_id = ? and api_id is null", filterId).Find(&mappings).Error; err != nil {
		return nil, err
	}
	if len(mappings) > 0 {
		voMappings := make([]*vo.ApiParamMapping, 0, len(mappings))
		for index, mapping := range mappings {
			voMappings[index] = &vo.ApiParamMapping{
				ApiParamMapping: mapping,
			}
		}
	}
	if err = result.Unmarshal(); err != nil {
		return nil, err
	}
	return result, err
}

func (r *routerService) AddFilter(filter *vo.ApiFilterInfo) (err error) {
	tx := r.Begin()
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()
	if err = tx.Save(&filter.Filter).Error; err != nil {
		return
	}
	if len(filter.ParamMappings) > 0 {
		return r.saveFilterMapping(tx, filter.ParamMappings, filter.Filter.ID)
	}
	return nil
}

func (r *routerService) saveFilterMapping(tx *gorm.DB, paramMappings []*vo.ApiParamMapping, filterId int64) error {
	length := len(paramMappings)
	valueStrings := make([]string, 0, length)
	valueArgs := make([]interface{}, 0, length)
	var err error
	for _, mapping := range paramMappings {
		if err = mapping.Marshall(); err != nil {
			return err
		}
		valueStrings = append(valueStrings, fmt.Sprintf("(%d, %d, %d,'%s')", filterId, entry.ParamMapping, mapping.Index, mapping.Explain))
	}
	smt := "INSERT INTO d_api_param_mapping(chain_id, type_id, index, explain) VALUES " + strings.Join(valueStrings, ",")
	return tx.Exec(smt, valueArgs...).Error
}

func (r *routerService) ModifyFilter(filter *vo.ApiFilterInfo) (err error) {
	tx := r.Begin()
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()
	if err = tx.Save(&filter.Filter).Error; err != nil {
		return
	}
	if err = tx.Delete(entry.ApiParamMapping{}, "chain_id = ? and api_id is null",
		filter.Filter.ID).Error; err != nil {
		return
	}
	if len(filter.ParamMappings) > 0 {
		return r.saveFilterMapping(tx, filter.ParamMappings, filter.Filter.ID)
	}
	return nil
}

func (r *routerService) DeleterFilter(filterId int64) (err error) {
	var count int
	if err = r.Table("d_api_config").Where("filter_id = ?",
		filterId).Count(&count).Error; err != nil {
		return
	}
	tx := r.Begin()
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()
	if err = tx.Delete(entry.ApiParamMapping{}, "chain_id = ? and api_id is null",
		filterId).Error; err != nil {
		return
	}
	return tx.Delete(entry.ApiFilter{}, "id = ?", filterId).Error
}

func (r *routerService) ListFilters() ([]entry.ApiFilter, error) {
	var filters []entry.ApiFilter
	if err := r.Find(&filters).Error; err != nil && err == gorm.ErrRecordNotFound {
		return nil, err
	}
	return filters, nil
}

func (r *routerService) SearchByMethodName(methodName string) ([]entry.ApiConfig, error) {
	var result []entry.ApiConfig
	db := r.Where("is_delete = 0 and status = ?", entry.Available)
	methodNameLike := strings.TrimSpace(methodName)
	if methodNameLike != "" {
		db = db.Where("method_name Like ?", "%"+methodNameLike+"%")
	}
	err := db.Find(&result).Error
	return result, err
}

func (r *routerService) ListAllAvailableEntry() ([]*entry.ApiConfig, error) {
	var result []*entry.ApiConfig
	err := r.Where("is_delete = 0 and status = ?", entry.Available).Find(&result).Error
	return result, err
}

func (r *routerService) GetApiIdsByMethodId(methodId int64) ([]int64, error) {
	var result []int64
	err := r.Table("d_api_config").
		Select("distinct(d_api_config.id)").
		Joins("JOIN d_api_chain on d_api_chain.api_id = d_api_config.id and d_api_chain.is_delete = 0").
		Where("d_api_chain.method_id = ?", methodId).Scan(&result).Error
	if err != nil && err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return result, err
}

func (r *routerService) GetConfigById(configId int64) (*entry.ApiConfig, error) {
	result := new(entry.ApiConfig)
	err := r.Where("id = ?", configId).Find(result).Error
	return result, err
}

func (r *routerService) GetApiMethodNamesByReferenceId(referenceId int64) ([]string, error) {
	var result []string
	err := r.Table("d_api_config").
		Select("distinct(d_api_config.method)").
		Joins("JOIN d_api_chain on d_api_chain.api_id = d_api_config.id").
		Where("d_api_chain.reference_id = ?", referenceId).Scan(&result).Error
	if err != nil && err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return result, err
}

func (r *routerService) ModifyConfigStatus(configId int64, status int) (err error) {
	defer func() {
		if err == nil && status == entry.Available {
			extension.GetConfigMode().Notify(extension.ModeEvent{
				Domain: extension.Api,
				Type:   extension.Add,
				Key:    configId,
			})
		}
	}()
	return r.Where("id = ?", configId).UpdateColumn("status", status).Error
}

func (r *routerService) UpdateConfig(apiConfig *vo.ApiConfigInfo) (err error) {
	tx := r.Begin()
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			apiConfigDB, err := r.GetConfigById(apiConfig.ApiConfig.ID)
			if err != nil {
				logger.Errorf("UpdateConfig.GetConfigById[%d], error: %v", apiConfig.ApiConfig.ID, err)
				return
			}
			if apiConfigDB.Status == entry.Available {
				extension.GetConfigMode().Notify(extension.ModeEvent{
					Domain: extension.Api,
					Type:   extension.Modify,
					Key:    apiConfig.ApiConfig.ID,
				})
			}
		}
	}()
	if err = r.deleteConfigRelations(tx, apiConfig.ApiConfig.ID); err != nil {
		return
	}
	return r.saveChains(tx, apiConfig)
}

func (r *routerService) GetByConfigId(configId int64) (*vo.ApiConfigInfo, error) {
	var (
		err           error
		apiConfig     entry.ApiConfig
		chains        []entry.ApiChain
		mappings      []entry.ApiParamMapping
		filter        entry.ApiFilter
		filterMapping []entry.ApiParamMapping
	)
	if err = r.Where("id = ?", configId).Find(&apiConfig).Error; err != nil {
		return nil, err
	}
	if err = r.Where("api_id = ?", configId).Find(&chains).Error; err != nil {
		return nil, err
	}
	if err = r.Where("api_id = ?", configId).Find(&mappings).Error; err != nil {
		return nil, err
	}
	if apiConfig.FilterId != 0 {
		if err = r.Where("id = ?", apiConfig.ID).Find(&filter).Error; err != nil && err != gorm.ErrRecordNotFound {
			return nil, err
		} else if err = r.Where("chain_id = ? and api_id is null", filter.ID).Error; err != nil && err != gorm.ErrRecordNotFound {
			return nil, err
		}
	}
	infos, err := r.join([]entry.ApiConfig{apiConfig}, chains, mappings, []entry.ApiFilter{filter}, filterMapping)
	if err != nil {
		return nil, err
	}
	return infos[0], err
}

func (r *routerService) ListAllAvailable() ([]*vo.ApiConfigInfo, error) {
	var (
		apiConfigs     []entry.ApiConfig
		chains         []entry.ApiChain
		mappings       []entry.ApiParamMapping
		filters        []entry.ApiFilter
		filterMappings []entry.ApiParamMapping
	)
	if err := r.Where("status = ?", entry.Available).Find(&apiConfigs).Error; err != nil {
		return nil, err
	}
	if err := r.Table("d_api_chain").
		Select("d_api_chain.id, d_api_chain.api_id, d_api_chain.type_id, d_api_chain.reference_id, d_api_chain.method_id, d_api_chain.seq").
		Joins("JOIN d_api_config on d_api_config.id = d_api_chain.api_id").
		Where("d_api_config.status = ?", entry.Available).Order("d_api_chain.id").Find(&chains).Error; err != nil {
		return nil, err
	}
	if err := r.Table("d_api_param_mapping").
		Select("d_api_param_mapping.id, d_api_param_mapping.api_id, d_api_param_mapping.chain_id, d_api_param_mapping.type_id, d_api_param_mapping.explain").
		Joins("JOIN d_api_config on d_api_config.id = d_api_param_mapping.api_id").
		Where("d_api_config.status = ?", entry.Available).Order("d_api_param_mapping.index").Find(&mappings).Error; err != nil {
		return nil, err
	}
	if err := r.Table("d_api_filter").
		Select("d_api_filter.name, d_api_filter.reference_id, d_api_filter.method_id").
		Joins("JOIN d_api_config on d_api_config.filter_id = d_api_filter.id").
		Where("d_api_config.status = ?", entry.Available).Find(&filters).Error; err != nil {
		return nil, err

	}
	if err := r.Table("d_api_param_mapping").
		Select("d_api_param_mapping.id, d_api_param_mapping.api_id, d_api_param_mapping.chain_id, d_api_param_mapping.type_id, d_api_param_mapping.explain").
		Joins("JOIN d_api_filter on d_api_filter.id = d_api_param_mapping.chain_id and d_api_param_mapping.api_id is null").
		Joins("JOIN d_api_config on d_api_config.filter_id = d_api_filter.id").
		Where("d_api_config.status = ?", entry.Available).Order("d_api_param_mapping.index").Find(&filterMappings).Error; err != nil {
		return nil, err
	}
	return r.join(apiConfigs, chains, mappings, filters, filterMappings)
}

func (r *routerService) join(apiConfigs []entry.ApiConfig, chains []entry.ApiChain,
	mappings []entry.ApiParamMapping, filters []entry.ApiFilter, filterMapping []entry.ApiParamMapping) ([]*vo.ApiConfigInfo, error) {
	var (
		voMapping    *vo.ApiParamMapping
		voChain      *vo.ApiChainInfo
		tempMappings []*vo.ApiParamMapping
		tempChains   []*vo.ApiChainInfo
		err          error
		ok           bool
	)
	apiInfos := make([]*vo.ApiConfigInfo, len(apiConfigs))
	chainMap := make(map[int64][]*vo.ApiChainInfo)
	paramMaps := make(map[int64][]*vo.ApiParamMapping)
	filterMap := make(map[int64]*vo.ApiFilterInfo)
	filterParamMapping := make(map[int64][]*vo.ApiParamMapping)
	for _, param := range mappings {
		tempMappings, ok = paramMaps[param.ChainId]
		if !ok {
			tempMappings = make([]*vo.ApiParamMapping, 0)
		}
		voMapping = &vo.ApiParamMapping{
			ApiParamMapping: param,
		}
		if err = voMapping.Unmarshal(); err != nil {
			return nil, err
		}
		tempMappings = append(tempMappings, voMapping)
		paramMaps[param.ApiId] = tempMappings
	}
	for _, chain := range chains {
		tempChains, ok = chainMap[chain.ID]
		if !ok {
			tempChains = make([]*vo.ApiChainInfo, 0)
		}
		voChain = &vo.ApiChainInfo{
			Chain: chain,
		}
		tempMappings, ok = paramMaps[chain.ID]
		if ok {
			voChain.ParamMappings = make([]*vo.ApiParamMapping, 0)
			for _, mapping := range tempMappings {
				if mapping.TypeId == entry.ParamMapping {
					voChain.ParamMappings = append(voChain.ParamMappings, mapping)
				} else if mapping.TypeId == entry.ResultMapping {
					voChain.ResultMapping = mapping
				}
			}
		}
		tempChains = append(tempChains, voChain)
		chainMap[chain.ID] = tempChains
	}
	for _, mapping := range filterMapping {
		tempMappings, ok = filterParamMapping[mapping.ChainId]
		if !ok {
			tempMappings = make([]*vo.ApiParamMapping, 0)
		}
		voMapping = &vo.ApiParamMapping{
			ApiParamMapping: mapping,
		}
		if err = voMapping.Unmarshal(); err != nil {
			return nil, err
		}
		tempMappings = append(tempMappings, voMapping)
		filterParamMapping[mapping.ChainId] = tempMappings
	}
	for _, filter := range filters {
		filterMap[filter.ID] = &vo.ApiFilterInfo{
			Filter:        filter,
			ParamMappings: filterParamMapping[filter.ID],
		}
	}
	for index, apiConfig := range apiConfigs {
		apiInfos[index] = &vo.ApiConfigInfo{
			ApiConfig: apiConfig,
			Filter:    filterMap[apiConfig.FilterId],
			Chains:    chainMap[apiConfig.ID],
		}
	}
	return apiInfos, nil
}

func (r *routerService) AddConfig(apiConfig *vo.ApiConfigInfo) (err error) {
	tx := r.Begin()
	defer func() {
		if err != nil {
			tx.Rollback()
		} else if apiConfig.ApiConfig.Status == entry.Available {
			extension.GetConfigMode().Notify(extension.ModeEvent{
				Domain: extension.Api,
				Type:   extension.Add,
				Key:    apiConfig.ApiConfig.ID,
			})
		}
	}()
	if err = tx.Save(&(apiConfig.ApiConfig)).Error; err != nil {
		return
	}
	return r.saveChains(tx, apiConfig)
}

func (r *routerService) saveChains(tx *gorm.DB, apiConfig *vo.ApiConfigInfo) (err error) {
	if apiConfig.ApiConfig.ID == 0 {
		return errors.New("miss api configId")
	}
	chainsLength := len(apiConfig.Chains)
	if chainsLength == 0 {
		return nil
	}
	var (
		valueStrings []string
		valueArgs    []interface{}
		isRollBack   bool
	)
	if tx == nil {
		tx = r.Begin()
		isRollBack = true
	}
	defer func() {
		if isRollBack && err != nil {
			tx.Rollback()
		}
	}()
	for _, chain := range apiConfig.Chains {
		if err = tx.Save(&(chain.Chain)).Error; err != nil {
			return
		}
		if err = chain.Marshal(); err != nil {
			return
		}
		valueStrings = make([]string, 0)
		valueArgs = make([]interface{}, 0)
		for _, mapping := range chain.ParamMappings {
			valueStrings = append(valueStrings, fmt.Sprintf("(%d, %d, %d, %d,'%s')", apiConfig.ApiConfig.ID,
				chain.Chain.ID, entry.ParamMapping, mapping.Index, mapping.Explain))
		}
		if chain.ResultMapping != nil {
			valueStrings = append(valueStrings, fmt.Sprintf("(%d, %d, %d, %d,'%s')", apiConfig.ApiConfig.ID,
				chain.Chain.ID, entry.ResultMapping, chain.ResultMapping.Index, chain.ResultMapping.Explain))
		}
		smt := "INSERT INTO d_api_param_mapping(api_id, chain_id, type_id, index, explain) VALUES " + strings.Join(valueStrings, ",")
		if err = tx.Exec(smt, valueArgs...).Error; err != nil {
			return
		}
	}
	return
}

func (r *routerService) DeleteConfig(configId int64) (err error) {
	tx := r.Begin()
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			extension.GetConfigMode().Notify(extension.ModeEvent{
				Domain: extension.Api,
				Type:   extension.Delete,
				Key:    configId,
			})
		}
	}()
	if err = tx.Delete(entry.ApiParamMapping{}, "api_id = ?", configId).Error; err != nil {
		return
	}
	return r.deleteConfigRelations(tx, configId)
}

func (r *routerService) deleteConfigRelations(tx *gorm.DB, configId int64) (err error) {
	needRollBack := false
	if tx == nil {
		tx = r.Begin()
		needRollBack = true
	}
	defer func() {
		if needRollBack && err != nil {
			tx.Rollback()
		}
	}()
	if err = tx.Delete(entry.ApiChain{}, "api_id = ?", configId).Error; err != nil {
		return
	}
	return tx.Delete(entry.ApiConfig{}, "id = ?", configId).Error
}

func NewRouterService(db *gorm.DB) service.RouterService {
	return &routerService{db, NewMethodService(db)}
}

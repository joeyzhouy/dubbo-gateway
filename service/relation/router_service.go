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
		apiConfig entry.ApiConfig
		chains    []entry.ApiChain
		mappings  []entry.ApiParamMapping
	)
	if err := r.Where("id = ?", configId).Find(&apiConfig).Error; err != nil {
		return nil, err
	}
	if err := r.Where("api_id = ?", configId).Find(&chains).Error; err != nil {
		return nil, err
	}
	if err := r.Where("api_id = ?", configId).Find(&mappings).Error; err != nil {
		return nil, err
	}
	infos, err := r.join([]entry.ApiConfig{apiConfig}, chains, mappings)
	if err != nil {
		return nil, err
	}
	return infos[0], err
}

func (r *routerService) ListAllAvailable() ([]*vo.ApiConfigInfo, error) {
	var (
		apiConfigs []entry.ApiConfig
		chains     []entry.ApiChain
		mappings   []entry.ApiParamMapping
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
		Select("d_api_param_mapping.id, d_api_param_mapping.api_id, d_api_param_mapping.chain_id, d_api_param_mapping,type_id, d_api_param_mapping.param_id, d_api_param_mapping.explain").
		Joins("JOIN d_api_config on d_api_config.id = d_api_param_mapping.api_id").
		Where("d_api_config.status = ?", entry.Available).Order("d_api_param_mapping.id").Find(&mappings).Error; err != nil {
		return nil, err
	}
	return r.join(apiConfigs, chains, mappings)
}

func (r *routerService) join(apiConfigs []entry.ApiConfig, chains []entry.ApiChain,
	mappings []entry.ApiParamMapping) ([]*vo.ApiConfigInfo, error) {
	var (
		apiConfigInfo *vo.ApiConfigInfo
		chainMap      map[int64][]entry.ApiChain
		paramMaps     map[int64][]entry.ApiParamMapping
	)
	apiInfos := make([]*vo.ApiConfigInfo, len(apiConfigs))
	chainMap = make(map[int64][]entry.ApiChain)
	paramMaps = make(map[int64][]entry.ApiParamMapping)
	for _, chain := range chains {
		configChains, ok := chainMap[chain.ID]
		if !ok {
			configChains = make([]entry.ApiChain, 0)
		}
		configChains = append(configChains, chain)
		chainMap[chain.ID] = configChains
	}
	for _, param := range mappings {
		chainParams, ok := paramMaps[param.ChainId]
		if !ok {
			chainParams = make([]entry.ApiParamMapping, 0)
		}
		chainParams = append(chainParams, param)
		paramMaps[param.ApiId] = chainParams
	}
	for index, apiConfig := range apiConfigs {
		apiConfigInfo = &vo.ApiConfigInfo{
			ApiConfig: apiConfig,
		}
		if err := apiConfigInfo.FillChains(chainMap[apiConfig.ID], paramMaps[apiConfig.ID]); err != nil {
			return nil, err
		}
		apiInfos[index] = apiConfigInfo
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
	filtersLength := len(apiConfig.FilterChains)
	if chainsLength == 0 && filtersLength == 0 {
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
	apiChainMapping := map[int][]vo.ApiChainInfo{
		entry.ChainFilter: apiConfig.FilterChains,
		entry.ChainMethod: apiConfig.Chains,
	}
	for key, value := range apiChainMapping {
		if len(value) == 0 {
			continue
		}
		for _, chain := range value {
			chain.Chain.TypeId = key
			if err = tx.Save(&(chain.Chain)).Error; err != nil {
				return
			}
			length := len(chain.ParamMappings) + len(chain.ResultMapping)
			valueStrings = make([]string, 0, length)
			valueArgs = make([]interface{}, 0, length)
			for _, mapping := range chain.ParamMappings {
				valueStrings = append(valueStrings, fmt.Sprintf("(%d, %d, %d, %d, '%s')", apiConfig.ApiConfig.ID,
					chain.Chain.ID, mapping.TypeId, mapping.ParamId, mapping.Explain))
			}
			for _, mapping := range chain.ResultMapping {
				valueStrings = append(valueStrings, fmt.Sprintf("(%d, %d, %d, %d, '%s')", apiConfig.ApiConfig.ID,
					chain.Chain.ID, mapping.TypeId, mapping.ParamId, mapping.Explain))
			}
			smt := "INSERT INTO d_api_param_mapping(api_id, chain_id, type_id, param_id, explain) VALUES " + strings.Join(valueStrings, ",")
			if err = tx.Exec(smt, valueArgs...).Error; err != nil {
				return
			}
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

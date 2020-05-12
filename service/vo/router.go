package vo

import (
	"dubbo-gateway/common"
	"dubbo-gateway/service/entry"
	"github.com/go-errors/errors"
	jsoniter "github.com/json-iterator/go"
	perrors "github.com/pkg/errors"
	"strings"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

type ApiConfigInfo struct {
	entry.ApiConfig
	ResultMapping *common.ApiParamExplain `json:"resultMapping,omitempty"`
	Filter        *ApiFilterInfo          `json:"filters,omitempty"`
	Chains        []*ApiChainInfo         `json:"chains,omitempty"`
}

func (a *ApiConfigInfo) Marshal() error {
	if a.ResultMapping != nil {
		bs, err := json.Marshal(a.ResultMapping)
		if err != nil {
			return err
		}
		a.ApiConfig.ResultMapping = string(bs)
	}
	return nil
}

func (a *ApiConfigInfo) Unmarshal() error {
	str := strings.TrimSpace(a.ApiConfig.ResultMapping)
	if str == "" {
		return nil
	}
	a.ResultMapping = new(common.ApiParamExplain)
	return json.Unmarshal([]byte(str), a.ResultMapping)
}

func (a *ApiConfigInfo) ConvertCache(paramMap map[int64]*MethodDeclaration) (*common.ApiInfo, *common.ApiFilter, error) {
	var (
		filter *common.ApiFilter
		err    error
	)
	info := &common.ApiInfo{
		ApiId:    a.ApiConfig.ID,
		Method:   a.ApiConfig.Method,
		FilterId: a.ApiConfig.FilterId,
	}
	if len(a.Chains) > 0 {
		var son *common.ApiChain
		for i := len(a.Chains) - 1; i >= 0; i-- {
			temp, err := a.Chains[i].ConvertCache(paramMap)
			if err != nil {
				return nil, nil, err
			}
			if son != nil {
				temp.Next = son
			}
			son = temp
		}
		info.MethodChain = son
	}
	if a.ResultMapping == nil && a.ApiConfig.ResultMapping != "" {
		if err = a.Unmarshal(); err != nil {
			return nil, nil, err
		}
	}
	info.ResultRule = a.ResultMapping
	if a.Filter != nil {
		if filter, err = a.Filter.ConvertCache(paramMap); err != nil {
			return nil, nil, err
		}

	}
	return info, filter, nil
}

type ApiParamMapping struct {
	entry.ApiParamMapping
	*common.ApiParamExplain `json:"explain"`
}

func (a *ApiParamMapping) Unmarshal() error {
	if a.ApiParamMapping.Explain == "" {
		return nil
	}
	a.ApiParamExplain = new(common.ApiParamExplain)
	return json.Unmarshal([]byte(a.ApiParamMapping.Explain), a.ApiParamExplain)
}

func (a *ApiParamMapping) Marshall() error {
	if a.ApiParamExplain != nil {
		bs, err := json.Marshal(a.ApiParamExplain)
		if err != nil {
			return err
		}
		a.ApiParamMapping.Explain = string(bs)
	}
	return nil
}

type ApiChainInfo struct {
	entry.ApiChain
	ParamMappings []*ApiParamMapping `json:"paramMappings,omitempty"`
}

func (a *ApiChainInfo) ConvertCache(paramMap map[int64]*MethodDeclaration) (*common.ApiChain, error) {
	md, ok := paramMap[a.ApiChain.MethodId]
	if !ok {
		return nil, errors.New("miss methodId")
	}
	chain := &common.ApiChain{
		ReferenceId: a.ApiChain.ReferenceId,
		ChainId:     a.ApiChain.ID,
		MethodName:  md.MethodName,
	}
	length := len(md.Params)
	if length > 0 {
		chain.ParamClass = make([]string, 0, length)
		chain.ParamTypes = make([]int, 0, length)
		for _, param := range md.Params {
			chain.ParamTypes = append(chain.ParamTypes, param.TypeId)
			chain.ParamClass = append(chain.ParamClass, param.Key)
		}
		chain.ParamRule = make([]*common.ApiParamExplain, len(a.ParamMappings))
		for index, mapping := range a.ParamMappings {
			chain.ParamRule[index] = mapping.ApiParamExplain
		}
	}
	return chain, nil
}

func (a *ApiChainInfo) Unmarshal() error {
	if len(a.ParamMappings) > 0 {
		for _, mapping := range a.ParamMappings {
			if err := mapping.Unmarshal(); err != nil {
				return err
			}
		}
	}
	return nil
}

func (a *ApiChainInfo) Marshal() error {
	if len(a.ParamMappings) > 0 {
		for _, mapping := range a.ParamMappings {
			if err := mapping.Marshall(); err != nil {
				return err
			}
		}
	}
	return nil
}

type ApiFilterInfo struct {
	entry.ApiFilter
	ParamMappings []*ApiParamMapping `json:"paramMappings,omitempty"`
}

func (a ApiFilterInfo) ConvertCache(paramMap map[int64]*MethodDeclaration) (*common.ApiFilter, error) {
	md, ok := paramMap[a.MethodId]
	if !ok {
		return nil, perrors.Errorf("miss methodId[%d], in paramMap: %v", a.MethodId, paramMap)
	}
	filter := &common.ApiFilter{
		FilterId:    a.ID,
		MethodName:  md.MethodName,
		ReferenceId: a.ReferenceId,
	}
	length := len(md.Params)
	if length > 0 {
		filter.ParamClass = make([]string, 0, length)
		filter.ParamTypes = make([]int, 0, length)
		for _, param := range md.Params {
			filter.ParamTypes = append(filter.ParamTypes, param.TypeId)
			filter.ParamClass = append(filter.ParamClass, param.Key)
		}
		filter.ParamRule = make([]*common.ApiParamExplain, len(a.ParamMappings))
		for index, mapping := range a.ParamMappings {
			filter.ParamRule[index] = mapping.ApiParamExplain
		}
	}
	return filter, nil
}

func (a *ApiFilterInfo) Unmarshal() error {
	if len(a.ParamMappings) > 0 {
		for _, mapping := range a.ParamMappings {
			if err := mapping.Unmarshal(); err != nil {
				return err
			}
		}
	}
	return nil
}

func (a *ApiFilterInfo) Marshall() error {
	if len(a.ParamMappings) > 0 {
		for _, mapping := range a.ParamMappings {
			if err := mapping.Marshall(); err != nil {
				return err
			}
		}
	}
	return nil
}

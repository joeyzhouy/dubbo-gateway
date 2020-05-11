package vo

import (
	"dubbo-gateway/common"
	"dubbo-gateway/service/entry"
	"github.com/go-errors/errors"
	jsoniter "github.com/json-iterator/go"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

type ApiConfigInfo struct {
	ApiConfig entry.ApiConfig `json:"config,omitempty"`
	Filter    *ApiFilterInfo  `json:"filters,omitempty"`
	Chains    []*ApiChainInfo `json:"chains,omitempty"`
}

func (a *ApiConfigInfo) ConvertCache(paramMap map[int64]*MethodDeclaration) (*common.ApiInfo, *common.ApiFilter, error) {
	var (
		filter *common.ApiFilter
		err    error
	)
	info := &common.ApiInfo{
		ApiId:  a.ApiConfig.ID,
		Method: a.ApiConfig.Method,
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

func (a *ApiParamMapping) convert() error {
	if a.ApiParamExplain == nil {
		return nil
	}
	bs, err := json.Marshal(a.ApiParamExplain)
	if err != nil {
		return err
	}
	a.ApiParamMapping.Explain = string(bs)
	return nil
}

type ApiChainInfo struct {
	Chain         entry.ApiChain     `json:"chain,omitempty"`
	ParamMappings []*ApiParamMapping `json:"paramMappings,omitempty"`
	ResultMapping *ApiParamMapping   `json:"resultMapping,omitempty"`
}

func (a *ApiChainInfo) ConvertCache(paramMap map[int64]*MethodDeclaration) (*common.ApiChain, error) {
	md, ok := paramMap[a.Chain.MethodId]
	if !ok {
		return nil, errors.New("miss methodId")
	}
	chain := &common.ApiChain{
		ReferenceId: a.Chain.ReferenceId,
		ChainId:     a.Chain.ID,
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
	if a.ResultMapping != nil {
		if err := a.ResultMapping.Unmarshal(); err != nil {
			return err
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
	if a.ResultMapping != nil {
		if err := a.ResultMapping.Marshall(); err != nil {
			return err
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
		return nil, errors.New("miss methodId")
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

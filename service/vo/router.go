package vo

import (
	"dubbo-gateway/service/entry"
	"encoding/json"
)

type ApiConfigInfo struct {
	ApiConfig    entry.ApiConfig `json:"config,omitempty"`
	FilterChains []ApiChainInfo  `json:"filters,omitempty"`
	Chains       []ApiChainInfo  `json:"chains,omitempty"`
}

func (a *ApiConfigInfo) ConvertCache() (*ApiInfo, error) {
	info := &ApiInfo{
		ApiId:  a.ApiConfig.ID,
		Method: a.ApiConfig.Method,
	}
	if len(a.FilterChains) > 0 {
		var son *ApiChain
		for i := len(a.FilterChains) - 1; i >= 0; i-- {
			temp, err := a.FilterChains[i].ConvertCache()
			if err != nil {
				return nil, err
			}
			if son != nil {
				temp.Next = son
			}
			son = temp
		}
		info.FilterChain = son
	}
	if len(a.Chains) > 0 {
		var son *ApiChain
		for i := len(a.Chains) - 1; i >= 0; i-- {
			temp, err := a.FilterChains[i].ConvertCache()
			if err != nil {
				return nil, err
			}
			if son != nil {
				temp.Next = son
			}
			son = temp
		}
		info.MethodChain = son
	}
	return info, nil
}

func (a *ApiConfigInfo) FillChains(chains []entry.ApiChain, mappings []entry.ApiParamMapping) error {
	var apiChainInfo ApiChainInfo
	filterChains := make([]ApiChainInfo, 0)
	methodChains := make([]ApiChainInfo, 0)
	paramMapping := make(map[int64][]entry.ApiParamMapping)
	for _, mapping := range mappings {
		m, ok := paramMapping[mapping.ChainId]
		if !ok {
			m = make([]entry.ApiParamMapping, 0)
		}
		m = append(m, mapping)
		paramMapping[mapping.ChainId] = m
	}
	for _, chain := range chains {
		apiChainInfo = ApiChainInfo{
			Chain: chain,
		}
		if err := apiChainInfo.FillMappings(paramMapping[chain.ID]); err != nil {
			return err
		}
		if chain.TypeId == entry.ChainFilter {
			filterChains = append(filterChains, apiChainInfo)
		} else if chain.TypeId == entry.ChainMethod {
			methodChains = append(methodChains, apiChainInfo)
		}
	}
	a.FilterChains = filterChains
	a.Chains = methodChains
	return nil
}

type ApiParamMapping struct {
	entry.ApiParamMapping
	*ApiParamExplain
}

func (a *ApiParamMapping) U() error {
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

type ApiParamExplain struct {
}

type ApiChainInfo struct {
	Chain         entry.ApiChain    `json:"chain,omitempty"`
	ParamMappings []ApiParamMapping `json:"paramMappings,omitempty"`
	ResultMapping []ApiParamMapping `json:"resultMapping,omitempty"`
}

func (a *ApiChainInfo) ConvertCache() (*ApiChain, error) {
	//TODO
	chain := &ApiChain{
		ReferenceId: a.Chain.ReferenceId,
		ChainId: a.Chain.ID,
		//MethodName: a.Chain.me
	}
	return chain, nil
}

func (a *ApiChainInfo) FillMappings(mappings []entry.ApiParamMapping) error {
	if len(mappings) == 0 {
		return nil
	}
	paramMapping := make([]ApiParamMapping, 0)
	resultMapping := make([]ApiParamMapping, 0)
	for _, mapping := range mappings {
		apiParamMapping := ApiParamMapping{
			ApiParamMapping: mapping,
		}
		if mapping.TypeId == entry.ParamMapping {
			paramMapping = append(paramMapping, apiParamMapping)
		} else if mapping.TypeId == entry.ResultMapping {
			resultMapping = append(resultMapping, apiParamMapping)
		}
	}
	a.ParamMappings = paramMapping
	a.ResultMapping = resultMapping
	return nil
}

package vo

import "sync"

type ApiCache struct {
	sync.RWMutex
	//methodMapping map[string]
	Mappings map[string]*ApiInfo
}

type ApiInfo struct {
	Method      string    `json:"method"`
	ApiId       int64     `json:"apiId"`
	FilterChain *ApiChain `json:"filter"`
	MethodChain *ApiChain `json:"method"`
}

type ApiChain struct {
	ChainId     int64             `json:"chainId"`
	ReferenceId int64             `json:"referenceId"`
	MethodName  string            `json:"methodName"`
	ParamClass  []string          `json:"paramClass"`
	ParamRule   []ApiParamExplain `json:"paramRule"`
	ResultRule  []ApiParamExplain `json:"resultRule"`
	Next        *ApiChain         `json:"next"`
}

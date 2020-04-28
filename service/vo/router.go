package vo

import (
	"dubbo-gateway/service/entry"
)

type ApiConfigInfo struct {
	ApiConfig entry.ApiConfig `json:"config,omitempty"`
	Filter    ApiFilter       `json:"filter,omitempty"`
	Chains    []ApiChainInfo  `json:"chains,omitempty"`
}

type ApiFilter struct {
	entry.ApiFilter
	Method `json:"method"`
}

type ApiChainInfo struct {
	Chain  entry.ApiChain `json:"chain,omitempty"`
	Method `json:"method,omitempty"`
	Rules  []entry.ApiResultRule `json:"rules,omitempty"`
}

package vo

import "dubbo-gateway/service/entry"

type ApiConfigInfo struct {
	ApiConfig entry.ApiConfig `json:"config,omitempty"`
	Filter    entry.ApiFilter `json:"filter,omitempty"`
	Chains    []ApiChainInfo  `json:"chains,omitempty"`
}

type ApiChainInfo struct {
	Chain  entry.ApiChain        `json:"chain,omitempty"`
	Method entry.Method          `json:"method,omitempty"`
	Params []entry.MethodParam   `json:"params,omitempty"`
	Rules  []entry.ApiResultRule `json:"rules,omitempty"`
}

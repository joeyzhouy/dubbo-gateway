package vo

import "dubbo-gateway/service/entry"

type ApiConfigInfo struct {
	entry.ApiConfig
	entry.ApiFilter
	Chains []ApiChainInfo
}

type ApiChainInfo struct {
	entry.ApiChain
	entry.Method
	Params []entry.MethodParam
	Rules []entry.ApiResultRule
}


package vo

import "dubbo-gateway/service/entry"

type Method struct {
	entry.Method
	Params []entry.MethodParam
}

type MethodDesc struct {
	ID            int64  `json:"id"`
	MethodName    string `json:"methodName"`
	InterfaceName string `json:"interfaceName"`
}

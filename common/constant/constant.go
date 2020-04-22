package constant

const (
	ConfGatewayFilePath    = "CONF_GATEWAY_FILE_PATH"
	DefaultGatewayFilePath = "meta/gateway.yml"
	DefaultUserName        = "superUser"

	RootPath          = "/gateway"
	Ids               = "/ids"
	IdPath            = RootPath + Ids
	Nodes             = "/nodes"
	NodePath          = RootPath + Nodes
	Users             = "/users"
	UserPath          = RootPath + Users
	UserInfoPath      = UserPath + "/%d"
	UserNamePath      = UserPath + "/%s"
	Registries        = "/registries"
	RegistryPath      = UserInfoPath + Registries
	RegistryInfoPath  = RegistryPath + "/%d"
	References        = "/references"
	ReferencePath     = RegistryInfoPath + References
	ReferenceInfoPath = ReferencePath + "/%d"
	Methods           = "/methods"
	MethodPath        = ReferenceInfoPath + Methods
	//MethodInfoOPath   = MethodPath + "/%d"

	Api         = "/apis"
	ApiPath     = UserInfoPath + Api
	ApiInfoPath = ApiPath + "/%d"

	Filter         = "/filters"
	FilterPath     = RootPath + Filter
	FilterInfoPath = FilterPath + "/%d"

	RegistrySearchRoot  = RootPath + Registries
	RegistrySearch      = RegistrySearchRoot + "/%d"
	ReferenceSearchRoot = RootPath + References
	ReferenceSearch     = ReferenceSearchRoot + "/%d"
	MethodSearchRoot    = RootPath + Methods
	MethodSearch        = MethodSearchRoot + "/%d"
	ApiSearchRoot       = RootPath + Api
	ApiSearchInfo       = ApiSearchRoot + "/%d"
)

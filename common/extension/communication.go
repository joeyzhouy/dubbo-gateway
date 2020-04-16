package extension

type Mode interface {
	Start() error
	Add(apiId int64) error
	Remove(apiId int64) error
	Refresh() error
}

var modes map[string]func(deploy *Deploy) (Mode, error)

func GetMode(mode string) (Mode, error) {
	if modes[mode] == nil {
		panic("mode for " + mode + " is not existing, make sure you have import the package.")
	}
	return modes[mode](GetDeployConfig())
}

func SetMode(mode string, v func(deploy *Deploy) (Mode, error)) {
	modes[mode] = v
}

func GetConfigMode() (Mode, error) {
	return GetMode(GetDeployConfig().Config.Model)
}

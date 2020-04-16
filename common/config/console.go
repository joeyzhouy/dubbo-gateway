package config


type WebConfig struct {
	Name   string `yaml:"name"`
	Port   int    `yaml:"port"`
	Config struct {
		Session struct {
			Type    string `yaml:"type"`
			Timeout int    `yaml:"time_out"`
			Redis   struct {
				Network  string `yaml:"network"`
				Address  string `yaml:"address"`
				Password string `yaml:"password"`
				DB       int    `yaml:"db"`
			}
		} `yaml:"session"`
	} `yaml:"web"`
}
package config

import "fmt"

type DiscoverConfig struct {
	Timeout  string
	Protocol string
	Address  string
	UserName string
	Password string
}

func (d DiscoverConfig) GetKey() string {
	return fmt.Sprintf("%s://%s|%s|%s?%s", d.Protocol, d.Address, d.UserName, d.Password, d.Timeout)
}

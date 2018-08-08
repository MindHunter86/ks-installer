package config

import "github.com/spf13/viper"

type ServerConfig struct {
	*viper.Viper
}

func NewSytemConfig() (*ServerConfig, error) {
	viper.SetConfigName("ks-installer")

	viper.SetConfigType("yaml")

	viper.AddConfigPath("/etc/ks-installer")
	viper.AddConfigPath("/etc/sysconfig/ks-installer")
	viper.AddConfigPath("$HOME/.ks-installer")
	viper.AddConfigPath("./extras")

	if e := viper.ReadInConfig(); e != nil {
		return nil, e
	}

	return &ServerConfig{viper.GetViper()}, nil
}

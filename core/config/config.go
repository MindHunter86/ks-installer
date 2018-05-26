package config

import (
	"errors"
	"io/ioutil"
	"gopkg.in/yaml.v2"
)

type CoreConfig struct {
	Base struct {
		Log_Level string
		Mysql struct {
			Host string
			Username, Password, Database string
			Migrations_Path string
			Sql_Debug bool
		}
	}
}

func (m *CoreConfig) Parse(cfgPath string) (*CoreConfig, error) {
	// find and read configuration file:
	cfgFile,e := ioutil.ReadFile(cfgPath); if e != nil {
		return nil,errors.New("Could not read configuration file! Error: " + e.Error()) }

	// parse configuration file (YAML synt):
	if e := yaml.UnmarshalStrict(cfgFile, m); e != nil {
		return nil,errors.New("Could not parse configuration file! Error: " + e.Error()) }

	return m,nil
}

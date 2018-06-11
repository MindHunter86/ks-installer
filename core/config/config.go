package config

import (
	"errors"
	"io/ioutil"
	"gopkg.in/yaml.v2"
)

type CoreConfig struct {
	Base struct {
		Log_Level string
		Http struct {
			Listen, Host string
			Read_Timeout, Write_Timeout int }
		Api struct {
			Sign_Secret string }
		Mysql struct {
			Sql_Debug bool
			Host, Username, Password, Database string
			Migrations_Path string }
		Ipmi struct {
			Hostname_Tld string
			CIDR_Block string }
		Queue struct {
			Workers int
			Worker_Capacity int
			Chain_Buffer int }
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

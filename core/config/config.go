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
			Read_Timeout, Write_Timeout int
		}
		Mysql struct {
			Host string
			Username, Password, Database string
			Migrations_Path string
			Sql_Debug bool
		}
		Api struct {
			Sign_Secret string
		}
		Telegram struct {
			Botapi struct {
				Tgrm_Debug bool
				Token string
				Timeout int
			}
			Queue struct {
				Workers, Worker_Capacity, Chain_Buffer int
			}
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

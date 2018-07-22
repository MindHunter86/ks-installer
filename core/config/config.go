package config

import (
	"errors"
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

type CoreConfig struct {
	Base struct {
		Log_Level    string
		Dns_Resolver string
		Http         struct {
			Listen, Host                string
			Read_Timeout, Write_Timeout int
		}
		Api struct {
			Sign_Secret string
		}
		Mysql struct {
			Sql_Debug                          bool
			Host, Username, Password, Database string
			Migrations_Path                    string
		}
		Ipmi struct {
			Hostname_Tld string
			CIDR_Block   string
		}
		Queue struct {
			Workers             int
			Worker_Capacity     int
			Jobs_Chain_Buffer   int
			Max_Job_Fails       int
			Jobs_Retry_Interval int
		}
		Rsview struct {
			Url    string
			Client struct {
				Timeout              int
				Insecure_Skip_Verify bool
			}
			Authentication struct {
				Login, Password string
				Test_String     string
			}
			Access struct { // TODO rename it ?!
				Vlans      []string
				Port_Names []string
				Jun_Names  []string
			}
		}
		Raft struct {
			Is_Master bool
			Nodes []string
			Listen string
			Inmemory_Store bool
			Max_Pool_Size int
			Skip_Join_Errors bool
			Timeouts struct {
				Tcp, Raft int
			}
			Snapshots struct {
				Path string
				Retain_Count int
			}
		}
		Puppet struct {
			Projects map[string]string
			Endpoints map[string]map[string]string
		}
	}
}

func (m *CoreConfig) Parse(cfgPath string) (*CoreConfig, error) {
	// find and read configuration file:
	cfgFile, e := ioutil.ReadFile(cfgPath)
	if e != nil {
		return nil, errors.New("Could not read configuration file! Error: " + e.Error())
	}

	// parse configuration file (YAML synt):
	if e := yaml.UnmarshalStrict(cfgFile, m); e != nil {
		return nil, errors.New("Could not parse configuration file! Error: " + e.Error())
	}

	return m, nil
}

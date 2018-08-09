package config

import (
	"time"
)

type SysConfig struct {
	Base struct {
		LogLevel    string `viper:"log_level"`
		DNSResolver string `viper:"dns_resolver"`
		Http        struct {
			Listen, Host string
			ReadTimeout  time.Duration `viper:"read_timeout"`
			WriteTimeout time.Duration `viper:"write_timeout"`
		}
		Mysql struct {
			SqlDebug           bool `viper:"sql_debug"`
			Host, Database     string
			Username, Password string
			MigrationsPath     string `viper:"migrations_path"`
		}
		Api struct {
			SignSecret string `viper:"sign_secret"`
		}
		Ipmi struct {
			HostnameTLD string `viper:"hostname_tld"`
			CIDRBlock   string `viper:"cidr_block"`
		}
		Queue struct {
			Workers          int
			WorkersCapacity  int           `viper:"workers_capacity"`
			JobChanBuffer    int           `viper:"job_chain_buffer"`
			JobRetryMaxFails int           `viper:"job_retry_max_fails"`
			JobRetryInterval time.Duration `viper:"job_retry_interval"`
		}
		Rsview struct {
			Url    string
			Client struct {
				Timeout            time.Duration
				InsecureSkipVerify bool `viper:"insecure_skip_verify"`
			}
			Authentication struct {
				Login, Password string
				TestString      string `viper:"test_string"`
			}
			AllowRules struct {
				Vlans     []string
				PortNames []string `viper:"port_names"`
				JunNames  []string `viper:"jun_names"`
			}
		}
		Raft struct {
			Nodes          map[string]string
			InMemoryStore  bool `viper:"in_memory_store"`
			MaxPoolSize    int  `viper:"max_pool_size"`
			SkipJoinErrors bool `viper:"skip_join_errors"`
			Timeouts       struct {
				TCP    time.Duration
				Vote   time.Duration
				Commit time.Duration
			}
			Snapshots struct {
				Path        string
				RetainCount int `viper:"retain_count"`
			}
		}
		Puppet struct {
			Projects  map[string]string
			Endpoints map[string]map[string]string
		}
		BoltDB struct {
			Path        string
			Mode        uint32
			LockTimeout time.Duration `viper:"lock_timeout"`
			ReadOnly    bool          `viper:"read_only"`
			NoSync      bool          `viper:"no_sync"`
		}
	}
}

func NewSysConfig() *SysConfig {
	return &SysConfig{}
}

func NewSysConfigWithDefaults() *SysConfig {
	return new(SysConfig).SetDefaults()
}

func (m *SysConfig) SetDefaults() *SysConfig {
	m.Base.LogLevel = "warning"
	m.Base.DNSResolver = ""

	m.Base.Http.Listen = "127.0.0.1:8080"
	m.Base.Http.Host = "ks-installer.example.com"
	m.Base.Http.ReadTimeout = 10000 * time.Millisecond
	m.Base.Http.WriteTimeout = 10000 * time.Millisecond

	m.Base.Api.SignSecret = "secret"

	m.Base.Ipmi.HostnameTLD = "ipmi"
	m.Base.Ipmi.CIDRBlock = "10.0.0.0/8"

	m.Base.Queue.Workers = 1
	m.Base.Queue.WorkersCapacity = 10
	m.Base.Queue.JobChanBuffer = 10
	m.Base.Queue.JobRetryMaxFails = 1
	m.Base.Queue.JobRetryInterval = 5

	m.Base.Rsview.Url = "https://example.com"
	m.Base.Rsview.Client.Timeout = 1 * time.Second
	m.Base.Rsview.Client.InsecureSkipVerify = false
	m.Base.Rsview.Authentication.Login = "ks-installer"
	m.Base.Rsview.Authentication.Password = "1234"
	m.Base.Rsview.Authentication.TestString = "only include ports which are Down"
	m.Base.Rsview.AllowRules.Vlans = []string{}
	m.Base.Rsview.AllowRules.PortNames = []string{}
	m.Base.Rsview.AllowRules.JunNames = []string{}

	m.Base.Raft.Nodes = map[string]string{}
	m.Base.Raft.InMemoryStore = false
	m.Base.Raft.MaxPoolSize = 9
	m.Base.Raft.SkipJoinErrors = false
	m.Base.Raft.Snapshots.Path = "./raft.db"
	m.Base.Raft.Snapshots.RetainCount = 5
	m.Base.Raft.Timeouts.TCP = 10000 * time.Millisecond
	m.Base.Raft.Timeouts.Vote = 10000 * time.Millisecond
	m.Base.Raft.Timeouts.Commit = 10000 * time.Millisecond

	m.Base.BoltDB.Path = "./data.db"
	m.Base.BoltDB.Mode = uint32(0600)
	m.Base.BoltDB.LockTimeout = 5000 * time.Millisecond
	m.Base.BoltDB.ReadOnly = false
	m.Base.BoltDB.NoSync = false

	m.Base.Puppet.Endpoints = map[string]map[string]string{}
	m.Base.Puppet.Projects = map[string]string{}

	return m
}

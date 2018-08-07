package system

import "sync"
import "os"
import "os/signal"
import "syscall"

import "bitbucket.org/mh00net/ks-installer/app/server"
import "bitbucket.org/mh00net/ks-installer/core/config"
import "bitbucket.org/mh00net/ks-installer/core/http"
import "bitbucket.org/mh00net/ks-installer/core/sql"
import "bitbucket.org/mh00net/ks-installer/core/raft"

import "github.com/rs/zerolog"

const (
	sysModuleRaft = uint8(iota)
)

type System struct {
	pModules map[uint8]interface{}

	logger *zerolog.Logger
}

func New() *System {}
func (m *System) Init() error { return nil }
func (m *System) Bootstrap() error { return nil }
func (m *System) DeInit() {}


type Core struct {
	sql  sql.SqlDriver
	http *http.HttpService
	raft *raft.RaftService

	log *zerolog.Logger
	cfg *config.CoreConfig

	appWg sync.WaitGroup
	app   *server.App
}

func (m *Core) SetLogger(l *zerolog.Logger) *Core    { m.log = l; return m }
func (m *Core) SetConfig(c *config.CoreConfig) *Core { m.cfg = c; return m }
func (m *Core) Construct() (*Core, error) {
	var e error

	// application configuration:
	if m.app, e = new(server.App).SetLogger(m.log).SetConfig(m.cfg).Construct(); e != nil {
		return nil, e
	}

	// internal resources configuration:
	m.raft = raft.NewService(m.log)
	if e = m.raft.Init(m.cfg); e != nil {
		return nil, e
	}

//	if m.sql, e = new(sql.MysqlDriver).SetConfig(m.cfg).Construct(); e != nil {
//		return nil, e
//	}
//	m.app.SetSqlDb(m.sql.GetRawDBSession())

//	m.http = new(http.HttpService).SetConfig(m.cfg).SetLogger(m.log).Construct(server.NewApiController())

	return m, nil
}

func (m *Core) Bootstrap(tmpFlag bool) error {

	// define kernel signal catcher:
	var kernSignal = make(chan os.Signal)
	signal.Notify(kernSignal, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL, syscall.SIGQUIT)

	// define global error variables:
	var e error
	var epipe = make(chan error)

	// raft service bootstrap:
	go func(e chan error, wg sync.WaitGroup, t bool) {
		wg.Add(1)
		defer wg.Done()
		e <- m.raft.Bootstrap(t)
	}(epipe, m.appWg, tmpFlag)

	// http service bootstrap:
//	go func(e chan error, wg sync.WaitGroup) {
//		wg.Add(1)
//		defer wg.Done()
//		e <- m.http.Bootstrap()
//	}(epipe, m.appWg)

	// application bootstrap:
//	go func(e chan error, wg sync.WaitGroup) {
//		wg.Add(1)
//		defer wg.Done()
//		e <- m.app.Bootstrap()
//	}(epipe, m.appWg)

	// main application event loop:
LOOP:
	for {
		select {

		// kernel signal catcher:
		case <-kernSignal:
			m.log.Warn().Msg("Syscall.SIG* has been detected! Closing application...")
			break LOOP

		// application error catcher:
		case e = <-epipe:
			if e != nil {
				m.log.Error().Err(e).Msg("Runtime error! Abnormal application closing!")
			}
			break LOOP

			// TODO: automatic application re-bootstrap
		}
	}

	return m.Destruct(&e)
}

func (m *Core) Destruct(e *error) error {
	var err error

	// application destruct:
	if err = m.app.Destruct(); err != nil {
		m.log.Warn().Err(err).Msg("abnormal app exit")
	}

	// internal resources destruct:
	if err = m.raft.DeInit(); err != nil {
		m.log.Warn().Err(err).Msg("abnormal raft.DeInit() exit")
	}
//	if err = m.http.Destruct(); err != nil {
//		m.log.Warn().Err(err).Msg("abnormal http exit")
//	}
//	if err = m.sql.Destruct(); err != nil {
//		m.log.Warn().Err(err).Msg("abnormal sql exit")
//	}

	m.appWg.Wait()
	return *e
}

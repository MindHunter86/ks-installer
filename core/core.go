package core

import (
	"sync"

	"os"
	"os/signal"
	"syscall"

	"github.com/MindHunter86/ks-installer/app/server"
	"github.com/MindHunter86/ks-installer/core/boltdb"
	"github.com/MindHunter86/ks-installer/core/config"
	"github.com/MindHunter86/ks-installer/core/http"
	"github.com/MindHunter86/ks-installer/core/raft"
	"github.com/MindHunter86/ks-installer/core/sql"

	"github.com/rs/zerolog"
)

type Core struct {
	sql  sql.SqlDriver
	http *http.HttpService
	raft *raft.RaftService
	bolt *boltdb.BoltDB

	log *zerolog.Logger
	cfg *config.SysConfig

	appWg sync.WaitGroup
	app   *server.App
}

func (m *Core) SetConfig(config *config.SysConfig) *Core {
	m.cfg = config
	return m
}

func (m *Core) SetLogger(l *zerolog.Logger) *Core {
	m.log = l
	return m
}

func (m *Core) Construct() (*Core, error) {
	var e error

	// app database initialization:
	if m.bolt, e = boltdb.NewBoltDB(m.cfg, m.log); e != nil {
		return nil, e
	}

	// application initialization:
	if m.app, e = server.NewApp(m.log, m.cfg, m.bolt).Construct(); e != nil {
		return nil, e
	}

	// raft consensus proto initialization:
	m.raft = raft.NewService(m.log)
	if e = m.raft.Init(m.cfg, m.bolt.GetDB()); e != nil {
		return nil, e
	}

	// http service initialization:
	m.http = http.NewHTTPService(m.log, m.cfg).Construct(server.NewApiController())

	// todo: 2DELETE
	//	if m.sql, e = new(sql.MysqlDriver).SetConfig(m.cfg).Construct(); e != nil {
	//		return nil, e
	//	}
	//	m.app.SetSqlDb(m.sql.GetRawDBSession())

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
	if err = m.bolt.DeInit(); err != nil {
		m.log.Warn().Err(err).Msg("abnormal bolt.DeInit() exit")
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

package core

import (
	"sync"

	"os"
	"os/signal"
	"syscall"

	"bitbucket.org/mh00net/ks-installer/app"
	"bitbucket.org/mh00net/ks-installer/core/config"
	"bitbucket.org/mh00net/ks-installer/core/sql"

	"github.com/rs/zerolog"
)


type Core struct {
	sql sql.SqlDriver

	log *zerolog.Logger
	cfg *config.CoreConfig

	appWg sync.WaitGroup
	app *app.App
}


func (m *Core) SetLogger(l *zerolog.Logger) *Core { m.log = l; return m }
func (m *Core) SetConfig(c *config.CoreConfig) *Core { m.cfg = c; return m }
func (m *Core) Construct() (*Core, error) {
	var e error

	// application configuration:
	if m.app,e = new(app.App).SetLogger(m.log).SetConfig(m.cfg).Construct(); e != nil { return nil,e }

	// internal resources configuration:
	if m.sql,e = new(sql.MysqlDriver).SetConfig(m.cfg).Construct(); e != nil { return nil,e }

	return m,nil
}

func (m *Core) Bootstrap() error {
	// define kernel signal catcher:
	var kernSignal = make(chan os.Signal)
	signal.Notify(kernSignal, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL, syscall.SIGQUIT)

	// define global error variables:
	var e error
	var epipe = make(chan error)

	// application bootstrap:
	go func(e chan error, wg sync.WaitGroup) {
		wg.Add(1); defer wg.Done()
		epipe <- m.app.Bootstrap()
	}(epipe, m.appWg)

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
			if e != nil { m.log.Error().Err(e).Msg("Runtime error! Abnormal application closing!") }
			break LOOP
			// TODO: automatic application re-bootstrap

		}
	}

	return m.Destruct(&e)
}

func (m *Core) Destruct(e *error) error {
	// application destruct:
	m.app.Destruct()

	// internal resources destruct:
	m.sql.Destruct()

	m.appWg.Wait(); return *e
}

package server

import "database/sql"
import "bitbucket.org/mh00net/ks-installer/core/config"
import "github.com/rs/zerolog"

// PROJECT TODO:

const appVersion = "0.1"

var (
	globLogger    *zerolog.Logger
	globConfig    *config.CoreConfig
	globApi       *apiController
	globSqlDB     *sql.DB
	globQueueChan chan *queueJob
	globRsview    *rsviewClient
)

type App struct {
	queueDp *queueDispatcher
}

// Common methods:
func (m *App) Construct() (*App, error) {
	m.queueDp = newQueueDispatcher()
	globQueueChan = m.queueDp.getQueueChan()

	var e uint8
	globRsview, e = newRsviewClient()
	if e != errNotError {
		globLogger.Error().Str("err", apiErrorsDetail[e]).Msg("RSVIEW ERROR!")
	}

	return m, nil
}

func (m *App) Bootstrap() error {
	m.queueDp.bootstrap()
	return nil
}

func (m *App) Destruct() error {
	m.queueDp.destruct()
	return nil
}

// Helper methods:
func (m *App) SetLogger(l *zerolog.Logger) *App    { globLogger = l; return m }
func (m *App) SetConfig(c *config.CoreConfig) *App { globConfig = c; return m }
func (m *App) SetSqlDb(s *sql.DB) *App             { globSqlDB = s; return m }

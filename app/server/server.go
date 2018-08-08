package server

import "database/sql"
import (
	"github.com/rs/zerolog"
	"github.com/spf13/viper"
		"github.com/MindHunter86/ks-installer/core/boltdb"
)

// PROJECT TODO:

const appVersion = "0.1"

var (
	globLogger    *zerolog.Logger
	globConfig    *viper.Viper
	globApi       *apiController
	globSqlDB     *sql.DB
	globBoldDB    *boltdb.BoltDB
	globQueueChan chan *queueJob
	globRsview    *rsviewClient
	globPuppet    *puppetClient
)

type App struct {
	queueDp *queueDispatcher
}

func NewApp(log *zerolog.Logger, config *viper.Viper, bolt *boltdb.BoltDB) *App {
	globConfig = config
	globBoldDB = bolt
	globLogger = log
	return new(App)
}

// Common methods:
func (m *App) Construct() (*App, error) {
	m.queueDp = newQueueDispatcher()
	globQueueChan = m.queueDp.getQueueChan()

	var err *appError
	globRsview, err = newRsviewClient()
	if err != nil {
		globLogger.Error().Str("err", apiErrorsDetail[err.code]).Msg("RSVIEW ERROR!")
	}

	globPuppet = newPuppetClient()
	if e := globPuppet.parseEndpoints(); e != nil {
		return nil, e
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
func (m *App) SetSqlDb(s *sql.DB) *App             { globSqlDB = s; return m }

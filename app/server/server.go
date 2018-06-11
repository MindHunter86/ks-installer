package server

import "database/sql"
import "bitbucket.org/mh00net/ks-installer/core/config"
import "github.com/rs/zerolog"


// PROJECT TODO:


const appVersion = "0.1"

var (
	globLogger *zerolog.Logger
	globConfig *config.CoreConfig
	globApi *apiController
	globSqlDB *sql.DB
)

type App struct {}


// Common methods:
func (m *App) Construct() (*App, error) {
	return m,nil
}

func (m *App) Bootstrap() error {
	return nil
}

func (m *App) Destruct() error {
	return nil
}

// Helper methods:
func (m *App) SetLogger(l *zerolog.Logger) (*App) { globLogger = l; return m }
func (m *App) SetConfig(c *config.CoreConfig) (*App) { globConfig = c; return m }
func (m *App) SetSqlDb(s *sql.DB) (*App) { globSqlDB = s; return m }

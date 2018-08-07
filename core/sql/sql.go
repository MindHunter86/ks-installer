package sql

import "database/sql"
import "github.com/MindHunter86/ks-installer/core/config"

type SqlDriver interface {
	GetRawDBSession() *sql.DB

	Construct() (SqlDriver, error)
	Destruct() error

	SetConfig(*config.CoreConfig) SqlDriver
}

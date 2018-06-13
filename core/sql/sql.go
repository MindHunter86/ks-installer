package sql

import "database/sql"
import "bitbucket.org/mh00net/ks-installer/core/config"

type SqlDriver interface {
	GetRawDBSession() *sql.DB

	Construct() (SqlDriver, error)
	Destruct() error

	SetConfig(*config.CoreConfig) SqlDriver
}

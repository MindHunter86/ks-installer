package sql

import "database/sql"

type SqlDriver interface {
	GetRawDBSession() *sql.DB

	Construct() (SqlDriver, error)
	Destruct() error
}

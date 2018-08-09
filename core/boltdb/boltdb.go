package boltdb

import (
	"os"

	"github.com/MindHunter86/ks-installer/core/config"
	bolt "github.com/coreos/bbolt"
	"github.com/rs/zerolog"
)

type BoltDB struct {
	db  *bolt.DB
	log *zerolog.Logger
}

func NewBoltDB(cnf *config.SysConfig, l *zerolog.Logger) (*BoltDB, error) {
	var e error

	var m = &BoltDB{
		log: l,
	}

	m.log.Debug().Str("givenPath", cnf.Base.BoltDB.Path).Msg("")

	m.db, e = bolt.Open(cnf.Base.BoltDB.Path, os.FileMode(cnf.Base.BoltDB.Mode), &bolt.Options{
		Timeout:  cnf.Base.BoltDB.LockTimeout,
		ReadOnly: cnf.Base.BoltDB.ReadOnly,
		NoSync:   cnf.Base.BoltDB.NoSync,
	})
	if e != nil {
		return nil, e
	}

	m.log.Debug().Msg("boltdb instance has been successfully created")

	return m, nil
}

func (m *BoltDB) Init() error {
	var (
		e    error
		dbTx *bolt.Tx
	)

	dbTx, e = m.db.Begin(true)
	if e != nil {
		return e
	}
	defer dbTx.Rollback()

	if dbTx.Bucket([]byte("system")) == nil {
		m.log.Warn().Msg("Internal store is not valid! Trying to initialize new schema...")

		if _, e = dbTx.CreateBucket([]byte("system")); e != nil {
			return e
		}

		if e = dbTx.Commit(); e != nil {
			return e
		}
	}

	return nil
}

func (m *BoltDB) Bootstrap() error { return nil }

func (m *BoltDB) DeInit() error {
	return m.db.Close()
}

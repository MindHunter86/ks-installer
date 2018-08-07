package boltdb

import (
	bolt "github.com/coreos/bbolt"
	"github.com/rs/zerolog"
	"github.com/MindHunter86/ks-installer/core/config"
	"time"
	"os"
)

type BoltDB struct {
	db *bolt.DB
	log *zerolog.Logger
}

func NewBoltDB(cnf *config.CoreConfig, l *zerolog.Logger) (*BoltDB, error) {
	var (
		e error
		m *BoltDB
	)

	m.log = l

	m.db, e = bolt.Open(cnf.Base.Store.Path, os.FileMode(cnf.Base.Store.Mode), &bolt.Options{
		Timeout: time.Duration(cnf.Base.Store.Lock_Timeout) * time.Millisecond,
		ReadOnly: cnf.Base.Store.Read_Only,
		NoSync: cnf.Base.Store.No_Sync,
	})
	if e != nil {
		return nil, e
	}

	return m, nil
}

func (m *BoltDB) GetDB() *bolt.DB {
	return m.db
}

func (m *BoltDB) Init() error {
	var (
		e error
		dbTx *bolt.Tx
	)

	dbTx, e = m.db.Begin(true); if e != nil {
		return e
	}
	defer dbTx.Rollback()

	if dbTx.Bucket([]byte("system")) == nil {
		m.log.Warn().Msg("Internal store is not valid! Trying to initialize new schema...")

		if _,e = dbTx.CreateBucket([]byte("system")); e != nil {
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
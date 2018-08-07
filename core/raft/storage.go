package raft

import (
	"sync"
	"io"
	hraft "github.com/hashicorp/raft"
	bolt "github.com/coreos/bbolt"
)

type Store struct {
	bdb *bolt.DB

	sync.RWMutex
	m map[string]string
}

func newStore(b *bolt.DB) *Store {
	return &Store{
		bdb: b,
		m: make(map[string]string),
	}
}

func (m *Store) Get(key string) string {
	m.RLock()
	defer m.RUnlock()
	return m.m[key]
}

func (m *Store) Set(key, value string) {
	m.Lock()
	defer m.Unlock()
	m.m[key] = value
}

func (m *Store) Delete(key string) {
	m.Lock()
	defer m.Unlock()
}


type raftFSM Store

func (m *raftFSM) Apply(l *hraft.Log) interface{} {
	return nil
}
func (m *raftFSM) Snapshot() (hraft.FSMSnapshot,error) {
	return nil,nil
}
func (m *raftFSM) Restore(rc io.ReadCloser) error {
	return nil
}


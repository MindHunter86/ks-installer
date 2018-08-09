package raft

import (
	"encoding/json"
	"errors"
	"io"
	"sync"
	"time"

	"github.com/coreos/bbolt"
	hraft "github.com/hashicorp/raft"
)

const (
	raftActSet = uint8(iota)
	raftActDel
)

var (
	errRaftIsNotLeader     = errors.New("the current node is not leader")
	errRaftAbnormalCommand = errors.New("abnormal command has been received")
)

type (
	Store struct {
		bdb *bolt.DB
		rft *hraft.Raft

		commTimeout time.Duration

		sync.RWMutex
		m map[string]string
	}

	raftFSM Store

	raftCmd struct {
		Act                uint8
		Bucket, Key, Value string
	}

	fsmSnapshot struct {
		store map[string]string
	}
)

// Store methods:
func newStore(b *bolt.DB, t time.Duration) *Store {
	return &Store{
		bdb:         b,
		m:           make(map[string]string),
		commTimeout: t,
	}
}

func (m *Store) Get(key string) string {
	m.RLock()
	defer m.RUnlock()
	return m.m[key]
}

func (m *Store) Set(bucket, key, value string) error {
	if m.rft.State() != hraft.Leader {
		return errRaftIsNotLeader // todo: add leader-forwarding (masterhost:port/internal/master_request?req=%s)
	}

	buf, e := json.Marshal(&raftCmd{
		Act:    raftActSet,
		Bucket: bucket,
		Key:    key,
		Value:  value,
	})
	if e != nil {
		return e
	}

	fut := m.rft.Apply(buf, m.commTimeout)
	return fut.Error()
}

func (m *Store) Del(bucket, key string) error {
	if m.rft.State() != hraft.Leader {
		return errRaftIsNotLeader // todo: add leader-forwarding (masterhost:port/internal/master_request?req=%s)
	}

	buf, e := json.Marshal(&raftCmd{
		Act:    raftActDel,
		Bucket: bucket,
		Key:    key,
	})
	if e != nil {
		return e
	}

	fut := m.rft.Apply(buf, m.commTimeout)
	return fut.Error()
}

// raftFSM methods:
func (m *raftFSM) Apply(l *hraft.Log) interface{} {
	var cmd *raftCmd

	if e := json.Unmarshal(l.Data, &cmd); e != nil {
		panic(e)
	}

	switch cmd.Act {
	case raftActSet:
		return m.applySet(cmd.Bucket, cmd.Key, cmd.Value)
	case raftActDel:
		return m.applyDel(cmd.Bucket, cmd.Key)
	default:
		panic(errRaftAbnormalCommand.Error())
	}
}

func (m *raftFSM) Snapshot() (hraft.FSMSnapshot, error) {
	m.RLock()
	defer m.RUnlock()

	storeCopy := make(map[string]string)
	for k, v := range m.m {
		storeCopy[k] = v
	}

	return &fsmSnapshot{
		store: storeCopy,
	}, nil
}

func (m *raftFSM) Restore(rc io.ReadCloser) error {
	storeCopy := make(map[string]string)

	if e := json.NewDecoder(rc).Decode(&storeCopy); e != nil {
		return e
	}

	// Set the state from the snapshot, no lock required according to Hashicorp docs.
	m.m = storeCopy
	return nil
}

func (m *raftFSM) applySet(bucket, key, value string) interface{} {
	m.Lock()
	defer m.Unlock()
	m.m[key] = value // todo: replce in mem map with boldb
	return nil
}

func (m *raftFSM) applyDel(bucket, key string) interface{} {
	m.Lock()
	defer m.Unlock()
	delete(m.m, key)
	return nil
}

// fsmSnapshot methods:
func (m *fsmSnapshot) Persist(sink hraft.SnapshotSink) error {
	err := func() error {
		buf, e := json.Marshal(m.store)
		if e != nil {
			return e
		}

		if _, e := sink.Write(buf); e != nil {
			return e
		}

		return sink.Close()
	}()

	if err != nil {
		sink.Cancel()
	}

	return nil
}

func (m *fsmSnapshot) Release() {}

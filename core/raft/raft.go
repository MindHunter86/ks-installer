package raft

import "errors"
import "os"
import "net"
import "path/filepath"
import "time"

import "github.com/MindHunter86/ks-installer/core/config"

import "github.com/rs/zerolog"
import hraft "github.com/hashicorp/raft"
import hraftboltdb "github.com/hashicorp/raft-boltdb"
import bolt "github.com/coreos/bbolt"

var (
	errRaftAbnormalNodesCount = errors.New("The number of nodes can not be zero!")
	errRaftMissingLocalID     = errors.New("Could not find local ID from raft node list")
)

type RaftService struct {
	raft  *hraft.Raft
	store *Store

	localId      string
	nodes        map[string]*net.TCPAddr
	skipJoinErrs bool

	config        *hraft.Config
	configuration *hraft.Configuration
	logStore      hraft.LogStore
	stableStore   hraft.StableStore
	snapStore     hraft.SnapshotStore
	transport     hraft.Transport

	logger   *zerolog.Logger
	donePipe chan struct{}
}

func NewService(l *zerolog.Logger) *RaftService {
	return &RaftService{
		logger:   l,
		nodes:    make(map[string]*net.TCPAddr),
		donePipe: make(chan struct{}, 1),
	}
}

func (m *RaftService) Init(c *config.SysConfig, b *bolt.DB) error {
	var e error
	var clusterServers []hraft.Server

	if len(c.Base.Raft.Nodes) == 0 {
		return errRaftAbnormalNodesCount
	}

	for id, ip := range c.Base.Raft.Nodes {
		addr, e := net.ResolveTCPAddr("tcp", ip)
		if e != nil {
			m.logger.Error().Err(e).Msg("unable to build node list")
		}

		if id == "" {
			id, e = os.Hostname()
			if e != nil {
				return e
			}

			m.localId = id
			m.logger.Debug().Strs("localRaftNode", []string{
				id,
				addr.String()}).Msg("Local Raft IP parsed")
		}

		m.nodes[id] = addr

		clusterServers = append(clusterServers, hraft.Server{
			ID:      hraft.ServerID(id),
			Address: hraft.ServerAddress(addr.String()),
		})
	}

	if m.localId == "" {
		return errRaftMissingLocalID
	}

	m.transport, e = hraft.NewTCPTransport(
		m.nodes[m.localId].String(),
		m.nodes[m.localId],
		c.Base.Raft.MaxPoolSize,
		c.Base.Raft.Timeouts.TCP,
		os.Stderr,
	)
	if e != nil {
		return e
	}

	m.snapStore, e = hraft.NewFileSnapshotStore(
		filepath.Dir(c.Base.Raft.Snapshots.Path),
		c.Base.Raft.Snapshots.RetainCount,
		os.Stderr,
	)
	if e != nil {
		return e
	}

	switch c.Base.Raft.InMemoryStore {
	case true:
		m.logStore = hraft.NewInmemStore()
		m.stableStore = hraft.NewInmemStore()
	default:
		boltStore, e := hraftboltdb.NewBoltStore(c.Base.Raft.Snapshots.Path)
		if e != nil {
			return e
		}

		m.stableStore = boltStore
		m.logStore = boltStore
	}

	m.config = hraft.DefaultConfig()
	m.config.LocalID = hraft.ServerID(m.localId)
	m.configuration = &hraft.Configuration{
		Servers: clusterServers,
	}

	m.store = newStore(b, c.Base.Raft.Timeouts.Commit)

	m.skipJoinErrs = c.Base.Raft.SkipJoinErrors
	return nil
}

func (m *RaftService) Bootstrap(forceBootstrap bool) error {

	var e error
	if m.raft, e = hraft.NewRaft(m.config, (*raftFSM)(m.store), m.logStore, m.stableStore, m.snapStore, m.transport); e != nil {
		return e
	}

	if ft := m.raft.BootstrapCluster(*m.configuration); ft.Error() != nil {
		if ft.Error() != hraft.ErrCantBootstrap {
			m.logger.Error().Err(ft.Error()).Msg("unable to bootstrap the cluster")
			return ft.Error()
		}

		if forceBootstrap {
			m.logger.Warn().Msg("unable to bootstrap the new cluster because of its existence, trying to reconnect to nodes")
			for id, addr := range m.nodes {
				m.logger.Debug().Str("node", id).Msg("trying to join a new peer")
				if e := m.join(id, addr.String()); e != nil {
					m.logger.Warn().Err(e).Str("node", id).Msg("")
					if !m.skipJoinErrs {
						return errors.New("unable to bootstrap the cluster while one of the nodes is fails")
					}
				}
			}
		}
	}

	m.logger.Debug().Msg("raft bootstrap done")

	if m.logger.Debug().Enabled() {
		tckr := time.NewTicker(5 * time.Second)
		tckr.Stop() // XXX

		for {
			select {
			case <-m.donePipe:
				tckr.Stop()
				return nil
			case <-tckr.C:
				m.logger.Debug().Str("raft_state", m.raft.State().String()).Msg("raft state")
				for k, v := range m.raft.Stats() {
					m.logger.Debug().Str(k, v).Msg("raft debug stats")
				}
			}
		}
	}

	<-m.donePipe
	return nil
}

func (m *RaftService) DeInit() error {

	close(m.donePipe)

	defer func(l *zerolog.Logger) {
		if r := recover(); r != nil {
			l.Error().Interface("caught panic", r).Msg("PANIC! The function has been recovered!")
		}
	}(m.logger)

	return m.raft.Shutdown().Error()
}

func (m *RaftService) join(nodeId, nodeAddr string) error {
	m.logger.Info().Str("node", nodeId).Str("addr", nodeAddr).Msg("recevied join request from remote node")

	cnfFuture := m.raft.GetConfiguration()
	if cnfFuture.Error() != nil {
		m.logger.Error().Err(cnfFuture.Error()).Msg("failed to get the raft configuration")
		return cnfFuture.Error()
	}

	for _, srv := range cnfFuture.Configuration().Servers {
		if srv.ID == hraft.ServerID(nodeId) || srv.Address == hraft.ServerAddress(nodeAddr) {
			if srv.ID == hraft.ServerID(nodeId) && srv.Address == hraft.ServerAddress(nodeAddr) {
				m.logger.Info().Str("node", nodeId).Str("addr", nodeAddr).Msg("remote node already member of the cluster, ignoring join request")
				return nil
			}

			future := m.raft.RemoveServer(srv.ID, 0, 0)
			if future.Error() != nil {
				m.logger.Error().Err(future.Error()).Str("node", nodeId).Str("addr", nodeAddr).Msg("error removing existing node")
				return future.Error()
			}
		}
	}

	newVoter := m.raft.AddVoter(hraft.ServerID(nodeId), hraft.ServerAddress(nodeAddr), 0, 0)
	if newVoter.Error() != nil {
		return newVoter.Error()
	}

	m.logger.Info().Str("node", nodeId).Str("addr", nodeAddr).Msg("node joined successfully")
	return nil
}

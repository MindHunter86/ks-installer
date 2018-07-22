package raft

import "context"
import "errors"
import "io"
import "os"
import "sync"
import "net"
import "path/filepath"
import "time"

import "bitbucket.org/mh00net/ks-installer/core/config"

import "github.com/rs/zerolog"
import hraft "github.com/hashicorp/raft"
import hraftboltdb "github.com/hashicorp/raft-boltdb"


type RaftService struct {
	raft *hraft.Raft
	store *Store

	is_master bool
	skipJoinErrs bool
	nodes map[string]string
	config *hraft.Configuration

	logger *zerolog.Logger
	donePipe chan struct{}
}

type Store struct {
	sync.Mutex
	m map[string]string
}

func (m *Store) Get(string) (string,bool) { return "",false }
func (m *Store) Set(string, string) (bool) { return false }
func (m *Store) Delete(string) (bool) { return false }


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


func newStore() *Store {
	return &Store{
		m: make(map[string]string),
	}
}

func NewService(l *zerolog.Logger) (*RaftService) {
	return &RaftService{
		logger: l,
		store: newStore(),
		nodes: make(map[string]string),
		donePipe: make(chan struct{}, 1),
	}
}

func (m *RaftService) Init(c *config.CoreConfig) error {
	var e error

	addr,e := net.ResolveTCPAddr("tcp", c.Base.Raft.Listen); if e != nil {
		return e
	}

	tcpTrans,e := hraft.NewTCPTransport(
		c.Base.Raft.Listen,
		addr,
		c.Base.Raft.Max_Pool_Size,
		time.Duration(c.Base.Raft.Timeouts.Tcp) * time.Millisecond,
		os.Stderr )
	if e != nil { return e }

	snapStore,e := hraft.NewFileSnapshotStore(
		filepath.Dir(c.Base.Raft.Snapshots.Path),
		c.Base.Raft.Snapshots.Retain_Count,
		os.Stderr )
	if e != nil { return e }

	var logStore hraft.LogStore
	var stableStore hraft.StableStore

	if c.Base.Raft.Inmemory_Store {
		logStore = hraft.NewInmemStore()
		stableStore = hraft.NewInmemStore()
	} else {
		boltDb,e := hraftboltdb.NewBoltStore(c.Base.Raft.Snapshots.Path)
		if e != nil { return e }

		logStore = boltDb
		stableStore = boltDb
	}

	var raftConfig = hraft.DefaultConfig()
	if hostname,e := os.Hostname(); e != nil {
		raftConfig.LocalID = hraft.ServerID(addr.IP.String())
		m.logger.Warn().Err(e).Msg("unable to get local hostname, ipv4 address will be used as node ID")
	} else {
		raftConfig.LocalID = hraft.ServerID(hostname)
	}

	if m.raft,e = hraft.NewRaft(raftConfig, (*raftFSM)(m.store), logStore, stableStore, snapStore, tcpTrans); e != nil {
		return e
	}

	m.is_master = c.Base.Raft.Is_Master

	// if node is not master, complete the init():
	if ! m.is_master { return nil }

	// else continue the initialization:
	m.skipJoinErrs = c.Base.Raft.Skip_Join_Errors

	m.config = &hraft.Configuration{
		Servers: []hraft.Server{
			{
				ID: raftConfig.LocalID,
				Address: tcpTrans.LocalAddr(),
			},
		},
	}

	var resolver = new(net.Resolver)
	if c.Base.Dns_Resolver != "" {
		resolver.Dial = func(ctx context.Context, network, server string) (net.Conn, error) {
			return new(net.Dialer).DialContext(ctx, network, c.Base.Dns_Resolver)
		}
	}

	for _,node := range c.Base.Raft.Nodes {

		ip,e := net.ResolveTCPAddr("tcp4", node); if e != nil {
			m.logger.Warn().Str("node", node).Msg("node has an invalid ipv4 address, it will be omitted")
			continue
		}

		fqdn,e := resolver.LookupAddr(context.Background(), ip.IP.String())
		if e != nil {
			return e
		}

		if len(fqdn) == 0 {
			m.logger.Warn().Str("node", node).Msg("unable to resolve node, ipv4 address will be used as node ID")
			m.nodes[ip.IP.String()] = ip.String()
			continue
		}

		if len(fqdn) > 1 {
			m.logger.Warn().Str("node", node).Strs("hostnames", fqdn).Msg("resolved node has 2 or more hostnames")
			m.logger.Warn().Str("node", node).Str("selected", fqdn[0]).Msg("to resolve the conflict the first hostname will be used")
		}

		m.nodes[fqdn[0]] = ip.String()
	}

	return nil
}

func (m *RaftService) Bootstrap() error {

	if m.is_master {
		m.logger.Debug().Msg("node is master, trying to bootstrap the cluster")

		f := m.raft.BootstrapCluster(*m.config); if f.Error() != nil {
			return f.Error()
		}

		for {
			time.Sleep(1)
			if m.raft.State() == hraft.Leader {
				break
			}
		}

		for id,addr := range m.nodes {
			if e := m.join(id, addr); e != nil {
				m.logger.Warn().Err(e).Str("node", id).Msg("")
				if ! m.skipJoinErrs {
					return errors.New("unable to bootstrap the cluster while one of the nodes is fails")
				}
			}
		}
	} else {
		m.logger.Debug().Msg("node is not master, waiting for cluster invitation")

		// is no master wait while raftState == Follower ? (Candidate ?)
		for {
			if m.raft.State() != hraft.Follower {
				break
			}

			m.logger.Debug().Msg("no cluster invitation received, sleep")
			time.Sleep(5 * time.Second)
		}

	}

	if m.logger.Debug().Enabled() {
		tckr := time.NewTicker(5 * time.Second)

		for {
			select {
				case <-m.donePipe:
					tckr.Stop()
					return nil
				case <-tckr.C:
					m.logger.Debug().Str("raft_state", m.raft.State().String()).Msg("raft state")
					for k,v := range m.raft.Stats() {
						m.logger.Debug().Str(k,v).Msg("raft debug stats")
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

	cnfFuture := m.raft.GetConfiguration(); if cnfFuture.Error() != nil {
		m.logger.Error().Err(cnfFuture.Error()).Msg("failed to get the raft configuration")
		return cnfFuture.Error()
	}

	for _,srv := range cnfFuture.Configuration().Servers {
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

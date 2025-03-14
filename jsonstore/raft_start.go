package jsonstore

import (
	"errors"
	"fmt"
	"io"
	"time"

	hclog "github.com/hashicorp/go-hclog"
	"github.com/hashicorp/raft"
)

var (
	LeaderDifferent = errors.New("Different Leader")
)

type RaftInterface struct {
	raftinterface *raft.Raft
	configfile    string
	logstorefile  string
	snapshotdir   string
	myid          string
	myaddr        string
	mytransport   *raft.NetworkTransport
	stablestore   *JsonStableStore
	logstore      *JsonLogStore
	configuration *raft.Configuration
	snapshotstore *raft.FileSnapshotStore
	config        *raft.Config
	logger        hclog.Logger
}

func NewRaftInterface(configfile, logstorefile, stablestorefile, snapshotstoredir,
	transport string, serverid string, logger hclog.Logger, writer io.Writer) (*RaftInterface, error) {
	configuration, err := BootstrapConfig(configfile)
	if err != nil {
		return nil, err
	}
	stablestore, err := NewJsonStableStore(stablestorefile)
	if err != nil {
		return nil, err
	}

	logstore, err := NewJsonLogStore(logstorefile)
	if err != nil {
		return nil, err
	}
	snapshotstore, err := raft.NewFileSnapshotStoreWithLogger(snapshotstoredir, 3, logger)
	if err != nil {
		return nil, err
	}

	conf := raft.DefaultConfig()
	conf.SnapshotThreshold = 400
	conf.SnapshotInterval = time.Second * 60
	conf.Logger = logger
	conf.LocalID = raft.ServerID(serverid)
	tcptransport, err := raft.NewTCPTransport(transport, nil, 10, 10*time.Second, writer)
	err = raft.BootstrapCluster(conf, logstore,
		stablestore,
		snapshotstore, tcptransport, *configuration)

	// Error stating cluster already bootstrapped can be be safely ignored
	if err != nil && err != raft.ErrCantBootstrap {
		panic(err)
	}
	fsm, err := NewFsm()
	if err != nil {
		return nil, err
	}
	raftobj, err := raft.NewRaft(conf, fsm, logstore, stablestore, snapshotstore, tcptransport)
	raftin := &RaftInterface{}
	raftin.configfile = configfile
	raftin.logstore = logstore
	raftin.config = conf
	raftin.stablestore = stablestore
	raftin.snapshotstore = snapshotstore
	raftin.myid = string(conf.LocalID)
	raftin.myaddr = transport
	raftin.mytransport = tcptransport
	raftin.snapshotdir = snapshotstoredir
	raftin.logstorefile = logstorefile
	raftin.raftinterface = raftobj
	raftin.logger = logger

	return raftin, nil

}

func (raftin *RaftInterface) Leader() string {
	server := raftin.raftinterface.Leader()
	raftin.logger.Info("Leader", "Server", server)
	return string(server)
}

func (raftin *RaftInterface) LeaderWithID() (string, string) {
	server, id := raftin.raftinterface.LeaderWithID()
	raftin.logger.Info("Leader", "Server", server)
	return string(server), string(id)
}

func (raftin *RaftInterface) AddKV(key string, value string) error {
	cmd := fmt.Sprintf("A:%d:%d:%s%s", len(key), len(value), key, value)
	future := raftin.raftinterface.Apply([]byte(cmd), 30*time.Second)
	err := future.Error()
	if err == raft.ErrNotLeader {
		err = LeaderDifferent
	}
	return err
}

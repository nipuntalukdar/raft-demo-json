package jsonstore

import (
	"io"
	"time"

	hclog "github.com/hashicorp/go-hclog"
	"github.com/hashicorp/raft"
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
	logger        *hclog.Logger
}

func NewRaftInterface(configfile, logstorefile, stablestorefile, snapshotstoredir,
	transport string, serverid string, logger *hclog.Logger, writer io.Writer) (*RaftInterface, error) {
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
	snapshotstore, err := raft.NewFileSnapshotStoreWithLogger(snapshotstoredir, 3, *logger)
	if err != nil {
		return nil, err
	}

	conf := raft.DefaultConfig()
	conf.SnapshotThreshold = 400
	conf.SnapshotInterval = time.Second * 60
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

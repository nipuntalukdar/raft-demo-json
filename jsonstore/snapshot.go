package jsonstore

import (
	"github.com/hashicorp/raft"
)

type Snapshot struct {
	data []byte
}

func NewSnapshot(data []byte) Snapshot {
	return Snapshot{data: data}
}

func (snapshot Snapshot) Persist(sink raft.SnapshotSink) error {
	_, err := sink.Write(snapshot.data)
	return err
}

func (snapshot Snapshot) Release() {
	snapshot.data = nil
}

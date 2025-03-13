package jsonstore

import (
	"github.com/hashicorp/raft"
	"testing"
	"time"
)

func Test_JsonLogStore(t *testing.T) {
	js, err := NewJsonLogStore("/tmp/a.json")
	if err != nil {
		t.Fatal("Failed to open file", err)
	}

	log := &raft.Log{AppendedAt: time.Now(), Index: 1, Term: 1, Type: raft.LogCommand, Data: []byte("Hello")}
	js.StoreLog(log)
	js.StoreLog(log)

	var logs []*raft.Log

	for i := 0; i < 200; i++ {
		logs = append(logs, &raft.Log{AppendedAt: time.Now(), Index: 2 + uint64(i), Term: 1,
			Type: raft.LogCommand, Data: []byte("Hello2")})
	}
	js.StoreLogs(logs)

	js.DeleteRange(3, 198)

}

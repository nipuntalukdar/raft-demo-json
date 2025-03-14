package jsonstore

import (
	"testing"
	"time"

	"github.com/hashicorp/raft"
)

func TestFsm(t *testing.T) {
	fsm, err := NewFsm()
	if err != nil {
		t.Fatal(err)
	}
	log := &raft.Log{AppendedAt: time.Now(), Index: 1, Term: 1, Type: raft.LogCommand,
		Data: []byte("A:5:5:HelloWorld")}
	fsm.Apply(log)
	log = &raft.Log{AppendedAt: time.Now(), Index: 1, Term: 1, Type: raft.LogCommand,
		Data: []byte("A:2:5:HiHello")}
	fsm.Apply(log)
	value, err := fsm.Get("Hi")
	if err != nil || value != "Hello" {
		t.Fatal("Key not found")
	}

	value, err = fsm.Get("Hello")
	if err != nil || value != "World" {
		t.Fatal("Key not found")
	}

	log = &raft.Log{AppendedAt: time.Now(), Index: 1, Term: 1, Type: raft.LogCommand, Data: []byte("D:Hi")}
	fsm.Apply(log)
	_, err = fsm.Get("Hi")
	if err != ErrKeyNotFound {
		t.Fatal("Key was not deleted")
	}

}

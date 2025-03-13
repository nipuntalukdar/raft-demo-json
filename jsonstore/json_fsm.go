package jsonstore

import (
	"encoding/json"
	"errors"
	"io"
	"strings"
	"sync"

	"github.com/hashicorp/raft"
)

var (
	ErrKeyNotFound = errors.New("not found")
)

type Fsm struct {
	kv   map[string]string
	lock *sync.Mutex
}

func NewFsm() (fsm *Fsm, err error) {
	kv := make(map[string]string)
	fsm = &Fsm{kv: kv, lock: &sync.Mutex{}}
	err = nil
	return
}

func (fsm *Fsm) Apply(log *raft.Log) interface{} {
	ds := string(log.Data)
	kvs := strings.Split(ds, ":")
	if len(kvs) < 2 || len(kvs) > 3 {
		return errors.New("Wrong log")
	}
	if kvs[0] == "A" && len(kvs) == 3 {
		// Add the key
		fsm.add(kvs[1], kvs[2])
	} else if kvs[0] == "D" && len(kvs) == 2 {
		// Delete the key
		fsm.delete(kvs[1])
	} else {
		return errors.New("Incorrect log")
	}
	return nil
}

func (fsm *Fsm) Snapshot() (raft.FSMSnapshot, error) {
	fsm.lock.Lock()
	defer fsm.lock.Unlock()
	data, err := json.Marshal(fsm.kv)
	if err == nil {
		return NewSnapshot(data), nil
	}
	return NewSnapshot(nil), err
}

func (fsm *Fsm) Restore(inp io.ReadCloser) error {
	fsm.lock.Lock()
	defer fsm.lock.Unlock()
	data, err := io.ReadAll(inp)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, &fsm.kv)
}

func (fsm *Fsm) add(key string, value string) {
	fsm.lock.Lock()
	defer fsm.lock.Unlock()
	fsm.kv[key] = value
}

func (fsm *Fsm) Get(key string) (value string, err error) {
	fsm.lock.Lock()
	defer fsm.lock.Unlock()
	value, exists := fsm.kv[key]
	if !exists {
		err = ErrKeyNotFound
	}
	return
}

func (fsm *Fsm) delete(key string) (err error) {
	fsm.lock.Lock()
	defer fsm.lock.Unlock()
	delete(fsm.kv, key)
	return
}

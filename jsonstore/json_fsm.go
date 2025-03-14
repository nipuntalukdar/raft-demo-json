package jsonstore

import (
	"encoding/json"
	"errors"
	"io"
	"strconv"
	"strings"
	"sync"

	"github.com/hashicorp/raft"
)

var (
	ErrKeyNotFound  = errors.New("not found")
	ErrIncorrectLog = errors.New("Incorrect log")
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
	first, second, found := strings.Cut(ds, ":")
	if !found || second == "" {
		return ErrIncorrectLog
	}
	if first == "A" {
		kvs := strings.SplitN(second, ":", 3)
		if len(kvs) != 3 {
			return ErrIncorrectLog
		}
		len1, err := strconv.Atoi(kvs[0])
		if err != nil {
			return ErrIncorrectLog
		}
		len2, err := strconv.Atoi(kvs[1])
		if err != nil || len(kvs[2]) != (len1+len2) {
			return ErrIncorrectLog
		}
		fsm.add(kvs[2][:len1], kvs[2][len1:])
	} else if first == "D" {
		fsm.delete(second)
	} else {
		return ErrIncorrectLog
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

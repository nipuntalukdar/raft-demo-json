package jsonstore

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/emirpasic/gods/v2/maps/treemap"
	"github.com/hashicorp/raft"
)

type JsonLogStore struct {
	jsonfilepath string
	kv           *treemap.Map[uint64, *raft.Log]
	lock         sync.Mutex
}

func NewJsonLogStore(jsonfilepath string) (js *JsonLogStore, err error) {
	kv := treemap.New[uint64, *raft.Log]()
	_, err = os.Stat(jsonfilepath)
	if err == nil {
		file, err := os.OpenFile(jsonfilepath, os.O_RDONLY, 0600)
		if err != nil {
			panic(err)
		}
		data, err := io.ReadAll(file)
		err = json.Unmarshal(data, &kv)
		if err != nil {
			panic(err)
		}
	} else {
		err = nil
	}
	js = &JsonLogStore{jsonfilepath: jsonfilepath, kv: kv, lock: sync.Mutex{}}
	return
}

func (js *JsonLogStore) FirstIndex() (index uint64, err error) {
	err = nil
	index = 0
	js.lock.Lock()
	defer js.lock.Unlock()
	index, _, ok := js.kv.Min()
	if !ok {
		index = 0
	}
	return
}

func (js *JsonLogStore) LastIndex() (index uint64, err error) {
	err = nil
	index = 0
	js.lock.Lock()
	defer js.lock.Unlock()
	index, _, ok := js.kv.Min()
	if !ok {
		index = 0
	}
	return
}

func (js *JsonLogStore) GetLog(index uint64, log *raft.Log) (err error) {
	js.lock.Lock()
	defer js.lock.Unlock()
	val, ok := js.kv.Get(index)
	if !ok {
		log = nil
		err = errors.New(fmt.Sprintf("Key:%d not found", index))
	}
	*log = *val
	return
}

func (js *JsonLogStore) StoreLog(log *raft.Log) error {
	js.lock.Lock()
	defer js.lock.Unlock()
	js.kv.Put(log.Index, log)
	return js.save()
}

func (js *JsonLogStore) StoreLogs(logs []*raft.Log) error {
	js.lock.Lock()
	defer js.lock.Unlock()
	for _, log := range logs {
		js.kv.Put(log.Index, log)
	}
	return js.save()
}

func (js *JsonLogStore) DeleteRange(min, max uint64) error {
	if min > max {
		min, max = max, min
	}
	js.lock.Lock()
	defer js.lock.Unlock()
	removed := 0
	for i := min; i <= max; i++ {
		key, _, ok := js.kv.Ceiling(i)
		if ok && key <= max {
			js.kv.Remove(key)
			removed++
		} else {
			break
		}
	}
	if removed > 0 {
		return js.save()
	}
	return nil
}

func (js *JsonLogStore) save() (err error) {
	data, err := json.Marshal(js.kv)
	if err == nil {
		err = os.WriteFile(js.jsonfilepath, data, 0600)
	}
	return
}

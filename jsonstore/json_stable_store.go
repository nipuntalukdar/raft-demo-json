package jsonstore

import (
	"encoding/json"
	"io"
	"os"
	"strconv"
	"sync"
)

type JsonStableStore struct {
	jsonfilepath string
	kv           map[string]string
	lock         sync.Mutex
}

func NewJsonStableStore(jsonfilepath string) (js *JsonStableStore, err error) {
	kv := make(map[string]string)
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
	}
	js = &JsonStableStore{jsonfilepath: jsonfilepath, kv: kv, lock: sync.Mutex{}}
	err = nil
	return
}

func (js *JsonStableStore) Set(key []byte, value []byte) error {
	js.lock.Lock()
	defer js.lock.Unlock()
	js.kv[string(key)] = string(value)
	js.save()
	return nil
}

func (js *JsonStableStore) SetUint64(key []byte, value uint64) error {
	js.lock.Lock()
	defer js.lock.Unlock()
	js.kv[string(key)] = strconv.FormatUint(value, 10)
	js.save()
	return nil
}

func (js *JsonStableStore) Get(key []byte) (value []byte, err error) {
	js.lock.Lock()
	defer js.lock.Unlock()
	value = []byte{}
	val, exists := js.kv[string(key)]
	if !exists {
		err = ErrKeyNotFound
	}
	value = []byte(val)
	return
}

func (js *JsonStableStore) GetUint64(key []byte) (value uint64, err error) {
	js.lock.Lock()
	defer js.lock.Unlock()
	value = 0
	err = nil
	val, exists := js.kv[string(key)]
	if !exists {
		err = ErrKeyNotFound
		return
	}
	value, err = strconv.ParseUint(val, 10, 64)
	if err != nil {
		value = 0
	}
	return
}

func (js *JsonStableStore) save() (err error) {
	data, err := json.Marshal(js.kv)
	if err == nil {
		err = os.WriteFile(js.jsonfilepath, data, 0600)
	}
	return
}

package jsonstore

import (
	"encoding/json"
	"github.com/hashicorp/raft"
	"io"
	"os"
)

func BootstrapConfig(configfile string) (*raft.Configuration, error) {

	var configuration raft.Configuration

	file, err := os.OpenFile(configfile, os.O_RDONLY, 0600)
	if err != nil {
		panic(err)
	}
	data, err := io.ReadAll(file)
	err = json.Unmarshal(data, &configuration.Servers)
	if err != nil {
		panic(err)
	}

	return &configuration, nil

}

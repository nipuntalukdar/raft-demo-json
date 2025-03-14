package jsonstore

import (
	"testing"
)

func TestBootstrap(t *testing.T) {
	configuration, err := BootstrapConfig("../sampleconfig/config.json")
	if err != nil {
		t.Fatal(err)
	}

	t.Log(*configuration)

}

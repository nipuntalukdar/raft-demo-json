package jsonstore

import (
	"testing"
)

func TestBootstrap(t *testing.T) {
	configuration, err := Bootstrap("../sampleconfig/config.json")
	if err != nil {
		t.Fatal(err)
	}

	t.Log(*configuration)

}

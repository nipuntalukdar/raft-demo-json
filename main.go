package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	hclog "github.com/hashicorp/go-hclog"
	"github.com/nipuntalukdar/raftdemojson/jsonstore"
	"github.com/nipuntalukdar/rollingwriter"
)

type Document struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type RequestData struct {
	Data []Document `json:"data"`
}

type Response struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

type kvStore struct {
	rinf *jsonstore.RaftInterface
	logger hclog.Logger
}

func newAdKVHandler(rinf *jsonstore.RaftInterface, logger hclog.Logger) *kvStore {
	return &kvStore{rinf: rinf, logger: logger}
}

func (kv *kvStore) handlePost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var requestData RequestData
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&requestData); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		kv.logger.Error("DecodeError", "Error", err)
		return
	}

	fmt.Println("Received documents:")
	for _, doc := range requestData.Data {
		kv.logger.Debug("Add Data", doc.Key, doc.Value)
		err := kv.rinf.AddKV(doc.Key, doc.Value)
		if err != nil {
			http.Error(w, "Internal Error", http.StatusInternalServerError)
			kv.logger.Error("KeyAadd", "Error", err)
		}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(Response{Status: "success"})
}

func main() {
	// Declare flags
	configFile := flag.String("config", "sampleconfig/config.json", "Path to configuration file")
	logstoreFile := flag.String("logstore", "log/logstore.json", "Path to logstore file")
	stablestoreFile := flag.String("stablestore", "log/stablestore.json", "Path to stablestore file")
	transport := flag.String("transport", "127.0.0.1:7000", "Address to listen on")
	snapshotDrr := flag.String("snapshotdir", "/tmp/snapshot", "Directory for snapshots")
	serverid := flag.String("serverid", "", "Server Id for this server")
	logfileconfig := flag.String("logfileconfig", "sampleconfig/logfile_config.json", "logfileconfig")
	httpserveraddr := flag.String("httpserveraddr", ":8000", "Http server address")

	flag.Parse()
	if *serverid == "" {
		fmt.Println("Server id must be passed ")
		os.Exit(1)
	}

	rollingwr, err := rollingwriter.NewWriterFromConfigFile(*logfileconfig)
	if err != nil {
		panic(err)
	}

	logger := hclog.New(&hclog.LoggerOptions{Name: "RaftDemo", Output: rollingwr,
		Level: hclog.Debug})

	raftin, err := jsonstore.NewRaftInterface(*configFile, *logstoreFile, *stablestoreFile, *snapshotDrr,
		*transport, *serverid, logger, rollingwr)

	if err != nil {
		panic(err)
	}
	time.Sleep(2 * time.Second)
	raftin.Leader()
	addkv := newAdKVHandler(raftin, logger)
	http.HandleFunc("/documents", addkv.handlePost)
	fmt.Println("Server listening on :8080")
	if err := http.ListenAndServe(*httpserveraddr, nil); err != nil {
		logger.Error("Error listening", "Error", err)
	}

	rollingwr.Close()

}

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
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

type HttpListerner struct {
	ID                  string `json:"ID"`
	HttpListenerAddress string `json:"HttpListenerAddress"`
}

type HttpListenerConfig struct {
	HttpListeners []HttpListerner
}

type Response struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

type kvStore struct {
	rinf          *jsonstore.RaftInterface
	logger        hclog.Logger
	httplisteners map[string]string
}

func newAdKVHandler(rinf *jsonstore.RaftInterface, logger hclog.Logger,
	httplisteners map[string]string) *kvStore {
	return &kvStore{rinf: rinf, logger: logger, httplisteners: httplisteners}
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

	kv.logger.Info("Received documents:")
	for _, doc := range requestData.Data {
		kv.logger.Debug("Add Data", doc.Key, doc.Value)
		err := kv.rinf.AddKV(doc.Key, doc.Value)
		if err != nil {
			if err != jsonstore.LeaderDifferent {
				http.Error(w, "Internal Error", http.StatusInternalServerError)
				kv.logger.Error("KeyAadd", "Error", err)
				return
			} else {
				leaderserver, leaderid := kv.rinf.LeaderWithID()
				kv.logger.Info("Different leader", "leader", leaderserver)
				if leaderserver != "" {
					leaderUrl := fmt.Sprintf("http://%s/documents", kv.httplisteners[leaderid])
					w.Header().Set("Location", leaderUrl)
					w.WriteHeader(http.StatusPermanentRedirect)

				} else {
					http.Error(w, "Internal Error", http.StatusInternalServerError)
				}
				return
			}
		}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(Response{Status: "success"})
}

func getHttpListeners(httplisteners string) (*HttpListenerConfig, error) {

	file, err := os.OpenFile(httplisteners, os.O_RDONLY, 0600)
	if err != nil {
		return nil, err
	}
	data, err := io.ReadAll(file)

	var httpconfig HttpListenerConfig
	err = json.Unmarshal(data, &httpconfig.HttpListeners)
	if err != nil {
		return nil, err
	}
	return &httpconfig, err
}

func main() {
	fmt.Println("Make sure that IDs for the http listener config and raft config match ")
	configFile := flag.String("config", "sampleconfig/config.json", "Path to configuration file")
	httpListentconfigFile := flag.String("httplistenerconfig", "sampleconfig/http_config.json",
		"Path to http listener config file")
	logstoreFile := flag.String("logstore", "log/logstore.json", "Path to logstore file")
	stablestoreFile := flag.String("stablestore", "log/stablestore.json", "Path to stablestore file")
	transport := flag.String("transport", "127.0.0.1:7000", "Address to listen on")
	snapshotDrr := flag.String("snapshotdir", "/tmp/snapshot", "Directory for snapshots")
	serverid := flag.String("serverid", "", "Server Id for this server")
	logfileconfig := flag.String("logfileconfig", "sampleconfig/logfile_config.json", "logfileconfig")

	flag.Parse()
	if *serverid == "" {
		fmt.Println("Server id must be passed ")
		os.Exit(1)
	}

	httpconfig, err := getHttpListeners(*httpListentconfigFile)
	if err != nil {
		panic(err)
	}
	http_listeners := make(map[string]string)
	for _, listener := range httpconfig.HttpListeners {
		http_listeners[listener.ID] = listener.HttpListenerAddress
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
	addkv := newAdKVHandler(raftin, logger, http_listeners)
	http.HandleFunc("/documents", addkv.handlePost)
	logger.Info("Server started", "raft-address", transport, "http-listener", http_listeners[*serverid])
	if err := http.ListenAndServe(http_listeners[*serverid], nil); err != nil {
		logger.Error("Error listening", "Error", err)
	}

	rollingwr.Close()

}

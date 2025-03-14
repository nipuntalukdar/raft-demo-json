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

type RequestKeys struct {
	Keys []string `json:"keys"`
}

type Servers struct {
	Servers []jsonstore.Server `json:"servers"`
}

type Response struct {
	Status     string             `json:"status"`
	Message    string             `json:"message,omitempty"`
	DeleteKeys []string           `json:"deleted,omitempty"`
	NotFound   []string           `json:"notfound,omitempty"`
	FoundKeys  map[string]string  `json:"found,omitempty"`
	Servers    []jsonstore.Server `json:"servers,omitempty"`
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

func (kv *kvStore) deleteKeys(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(Response{Status: "failed", Message: "Method not allowed"})
		return
	}

	var req RequestKeys
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Response{Status: "failed", Message: "Bad request body"})
		return
	}

	deletedKeys := []string{}
	notFoundKeys := []string{}
	baderr := err
	for _, key := range req.Keys {

		err = kv.rinf.Delete(key)
		if err != nil {
			if err == jsonstore.ErrKeyNotFound {
				notFoundKeys = append(notFoundKeys, key)
			} else if err == jsonstore.LeaderDifferent {
				leaderserver, leaderid := kv.rinf.LeaderWithID()
				kv.logger.Info("Different leader", "leader", leaderserver)
				if leaderserver != "" {
					leaderUrl := fmt.Sprintf("http://%s/delete", kv.httplisteners[leaderid])
					w.Header().Set("Location", leaderUrl)
					w.WriteHeader(http.StatusPermanentRedirect)

				} else {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusInternalServerError)
					json.NewEncoder(w).Encode(Response{Status: "failed", Message: "Leader not found"})
				}
				return

			} else {
				baderr = err
			}
		} else {
			deletedKeys = append(deletedKeys, key)
		}
	}
	response := Response{}
	if len(notFoundKeys) > 0 {
		response.NotFound = notFoundKeys
	}

	if len(deletedKeys) > 0 {
		response.DeleteKeys = deletedKeys
		w.WriteHeader(http.StatusOK)
		response.Status = "success"
	} else {
		response.Status = "failed"
		w.WriteHeader(http.StatusNotFound)
	}
	if baderr != nil {
		response.Status = "failed"
		w.WriteHeader(http.StatusInternalServerError)
		response.Message = baderr.Error()
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
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

	kv.logger.Info("Received keyvals:")
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
					leaderUrl := fmt.Sprintf("http://%s/keyvals", kv.httplisteners[leaderid])
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

func (kv *kvStore) getServers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	kv.logger.Info("Getting servers")

	servers, err := kv.rinf.GetServers()

	w.Header().Set("Content-Type", "application/json")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Response{Status: "failure"})
	} else {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(Response{Status: "success", Servers: servers})
	}
}

func (kv *kvStore) testPersist(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	kv.logger.Info("Persisting snapshot")
	kv.rinf.Persist()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(Response{Status: "success"})
}

func (kv *kvStore) getKeys(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(Response{Status: "failed", Message: "Method not allowed"})
		return
	}

	var req RequestKeys
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Response{Status: "failed", Message: "Bad request body"})
		return
	}

	foundkeys := make(map[string]string)
	notFoundKeys := []string{}
	baderr := err
	for _, key := range req.Keys {
		value, err := kv.rinf.Get(key)
		if err != nil {
			if err == jsonstore.ErrKeyNotFound {
				notFoundKeys = append(notFoundKeys, key)
			} else {
				baderr = err
			}
		} else {
			foundkeys[key] = value
		}
	}
	response := Response{}
	if len(notFoundKeys) > 0 {
		response.NotFound = notFoundKeys
	}

	if len(foundkeys) > 0 {
		response.FoundKeys = foundkeys
		w.WriteHeader(http.StatusOK)
		response.Status = "success"
	} else {
		response.Status = "failed"
		w.WriteHeader(http.StatusNotFound)
	}
	if baderr != nil {
		response.Status = "failed"
		w.WriteHeader(http.StatusInternalServerError)
		response.Message = baderr.Error()
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
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
	http.HandleFunc("/keyvals", addkv.handlePost)
	http.HandleFunc("/delete", addkv.deleteKeys)
	http.HandleFunc("/testpersist", addkv.testPersist)
	http.HandleFunc("/getkeys", addkv.getKeys)
	http.HandleFunc("/servers", addkv.getServers)

	logger.Info("Server started", "raft-address", transport, "http-listener", http_listeners[*serverid])
	if err := http.ListenAndServe(http_listeners[*serverid], nil); err != nil {
		logger.Error("Error listening", "Error", err)
	}

	rollingwr.Close()

}

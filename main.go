package main

import (
	"flag"
	"fmt"
	"os"

	hclog "github.com/hashicorp/go-hclog"
	"github.com/nipuntalukdar/raftdemojson/jsonstore"
	"github.com/nipuntalukdar/rollingwriter"
)

func main() {
	// Declare flags
	configFile := flag.String("config", "sampleconfig/config.json", "Path to configuration file")
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

	rollingwr, err := rollingwriter.NewWriterFromConfigFile(*logfileconfig)
	if err != nil {
		panic(err)
	}

	logger := hclog.New(&hclog.LoggerOptions{Name: "RaftDemo", Output: rollingwr,
		Level: hclog.Debug})

	_, err = jsonstore.NewRaftInterface(*configFile, *logstoreFile, *stablestoreFile, *snapshotDrr,
		*transport, *serverid, &logger, rollingwr)

	if err != nil {
		panic(err)
	}
	logger.Info("Good bye")
	rollingwr.Close()

}

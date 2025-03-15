# A replicated key-value store demo using RAFT consensus algorithm   

This application demonstrates a HA key-value store using RAFT algorithm. It uses the hashicorp [Raft](https://github.com/hashicorp/raft) library.
The stable store for configuration, the store for RAFT logs and snapshot store for the key-values are all JSON file based and hence they can be inspected easily for understanding how the data is laid out on them. 

The demo application Provides the below REST APIs:
   * /keyvavals
     * This  is to add a a list of key-values to the key-value store
   * /delete
     * It deletes a list of keys from the key-value store
   * /getkeys
     * It retrieves the values for a list of keys
   * /testpersist
     * It triggers a snapshotting of the key-values on all the nodes
   * /servers
     * It gets the current list of servers along with with their ids, nd whether a server is leader or not

## How to use this
* clone the repository, and execute the below commands.
```bash
cd raft-demo-json
# Build the demo binary
go get
go build
```
Above command builds the raftdemojson binary. Check its usage:
```bash
./raftdemojson --help
```
It outputs the below details:
```
Make sure that IDs for the http listener config and raft config match 
Usage of ./raftdemojson:
  -config string
    	Path to configuration file (default "sampleconfig/config.json")
  -httplistenerconfig string
    	Path to http listener config file (default "sampleconfig/http_config.json")
  -logfileconfig string
    	logfileconfig (default "sampleconfig/logfile_config.json")
  -logstore string
    	Path to logstore file (default "log/logstore.json")
  -serverid string
    	Server Id for this server
  -snapshotdir string
    	Directory for snapshots (default "/tmp/snapshot")
  -stablestore string
    	Path to stablestore file (default "log/stablestore.json")
  -transport string
    	Address to listen on (default "127.0.0.1:7000")

```
Make the environment for three node cluster locally from the configs in sampleconfig directory
```bash
bash mkenv.sh
```
Above commands creates a three directory t1, t2 and t3 and copies the raftdemojson binary and the configs to this directory. Check content of t1 directory.
```
tree t1
```
It will be something like:
```
$ tree t1
t1
├── raftdemojson
└── sampleconfig
    ├── config.json
    ├── http_config.json
    └── logfile_config.json
```
Now let us start the cluster
```bash
bash run.sh
```
It will print the below lines, one line from each servers.
```
Make sure that IDs for the http listener config and raft config match 
Make sure that IDs for the http listener config and raft config match 
Make sure that IDs for the http listener config and raft config match
```
The replicated HA key-value store should be up an running now. Let us check the list of servers in the cluster:
```bash
curl  -s http://127.0.0.1:8000/servers  | jq
```
It will return a list of setrver be something like:
```
{
  "status": "success",
  "servers": [
    {
      "Address": "127.0.0.1:7000",
      "Id": "id1",
      "Leader": false
    },
    {
      "Address": "127.0.0.1:7001",
      "Id": "id2",
      "Leader": false
    },
    {
      "Address": "127.0.0.1:7002",
      "Id": "id3",
      "Leader": true
    }
  ]
}
```
Everything looks good now:
Let us add few key-value pairs:
```bash
bash add_key_vals.sh
```
Now, let us try getting the server list again and do some delete and get of keys.
```Bash
bash    get_and_delete.sh
```
Output will be something like:
```
List of servers
{"status":"success","servers":[{"Address":"127.0.0.1:7000","Id":"id1","Leader":false},{"Address":"127.0.0.1:7001","Id":"id2","Leader":false},{"Address":"127.0.0.1:7002","Id":"id3","Leader":true}]}
Delete
{"status":"success","deleted":["bDEF139"],"notfound":["ABC","he","ram"]}
Get
{"status":"success","notfound":["bDEF139","he","ram"],"found":{"bDEF137":"v9999137","bDEF138":"v9999138"}}
```

## Few API call examples

To force a snapshot, issue the below command (It is not really necessary as snapshots are automatically triggerred periodically)
```bash
curl   -L  -X POST  http://localhost:8001/testpersist
```
Key-value add API example:
```bash
curl -L -X POST -H "Content-Type: application/json" -d '{"data": [{"key": "akey", "value": "some value"}, {"key": "anotherkey", "value": "another value"}]}' http://localhost:8000/keyvals
```
Delete key API example:
```bash
curl -L  -X DELETE -H "Content-Type: application/json" -d '{"keys": ["akey", "helli", "hi", "what"]}' http://localhost:8000/delete
```
Get key API example:
```bash
curl  -XGET -H "Content-Type: application/json" -d '{"keys": ["bDEF139", "bDEF138", "when", "How", "bDEF137"]}' http://localhost:8000/getkeys
```
Get the server list:
```Bash
curl  http://localhost:8000/servers
```


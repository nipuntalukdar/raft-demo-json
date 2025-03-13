cd t1
./raftdemojson -transport "127.0.0.1:7000"  --serverid 123 -snapshotdir snap -httpserveraddr ":8000" &
cd ../t2
./raftdemojson -transport "127.0.0.1:7001"  --serverid 125 -snapshotdir snap -httpserveraddr ":8001" &
cd ../t3
./raftdemojson -transport "127.0.0.1:7002"  --serverid 126 -snapshotdir snap -httpserveraddr ":8002" &

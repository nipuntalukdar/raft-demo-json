cd t1
./raftdemojson -transport "127.0.0.1:7000" --serverid id1 -snapshotdir snap &
cd ../t2
./raftdemojson -transport "127.0.0.1:7001" --serverid id2 -snapshotdir snap &
cd ../t3
./raftdemojson -transport "127.0.0.1:7002" --serverid id3 -snapshotdir snap &

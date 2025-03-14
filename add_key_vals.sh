x=1
while [ $x -lt 400 ]; do
    curl -L -X POST -H "Content-Type: application/json" -d '{"data": [{"key": "'"bbkey_$x"'", "value": "'"$x"'"}, {"key": "'"bDEF$x"'", "value": "'"v9999$x"'"}]}' http://localhost:8002/keyvals
    x=$((x + 1))
done

echo "List of servers"
curl http://127.0.0.1:8000/servers
echo "Delete"
curl -L  -X DELETE -H "Content-Type: application/json" -d '{"keys": ["bDEF139", "ABC", "he", "ram"]}' http://127.0.0.1:8000/delete
echo "Get"
curl  -XGET -H "Content-Type: application/json" -d '{"keys": ["bDEF139", "bDEF138", "he", "ram", "bDEF137"]}' http://127.0.0.1:8001/getkeys

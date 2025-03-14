curl -L -v -X DELETE -H "Content-Type: application/json" -d '{"keys": ["bDEF139", "ABC", "he", "ram"]}' http://127.0.0.1:8000/delete
curl  -XGET -H "Content-Type: application/json" -d '{"keys": ["bDEF139", "bDEF138", "he", "ram", "bDEF137"]}' http://127.0.0.1:8001/getkeys

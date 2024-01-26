#!/bin/bash
trap "rm server;kill 0" EXIT

go build -o server
./server -port=8001 &
./server -port=8002 &
./server -port=8003 -api=1 &

sleep 2
echo ">>> start test"
echo ">>> curl 1"
curl "http://localhost:9999/api?key=Tom" &
echo ">>> curl 2"
curl "http://localhost:9999/api?key=Tom" &

sleep 2
echo ">>>curl 3"
curl "http://localhost:9999/api?key=Tom" &

sleep 2
echo "test hot cache"
echo ">>>curl 4"
curl "http://localhost:9999/api?key=Tom" &

wait
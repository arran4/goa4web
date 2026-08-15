#!/bin/bash
go build -o goa4web ./cmd/goa4web/
./goa4web serve -db-driver sqlite3 &
SERVER_PID=$!
sleep 3
curl -s http://localhost:8080/private/topic/1/unread
kill $SERVER_PID

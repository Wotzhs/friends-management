#!/bin/sh
migrate -path ./migrations -database postgres://postgres@db/friends_management_test?sslmode=disable drop
migrate -path ./migrations -database postgres://postgres@db/friends_management_test?sslmode=disable up
sh scripts/start.sh &
sleep 1
go test *test.go
echo "test completed, exiting now"
exit
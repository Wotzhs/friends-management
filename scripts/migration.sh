#!/bin/sh
migrate -path ./migrations -database postgres://postgres@db/friends_management?sslmode=disable up
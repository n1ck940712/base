#!/bin/sh

echo "Building please wait..."
go build -o main cmd/main/main.go
echo "Building completed..."
./main

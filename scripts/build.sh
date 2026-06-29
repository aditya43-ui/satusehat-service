#!/bin/bash
echo "Building application..."
go build -o bin/service cmd/api/main.go

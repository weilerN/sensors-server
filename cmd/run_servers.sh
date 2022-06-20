#!/bin/bash


if [[ ! -z "$1" ]] && [[ "$1" == "-v" ]]; then
  echo "Running 3 server instances with debug mode"
  go run cmd/server.go -port=50051 -v &
  go run cmd/server.go -port=50052 -v &
  go run cmd/server.go -port=50053 -v &
else
  echo "Running 3 server instances"
  go run cmd/server.go -port=50051 &
  go run cmd/server.go -port=50052 &
  go run cmd/server.go -port=50053 &
fi


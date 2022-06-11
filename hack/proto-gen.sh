#!/bin/bash
SRC="api/proto/v1"
OUT=pkg/api/v1

if [ ! -d "$OUT" ]
  then mkdir -p "$OUT"
fi

protoc --proto_path=$SRC --go_out=$OUT --go_opt=paths=source_relative --go-grpc_out=$OUT  --go-grpc_opt=paths=source_relative bookmark-service.proto


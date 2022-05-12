#!/usr/bin/env bash

set -e

# If either of the commands fail below, run the following and then retry:
#
# go get google.golang.org/protobuf/cmd/protoc-gen-go@v1.26
# go get google.golang.org/grpc/cmd/protoc-gen-go-grpc@1.39.0

if ! hash go; then
  echo "please install go and try again"
  exit 1
fi
if ! hash protoc-gen-go; then
  echo "installing protoc-gen-go"
  go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.28
fi
if ! hash protoc-gen-go-grpc; then
  echo "installing protoc-gen-go-grpc"
  go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2
fi

cd "$(dirname "$0")"
rm -rf gen internal/ingester
mkdir -p gen internal/ingester

git clone \
  --depth 1 \
  https://github.com/influxdata/influxdb_iox \
  gen

filenames=$(find gen/generated_types/protos/influxdata/iox/ingester -name '*.proto')
protoc \
  --proto_path=gen/generated_types/protos/ \
  --plugin protoc-gen-go="$(which protoc-gen-go)" \
  --plugin protoc-gen-go-grpc="$(which protoc-gen-go-grpc)" \
  --go_out gen \
  --go-grpc_out gen \
  ${filenames}

mv gen/github.com/influxdata/iox/ingester/v1/*.go internal/ingester/

rm -rf gen

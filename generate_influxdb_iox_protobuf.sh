#!/usr/bin/env sh

set -e

if [ -z "$(which protoc-gen-go)" ]; then
  echo "installing protoc-gen-go"
  go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.26
fi
if [ -z "$(which protoc-gen-go-grpc)" ]; then
  echo "installing protoc-gen-go-grpc"
  go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@1.39.0
fi

cd $(dirname "$0")
rm -rf gen internal/management
mkdir -p gen internal/management

git clone \
  --depth 1 \
  https://github.com/influxdata/influxdb_iox \
  gen

filenames=$(find gen/generated_types/protos/influxdata/iox/management -name '*.proto')
protoc \
  --proto_path=gen/generated_types/protos/ \
  --plugin protoc-gen-go="$(which protoc-gen-go)" \
  --plugin protoc-gen-go-grpc="$(which protoc-gen-go-grpc)" \
  --go_out gen \
  --go-grpc_out gen \
  ${filenames}

mv gen/github.com/influxdata/iox/management/v1/*.go internal/management/

rm -rf gen

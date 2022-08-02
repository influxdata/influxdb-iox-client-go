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

# Generate Go files for the specified IOx service.
#
# $1: the IOx proto directory name to generate code for
#
# Generates directories by name from
#  https://github.com/influxdata/influxdb_iox/tree/main/generated_types/protos/influxdata/iox
function generate_service {
  filenames=$(find "gen/generated_types/protos/influxdata/iox/${1}" -name '*.proto')
  protoc \
    --proto_path=gen/generated_types/protos/ \
    --plugin protoc-gen-go="$(which protoc-gen-go)" \
    --plugin protoc-gen-go-grpc="$(which protoc-gen-go-grpc)" \
    --go_out gen \
    --go-grpc_out gen \
    ${filenames}

  mkdir "internal/${1}" 2>/dev/null || true
  mv gen/github.com/influxdata/iox/"${1}"/v1/*.go "internal/${1}/"
  
  echo "generated ${1}"
}

generate_service "ingester"
generate_service "schema"

rm -rf gen

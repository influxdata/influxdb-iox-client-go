#!/usr/bin/env bash

set -e

cd "$(dirname "$0")"
BASEDIR=$(pwd)

for package in . ./ioxsql; do
  echo checking ${package}
  cd ${BASEDIR}/${package}
  if ! go build; then
    fail=1
  fi
  if [[ -n $(gofmt -s -l . | head -n 1) ]]; then
    fail=1
    gofmt -s -d .
  fi
  if ! go vet; then
    fail=1
  fi
  if ! staticcheck -f stylish; then
    fail=1
  fi
done

echo

if [ -n "$fail" ]; then
  echo "at least one check failed"
  exit 1
else
  echo "all checks OK"
fi

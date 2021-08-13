#!/bin/sh

set -e

cd "$(dirname "$0")"
BASEDIR=$(pwd)

if ! go vet ./... ; then
  fail=1
fi

unformatted=$(gofmt -s -l -e "${BASEDIR}")
if [ ! -z "$unformatted" ] ; then
  for filename in $unformatted ; do
    gofmt -s -d "$filename"
  done
  fail=1
fi

if [ -n "$fail" ] ; then
  exit 1
fi

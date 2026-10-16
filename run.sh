#!/bin/sh
set -eu
cd "$(dirname "$0")"
if [ "${1:-}" = "--real" ]; then
  shift
  go run ./cmd/bcie -download "$@"
else
  go run ./cmd/bcie -arrests data/arrests_sample.csv -requests data/requests_sample.csv "$@"
fi

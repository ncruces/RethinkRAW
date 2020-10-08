#!/bin/sh
set -euo pipefail

tgt=RethinkRAW.app/Contents/MacOS

go clean
go generate
go build -tags memfs -ldflags "-s -w" -o "$tgt/rethinkraw"
go mod tidy

rm -rf "$tgt/data" log_*.txt
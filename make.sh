#!/bin/sh
set -eo pipefail

tgt="RethinkRAW.app/Contents/Resources/RethinkRAW.app/Contents/MacOS"
exe="$tgt/rethinkraw"
dat="$tgt/data"

if [[ "$1" == test ]]; then
    shift
    echo Test build...
    go build -race -o "$exe"
    exec "$exe" "$@"
else
    echo Release build...
    go clean
    go generate
    go build -tags memfs -ldflags "-s -w" -o "$exe"
    go mod tidy
    rm -rf "$dat"
fi
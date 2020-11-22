#!/bin/sh
set -eo pipefail

tgt="RethinkRAW.app/Contents/Resources/RethinkRAW.app/Contents/MacOS"
exe="$tgt/rethinkraw"
dat="$tgt/data"

if [ ! -f "$tgt/utils/exiftool/exiftool" ]; then
    echo Download ExifTool...
    url="https://github.com/ncruces/go-exiftool/releases/download/dist/exiftool_unix.tgz"
    curl -L "$url" 2> /dev/null | tar xz -C "$tgt/utils"
fi

if [[ "$1" == test ]]; then
    echo Test build...
    go build -race -o "$exe"
    shift && exec "$exe" "$@"
else
    echo Release build...
    go clean
    go generate
    go build -tags memfs -ldflags "-s -w" -o "$exe"
    go mod tidy
    rm -rf "$dat"
fi
#!/usr/bin/env bash
set -eo pipefail

cd -P -- "$(dirname -- "$0")"

tgt="RethinkRAW"

mkdir -p "$tgt/utils"
cp "build/exiftool_config.pl" "$tgt/utils"
ln -sf rethinkraw "$tgt/rethinkraw-server"

if [ ! -f "$tgt/utils/exiftool/exiftool" ]; then
    echo Download ExifTool...
    url="https://github.com/ncruces/go-exiftool/releases/download/dist/exiftool_unix.tgz"
    curl -sL "$url" | tar xz -C "$tgt/utils"
fi

if [ ! -f "$tgt/utils/dcraw.wasm" ]; then
    echo Download dcraw...
    url="https://github.com/ncruces/dcraw/releases/download/v9.28.7-wasm/dcraw.wasm.gz"
    curl -sL "$url" | gzip -dc > "$tgt/utils/dcraw.wasm"
    cp "$tgt/utils/dcraw.wasm" "pkg/dcraw/embed/dcraw.wasm"
fi

if [ ! -f "assets/normalize.css" ]; then
    echo Download normalize.css...
    curl -sL "https://unpkg.com/@csstools/normalize.css" > assets/normalize.css
fi

if [ ! -f "assets/dialog-polyfill.js" ]; then
    echo Download dialog-polyfill...
    curl -sL "https://unpkg.com/dialog-polyfill@0.5/dist/dialog-polyfill.js"  > assets/dialog-polyfill.js
    curl -sL "https://unpkg.com/dialog-polyfill@0.5/dist/dialog-polyfill.css" > assets/dialog-polyfill.css
fi

if [ ! -f "assets/fontawesome.css" ]; then
    echo Download Font Awesome...
    curl -sL "https://unpkg.com/@fortawesome/fontawesome-free@5.x/css/fontawesome.css"         > assets/fontawesome.css
    curl -sL "https://unpkg.com/@fortawesome/fontawesome-free@5.x/webfonts/fa-solid-900.woff2" > assets/fa-solid-900.woff2
fi

if [[ "$1" == test ]]; then
    echo Run tests...
    go test ./...
elif [[ "$1" == run ]]; then
    echo Run app...
    go build -race -o "$tgt/rethinkraw" && shift && exec "$tgt/rethinkraw" "$@"
elif [[ "$1" == serve ]]; then
    echo Run server...
    go build -race -o "$tgt/rethinkraw" && shift && exec "$tgt/rethinkraw-server" "$@"
elif [[ "$1" == install ]]; then
    echo Build installer...
    rm -rf "$tgt/data"
    export COPYFILE_DISABLE=1
    cp "assets/favicon-192.png" "$tgt/icon.png"
    tar cfj RethinkRAW.tbz "$tgt"
else
    echo Build release...
    export CGO_ENABLED=0
    go clean
    GOOS= GOARCH= go generate
    go build -tags memfs -ldflags "-s -w" -trimpath -o "$tgt/rethinkraw"
fi
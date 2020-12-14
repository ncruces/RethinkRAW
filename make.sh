#!/bin/sh
set -eo pipefail

cd -P -- "$(dirname -- "$0")"

tgt="RethinkRAW.app/Contents/Resources/RethinkRAW.app/Contents/MacOS"
exe="$tgt/rethinkraw"
dat="$tgt/data"

if [ ! -f "$tgt/utils/exiftool/exiftool" ]; then
    echo Download ExifTool...
    url="https://github.com/ncruces/go-exiftool/releases/download/dist/exiftool_unix.tgz"
    curl -sL "$url" | tar xz -C "$tgt/utils"
fi

if [ ! -f "assets/normalize.css" ]; then
    echo Download normalize.css...
    curl -sL "https://unpkg.com/normalize.css" > assets/normalize.css
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
    echo Test build...
    go build -race -o "$exe"
    shift && exec "$exe" "$@"
else
    echo Release build...
    osacompile -l JavaScript -o RethinkRAW.app darwin.js
    go clean
    go generate
    go build -tags memfs -ldflags "-s -w" -o "$exe"
    go mod tidy
    rm -rf "$dat"
fi
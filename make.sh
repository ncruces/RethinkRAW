#!/bin/sh
set -eo pipefail

cd -P -- "$(dirname -- "$0")"

tgt="RethinkRAW.app/Contents/Resources/RethinkRAW.app/Contents/MacOS"
srv="RethinkRAW.app/Contents/Resources/rethinkraw-server"

if [ ! -f "$tgt/utils/exiftool/exiftool" ]; then
    echo Download ExifTool...
    url="https://github.com/ncruces/go-exiftool/releases/download/dist/exiftool_unix.tgz"
    curl -sL "$url" | tar xz -C "$tgt/utils"
fi

if [ ! -f "$tgt/utils/dcraw" ]; then
    echo Download dcraw...
    url="https://github.com/ncruces/dcraw/releases/download/v9.28.2/dcraw.gz"
    curl -sL "$url" | gzcat > "$tgt/utils/dcraw" && chmod +x "$tgt/utils/dcraw"
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
    go build -race -o "$tgt/rethinkraw" && shift && exec "$srv" "$@"
elif [[ "$1" == install ]]; then
    echo Build installer...
    rm -rf "$tgt/data"
    tmp="$(mktemp -d)"
    ln -s /Applications "$tmp"
    cp -r RethinkRAW.app "$tmp"
    hdiutil create -volname RethinkRAW -srcfolder "$tmp" -format UDBZ -ov RethinkRAW.dmg
else
    echo Build release...
    osacompile -l JavaScript -o RethinkRAW.app build/droplet.js
    tmp="$(mktemp -d)"
    CGO_ENABLED=0
    go clean
    go generate
    GOARCH=amd64 go build -tags memfs -ldflags "-s -w" -trimpath -o "$tmp/rethinkraw_x64"
    GOARCH=arm64 go build -tags memfs -ldflags "-s -w" -trimpath -o "$tmp/rethinkraw_arm"
    go run github.com/randall77/makefat "$tgt/rethinkraw" "$tmp/rethinkraw_x64" "$tmp/rethinkraw_arm"
fi
#!/usr/bin/env bash
set -eo pipefail

cd -P -- "$(dirname -- "$0")"

app="RethinkRAW.app/Contents/Resources"
tgt="RethinkRAW.app/Contents/Resources/RethinkRAW.app/Contents"

osacompile -l JavaScript -o RethinkRAW.app build/droplet.js
mkdir -p "$tgt/Resources"
mkdir -p "$tgt/MacOS/utils"
cp "build/app.plist" "$tgt/Info.plist"
cp "build/icon.icns" "$tgt/Resources/"
cp "build/icon.icns" "$app/droplet.icns"
cp "build/exiftool_config.pl" "$tgt/MacOS/utils"
plutil -replace CFBundleVersion -string "0.10.5" RethinkRAW.app/Contents/Info.plist
plutil -replace CFBundleDocumentTypes -xml "$(cat build/doctypes.plist)" RethinkRAW.app/Contents/Info.plist
ln -sf "RethinkRAW.app/Contents/MacOS/rethinkraw" "$app/rethinkraw-server"

if [ ! -f "$tgt/MacOS/utils/exiftool/exiftool" ]; then
    echo Download ExifTool...
    url="https://github.com/ncruces/go-exiftool/releases/download/dist/exiftool_unix.tgz"
    curl -sL "$url" | tar xz -C "$tgt/MacOS/utils"
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
    go build -race -o "$tgt/MacOS/rethinkraw" && shift && exec "$tgt/MacOS/rethinkraw" "$@"
elif [[ "$1" == serve ]]; then
    echo Run server...
    go build -race -o "$tgt/MacOS/rethinkraw" && shift && exec "$app/rethinkraw-server" "$@"
elif [[ "$1" == install ]]; then
    echo Build installer...
    rm -rf "$tgt/MacOS/data"
    tmp="$(mktemp -d)"
    ln -s /Applications "$tmp"
    cp -r RethinkRAW.app "$tmp"
    hdiutil create -volname RethinkRAW -srcfolder "$tmp" -format UDBZ -ov RethinkRAW.dmg
else
    echo Build release...
    tmp="$(mktemp -d)"
    export CGO_ENABLED=0
    export GOOS=darwin
    go clean
    GOOS= GOARCH= go generate
    GOARCH=amd64 go build -tags memfs -ldflags "-s -w" -trimpath -o "$tmp/rethinkraw_x64"
    GOARCH=arm64 go build -tags memfs -ldflags "-s -w" -trimpath -o "$tmp/rethinkraw_arm"
    lipo -create "$tmp/rethinkraw_x64" "$tmp/rethinkraw_arm" -output "$tgt/MacOS/rethinkraw"
fi
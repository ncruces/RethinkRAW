@ECHO OFF
SETLOCAL EnableDelayedExpansion

CD /D "%~dp0"

SET "tgt=RethinkRAW"
SET "exe=%tgt%\RethinkRAW.exe"
SET "dat=%tgt%\data"

IF NOT EXIST %tgt%\utils\exiftool\exiftool.exe (
    ECHO Download ExifTool...
    SET "url=https://github.com/ncruces/go-exiftool/releases/download/dist/exiftool_windows.zip"
    go run github.com/ncruces/go-fetch -unpack !url! %tgt%\utils
)

IF NOT EXIST %tgt%\utils\dcraw.exe (
    ECHO Download dcraw...
    SET "url=https://github.com/ncruces/dcraw/releases/download/v9.28.2-win/dcraw.zip"
    go run github.com/ncruces/go-fetch -unpack !url! %tgt%\utils
)

IF NOT EXIST assets\normalize.css (
    ECHO Download normalize.css...
    go run github.com/ncruces/go-fetch "https://unpkg.com/normalize.css" assets\normalize.css
)

IF NOT EXIST assets\dialog-polyfill.js (
    ECHO Download dialog-polyfill...
    go run github.com/ncruces/go-fetch "https://unpkg.com/dialog-polyfill@0.5/dist/dialog-polyfill.js"  assets\dialog-polyfill.js
    go run github.com/ncruces/go-fetch "https://unpkg.com/dialog-polyfill@0.5/dist/dialog-polyfill.css" assets\dialog-polyfill.css
)

IF NOT EXIST assets\fontawesome.css (
    ECHO Download Font Awesome...
    go run github.com/ncruces/go-fetch "https://unpkg.com/@fortawesome/fontawesome-free@5.x/css/fontawesome.css"         assets\fontawesome.css
    go run github.com/ncruces/go-fetch "https://unpkg.com/@fortawesome/fontawesome-free@5.x/webfonts/fa-solid-900.woff2" assets\fa-solid-900.woff2
)

IF [%1]==[test] (
    ECHO Test build...
    go build -race -o %exe% && %exe%
) ELSE (
    ECHO Release build...
    go run github.com/josephspurrier/goversioninfo/cmd/goversioninfo -64 -icon=assets/favicon.ico -manifest=build/app.manifest build/versioninfo.json
    SET CGO_ENABLED=0
    go clean
    go generate
    go build -tags memfs -ldflags "-s -w" -o %exe%
    go mod tidy
    IF EXIST %dat% RD /S /Q %dat%
)
@ECHO OFF
SETLOCAL

CD /D "%~dp0"

SET "tgt=RethinkRAW"
SET "exe=%tgt%\RethinkRAW.exe"
SET "dat=%tgt%\data"

IF NOT EXIST %tgt%\utils\exiftool\exiftool (
    ECHO Download ExifTool...
    SET "url=https://github.com/ncruces/go-exiftool/releases/download/dist/exiftool_windows.zip"
    go run github.com/ncruces/go-fetch -unpack %url% %tgt%\utils
)

IF [%1]==[test] (
    ECHO Test build...
    go build -race -o %exe% || EXIT /B
    %exe%
) ELSE (
    ECHO Release build...
    go clean || EXIT /B
    go generate || EXIT /B
    go build -tags memfs -ldflags "-s -w" -o %exe% || EXIT /B
    go mod tidy || EXIT /B
    IF EXIST %dat% RD /S /Q %dat%
)
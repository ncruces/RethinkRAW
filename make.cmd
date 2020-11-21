@ECHO OFF


SET tgt="RethinkRAW"
SET exe="%tgt%\RethinkRAW.exe"
SET dat="%tgt%\data"

IF [%1]==[test] (
    SHIFT
    ECHO Test build...
    go build -race -o %exe% || EXIT /B
    %exe% %*
) ELSE (
    ECHO Release build...
    go clean || EXIT /B
    go generate || EXIT /B
    go build -tags memfs -ldflags "-s -w" -o %exe%  || EXIT /B
    go mod tidy || EXIT /B
    IF EXIST %dat% RD /S /Q %dat%
)
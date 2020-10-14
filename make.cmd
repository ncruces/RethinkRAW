@ECHO OFF

IF [%1]==[test] (
    ECHO Test build...
    go build -o RethinkRAW\RethinkRAW.exe || EXIT /B
    RethinkRAW\RethinkRAW.exe
) ELSE (
    ECHO Release build...
    go clean || EXIT /B
    go generate || EXIT /B
    go build -tags memfs -ldflags "-s -w" -o RethinkRAW\RethinkRAW.exe || EXIT /B
    go mod tidy || EXIT /B
    IF EXIST RethinkRAW\data RD /S /Q RethinkRAW\data
)
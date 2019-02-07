@ECHO OFF

IF [%1]==[test] (
    ECHO Test build...
    go build && RethinkRAW.exe
) ELSE (
    ECHO Release build...
    go generate && go build -tags imbed -ldflags -s
)
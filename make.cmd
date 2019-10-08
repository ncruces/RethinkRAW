@ECHO OFF

IF [%1]==[test] (
    ECHO Test build...
    go build
) ELSE (
    ECHO Release build...
    go clean && go generate && go build -tags memfs -ldflags -s
)
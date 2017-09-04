# Go Bob Status

Sync Jenkins build status with Github commit written in Golang

## Cli app

Type --help to see all variables that needs to be set


## Compile binary for Linux

CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o go-bob-status .
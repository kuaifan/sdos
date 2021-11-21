#!/bin/bash

COMMIT_SHA1=$(git rev-parse --short HEAD || echo "0.0.0")
BUILD_TIME=$(date "+%F %T")
goldflags="-s -w -X 'main.CommitSha1=${COMMIT_SHA1}' -X 'main.BuildTime=${BUILD_TIME}'"

if [ "$1" = "linux" ]; then
    env GOOS=linux GOARCH=amd64 go build -o sdos -ldflags "$goldflags" main.go && command -v upx &> /dev/null && upx sdos
else
    env CGO_ENABLED=0 go build -o sdos -ldflags "$goldflags" main.go && command -v upx &> /dev/null && upx sdos
fi

exit 0
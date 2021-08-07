export PATH := $(GOPATH)/bin:$(PATH)
export GO111MODULE=on

COMMIT_SHA1   	:= $(shell git rev-parse --short HEAD || echo "0.0.0")
BUILD_TIME      := $(shell date "+%F %T")

LDFLAGS := -s -w -X 'main.CommitSha1=${COMMIT_SHA1}' -X 'main.BuildTime=${BUILD_TIME}'

all:
	env CGO_ENABLED=0 go build -trimpath -ldflags "$(LDFLAGS)"

clean:
	rm -f sdos
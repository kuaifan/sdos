export PATH := $(GOPATH)/bin:$(PATH)
export GO111MODULE=on

V = 0
Q = $(if $(filter 1,$V),,@)
M = $(shell printf "\033[34;1m▶\033[0m")

GO = GOGC=off go
MODULE = $(shell env GO111MODULE=on $(GO) list -m)

VERSION			:= $(shell git describe --tags --always --match=v* 2> /dev/null || cat $(CURDIR)/.version 2> /dev/null || echo v0)
VERSION_HASH	:= $(shell git rev-parse --short HEAD)
OS_ARCHS		:=darwin:amd64 darwin:arm64 linux:amd64 linux:arm64

LDFLAGS := -s -w -X "$(MODULE)/version.Version=$(VERSION)" -X "$(MODULE)/version.CommitSHA=$(VERSION_HASH)"

## build: Build
.PHONY: build
build: | ; $(info $(M) building…)
	$Q CGO_ENABLED=0 $(GO) build -ldflags '$(LDFLAGS)' -o .

## build-all: Build all
.PHONY: build-all
build-all: | ; $(info $(M) building all…)
	@$(foreach n, $(OS_ARCHS),\
		os=$(shell echo "$(n)" | cut -d : -f 1);\
		arch=$(shell echo "$(n)" | cut -d : -f 2);\
		gomips=$(shell echo "$(n)" | cut -d : -f 3);\
		target_suffix=$${os}_$${arch};\
		env CGO_ENABLED=0 GOOS=$${os} GOARCH=$${arch} GOMIPS=$${gomips} go build -trimpath -ldflags "$(LDFLAGS)" -o ./release/sdos_$${target_suffix};\
	)

## help: Show this help
.PHONY: help
help:
	@sed -n 's/^## //p' $(MAKEFILE_LIST) | column -t -s ':' |  sed -e 's/^/ /' | sort
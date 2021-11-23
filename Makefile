export PATH := $(GOPATH)/bin:$(PATH)
export GO111MODULE=on
LDFLAGS := -s -w

COMMIT_SHA1   	:= $(shell git rev-parse --short HEAD || echo "0.0.0")
BUILD_TIME      := $(shell date "+%F %T")

LDFLAGS := -s -w -X 'main.CommitSha1=${COMMIT_SHA1}' -X 'main.BuildTime=${BUILD_TIME}'

os-archs=linux:amd64 linux:arm64
all:
	@$(foreach n, $(os-archs),\
		os=$(shell echo "$(n)" | cut -d : -f 1);\
		arch=$(shell echo "$(n)" | cut -d : -f 2);\
		gomips=$(shell echo "$(n)" | cut -d : -f 3);\
		target_suffix=$${os}_$${arch};\
		echo "Build $${os}-$${arch}...";\
		env CGO_ENABLED=0 GOOS=$${os} GOARCH=$${arch} GOMIPS=$${gomips} go build -trimpath -ldflags "$(LDFLAGS)" -o ./release/sdos_$${target_suffix};\
		echo "Build $${os}-$${arch} done";\
	)

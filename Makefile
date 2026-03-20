VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS := -s -w -X main.version=$(VERSION) -X main.buildTime=$(BUILD_TIME)

.PHONY: build build-all install clean

build:
	go build -ldflags "$(LDFLAGS)" -o bin/crabclaw ./cmd/crabclaw/

build-all:
	GOOS=darwin  GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o dist/crabclaw-darwin-arm64 ./cmd/crabclaw/
	GOOS=darwin  GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o dist/crabclaw-darwin-amd64 ./cmd/crabclaw/
	GOOS=linux   GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o dist/crabclaw-linux-amd64  ./cmd/crabclaw/
	GOOS=linux   GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o dist/crabclaw-linux-arm64  ./cmd/crabclaw/
	GOOS=windows GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o dist/crabclaw-windows-amd64.exe ./cmd/crabclaw/

install:
	go install -ldflags "$(LDFLAGS)" ./cmd/crabclaw/

clean:
	rm -rf bin/ dist/

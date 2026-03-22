VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS := -s -w -X main.version=$(VERSION) -X main.buildTime=$(BUILD_TIME)

.PHONY: build build-all install clean

build:
	go build -ldflags "$(LDFLAGS)" -o bin/crabclaw-skill ./cmd/crabclawskill/

build-all:
	GOOS=darwin  GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o dist/crabclaw-skill-darwin-arm64 ./cmd/crabclawskill/
	GOOS=darwin  GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o dist/crabclaw-skill-darwin-amd64 ./cmd/crabclawskill/
	GOOS=linux   GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o dist/crabclaw-skill-linux-amd64  ./cmd/crabclawskill/
	GOOS=linux   GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o dist/crabclaw-skill-linux-arm64  ./cmd/crabclawskill/
	GOOS=windows GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o dist/crabclaw-skill-windows-amd64.exe ./cmd/crabclawskill/

install:
	go install -ldflags "$(LDFLAGS)" ./cmd/crabclawskill/

clean:
	rm -rf bin/ dist/

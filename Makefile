BINARY := tpm
PKG := github.com/zhangshuaike/tpm
VERSION ?= 0.1.0
LDFLAGS := -s -w -X $(PKG)/cmd.version=$(VERSION)
DIST := dist

.PHONY: all build test vet fmt clean install release

all: build

## build: compile the binary for the current platform into ./bin
build:
	@mkdir -p bin
	go build -ldflags "$(LDFLAGS)" -o bin/$(BINARY) .

## test: run all unit tests
test:
	go test ./...

## vet: run go vet
vet:
	go vet ./...

## fmt: format all Go sources
fmt:
	go fmt ./...

## install: install the binary into GOBIN/GOPATH
install:
	go install -ldflags "$(LDFLAGS)" .

## clean: remove build artifacts
clean:
	rm -rf bin $(DIST)

## release: cross-compile single binaries for all supported platforms
release: clean
	@mkdir -p $(DIST)
	GOOS=darwin  GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o $(DIST)/$(BINARY)-darwin-amd64 .
	GOOS=darwin  GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o $(DIST)/$(BINARY)-darwin-arm64 .
	GOOS=linux   GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o $(DIST)/$(BINARY)-linux-amd64 .
	GOOS=linux   GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o $(DIST)/$(BINARY)-linux-arm64 .
	GOOS=windows GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o $(DIST)/$(BINARY)-windows-amd64.exe .
	GOOS=windows GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o $(DIST)/$(BINARY)-windows-arm64.exe .
	@echo "Built binaries in $(DIST)/"

SHELL := /usr/bin/env bash

APP := git-sweep
BIN := bin

.PHONY: all build test lint clean fmt

all: build

build:
	go build -o $(BIN)/$(APP) ./cmd/$(APP)

fmt:
	gofumpt -l -w . || true

lint:
	golangci-lint run

test:
	go test ./...

clean:
	rm -rf $(BIN)

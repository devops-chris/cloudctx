.PHONY: build clean test lint install check

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
DATE ?= $(shell date -u '+%Y-%m-%dT%H:%M:%SZ')

LDFLAGS := -s -w \
	-X main.version=$(VERSION) \
	-X main.commit=$(COMMIT) \
	-X main.date=$(DATE)

build:
	go build -ldflags "$(LDFLAGS)" -o cloudctx .

install: build
	mv cloudctx $(GOPATH)/bin/

test:
	go test -v ./...

lint:
	golangci-lint run

clean:
	rm -f cloudctx
	rm -rf dist/

deps:
	go mod download
	go mod tidy

release-dry:
	goreleaser release --snapshot --clean

# Run all checks before pushing (lint + test + build)
check: lint test build
	@echo "All checks passed!"


BINARY   := portwatch
VERSION  := $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS  := -ldflags "-X main.version=$(VERSION)"
GO       := go

.PHONY: all build test lint clean run install vet

all: build

build:
	$(GO) build $(LDFLAGS) -o bin/$(BINARY) ./cmd/portwatch

test:
	$(GO) test ./...

test-verbose:
	$(GO) test -v ./...

test-race:
	$(GO) test -race ./...

vet:
	$(GO) vet ./...

lint:
	golangci-lint run ./...

clean:
	rm -rf bin/

run: build
	./bin/$(BINARY) -config internal/config/example.yaml

install: build
	cp bin/$(BINARY) $(GOPATH)/bin/$(BINARY)

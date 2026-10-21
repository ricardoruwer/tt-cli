SOURCES := $(shell find . -name '*.go' -not -path './bin/*') go.mod go.sum

.PHONY: build test

build: bin/tt

bin/tt: $(SOURCES)
	go build -o bin/tt ./cmd/tt

test:
	go test ./...

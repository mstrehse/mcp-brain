# Makefile for brain mcp

.PHONY: all build install run test clean lint lint-install

all: build

build:
	go build

install:
	go install

run:
	go run main.go

test:
	go test ./...

clean:
	rm -f mcp-brain

lint:
	./golangci-lint run

lint-install:
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh | sh -s -- -b ./ v2.2.1

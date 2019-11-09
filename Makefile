.PHONY: all build test clean run deps

BINARY=build/eventstore

all: test build

build:
	go build -o $(BINARY) -v ./cmd/eventstore

test:
	go test -v ./...

integration-tests: build
	./integration-tests/run.sh

clean:
	rm -f $(BINARY)

run: build
	./$(BINARY)

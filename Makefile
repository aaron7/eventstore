.PHONY: all build test clean run deps

BINARY=build/eventstore

all: test build

build:
	go build -o $(BINARY) -v ./cmd/eventstore

test:
	go test -v ./...

clean:
	go clean
	rm -f $(BINARY)

run: build
	./$(BINARY)

dep:
	go get -u github.com/golang/dep/cmd/dep
	dep ensure
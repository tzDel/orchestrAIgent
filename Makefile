.PHONY: build test run clean

build:
	go build -o bin/agent-manager-mcp ./cmd/server

test:
	go test -v ./...

test-integration:
	go test -v ./... -tags=integration

run:
	go run ./cmd/server

clean:
	rm -rf bin/
	go clean

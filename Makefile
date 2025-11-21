.PHONY: build build-exe test run clean deps test-mcp test-cover test-race test-bench test-all test-layers test-script inspector help

help:
	@echo "Available targets:"
	@echo "  make deps          - Download and tidy dependencies"
	@echo "  make test          - Run all tests"
	@echo "  make test-script   - Run PowerShell test script (Windows)"
	@echo "  make test-mcp      - Run MCP server tests only"
	@echo "  make test-cover    - Run tests with coverage report"
	@echo "  make test-race     - Run tests with race detector"
	@echo "  make test-bench    - Run benchmark tests"
	@echo "  make build         - Build the server binary"
	@echo "  make build-exe     - Build Windows executable (.exe)"
	@echo "  make run           - Run the server in development mode"
	@echo "  make inspector     - Run server with MCP Inspector"
	@echo "  make clean         - Clean build artifacts"

deps:
	go mod tidy

build:
	go build -o bin/orchestrAIgent ./cmd/server

build-exe:
	go build -o bin/orchestrAIgent.exe ./cmd/server

test:
	go test -v ./...

test-script:
	powershell -ExecutionPolicy Bypass -File test_mcp.ps1

test-mcp:
	go test -v ./internal/infrastructure/mcp/...

test-cover:
	go test -cover ./...

test-race:
	go test -race ./...

test-bench:
	go test -bench=. ./...

run:
	go run ./cmd/server

inspector:
	npx @modelcontextprotocol/inspector go run ./cmd/server/main.go

clean:
	rm -rf bin/
	rm -f coverage.out
	go clean

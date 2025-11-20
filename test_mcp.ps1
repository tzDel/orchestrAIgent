#!/usr/bin/env pwsh

Write-Host "====================================" -ForegroundColor Cyan
Write-Host "  MCP Server Setup and Test Suite  " -ForegroundColor Cyan
Write-Host "====================================" -ForegroundColor Cyan
Write-Host ""

Write-Host "Step 1: Downloading dependencies..." -ForegroundColor Yellow
go mod tidy
if ($LASTEXITCODE -ne 0) {
    Write-Host "[FAILED] Could not download dependencies" -ForegroundColor Red
    exit 1
}
Write-Host "[SUCCESS] Dependencies downloaded" -ForegroundColor Green
Write-Host ""

Write-Host "Step 2: Running domain layer tests..." -ForegroundColor Yellow
go test -v ./internal/domain/...
if ($LASTEXITCODE -ne 0) {
    Write-Host "[FAILED] Domain tests failed" -ForegroundColor Red
    exit 1
}
Write-Host "[SUCCESS] Domain tests passed" -ForegroundColor Green
Write-Host ""

Write-Host "Step 3: Running application layer tests..." -ForegroundColor Yellow
go test -v ./internal/application/...
if ($LASTEXITCODE -ne 0) {
    Write-Host "[FAILED] Application tests failed" -ForegroundColor Red
    exit 1
}
Write-Host "[SUCCESS] Application tests passed" -ForegroundColor Green
Write-Host ""

Write-Host "Step 4: Running infrastructure layer tests..." -ForegroundColor Yellow
go test -v ./internal/infrastructure/...
if ($LASTEXITCODE -ne 0) {
    Write-Host "[FAILED] Infrastructure tests failed" -ForegroundColor Red
    exit 1
}
Write-Host "[SUCCESS] Infrastructure tests passed" -ForegroundColor Green
Write-Host ""

Write-Host "Step 5: Running MCP adapter tests..." -ForegroundColor Yellow
go test -v ./internal/adapters/mcp/...
if ($LASTEXITCODE -ne 0) {
    Write-Host "[FAILED] MCP adapter tests failed" -ForegroundColor Red
    exit 1
}
Write-Host "[SUCCESS] MCP adapter tests passed" -ForegroundColor Green
Write-Host ""

Write-Host "Step 6: Running full test suite..." -ForegroundColor Yellow
go test ./...
if ($LASTEXITCODE -ne 0) {
    Write-Host "[FAILED] Full test suite failed" -ForegroundColor Red
    exit 1
}
Write-Host "[SUCCESS] All tests passed" -ForegroundColor Green
Write-Host ""

Write-Host "Step 7: Checking test coverage..." -ForegroundColor Yellow
go test -cover ./...
if ($LASTEXITCODE -ne 0) {
    Write-Host "[FAILED] Coverage check failed" -ForegroundColor Red
    exit 1
}
Write-Host ""

Write-Host "====================================" -ForegroundColor Cyan
Write-Host "     ALL TESTS PASSED!             " -ForegroundColor Green
Write-Host "====================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "Next steps:" -ForegroundColor Yellow
Write-Host '  1. Build server: go build -o agent-manager-mcp.exe cmd/server/main.go' -ForegroundColor White
Write-Host '  2. Run server:   go run cmd/server/main.go' -ForegroundColor White
Write-Host '  3. Test with Inspector: npx @modelcontextprotocol/inspector go run cmd/server/main.go' -ForegroundColor White
Write-Host ""

# orchestrAIgent

An MCP (Model Context Protocol) server that manages isolated git worktrees for AI coding agents.

## Overview

This server allows AI agents (Copilot, Claude, GPT, Gemini, etc.) to work in isolated git worktrees, preventing conflicts and enabling safe parallel development.

## Quick Start

### Install Dependencies

```bash
make deps
```

### Run Tests

```bash
make test
```

### Build and Run

```bash
make build
make run
```

### View All Commands

```bash
make help
```

## Available Commands

| Command | Description |
|---------|-------------|
| `make deps` | Download and tidy dependencies |
| `make test` | Run all tests |
| `make test-all` | Run comprehensive test suite (deps + layer tests + coverage) |
| `make test-layers` | Run tests layer by layer (domain, application, infrastructure) |
| `make test-script` | Run PowerShell test script (Windows) |
| `make test-mcp` | Run MCP server tests only |
| `make test-cover` | Run tests with coverage report |
| `make test-race` | Run tests with race detector |
| `make test-bench` | Run benchmark tests |
| `make build` | Build the server binary |
| `make build-exe` | Build Windows executable (.exe) |
| `make run` | Run the server in development mode |
| `make inspector` | Run server with MCP Inspector |
| `make clean` | Clean build artifacts |

## Architecture

See [docs/architecture.md](docs/architecture.md) for detailed architecture documentation.

## MVP Features

- **Create Worktree**: Create isolated git worktree and branch for an agent
- **Run Tests**: Execute test command in agent's isolated worktree
- **Get Agent Info**: Retrieve agent status, worktree path, and branch information
- **Merge to Main**: Merge agent changes back to base branch
- **Cleanup Worktree**: Remove agent worktree and branch

## Configure with MCP Clients

### For Copilot CLI

1. Build the server:
   ```bash
   make build-exe
   ```

2. Configure in your Copilot settings

### For Claude Code

1. Build the server:
   ```bash
   make build-exe
   ```

2. Add mcp 
```shell
claude mcp add --scope project --transport stdio orchestrAIgent -- .\bin\orchestrAIgent.exe
```

## Project Status

Currently in development. See [docs/concept.md](docs/concept.md) for vision and roadmap.

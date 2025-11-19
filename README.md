# Agent Manager MCP Server

An MCP (Model Context Protocol) server that manages isolated git worktrees for AI coding agents.

## Overview

This server allows AI agents (Copilot, Claude, GPT, Gemini, etc.) to work in isolated git worktrees, preventing conflicts and enabling safe parallel development.

## Architecture

See [docs/architecture.md](docs/architecture.md) for detailed architecture documentation.

## MVP Features

- **Create Worktree**: Create isolated git worktree and branch for an agent
- **Run Tests**: Execute test command in agent's isolated worktree
- **Get Agent Info**: Retrieve agent status, worktree path, and branch information
- **Merge to Main**: Merge agent changes back to base branch
- **Cleanup Worktree**: Remove agent worktree and branch

## Configuration

Copy `config/config.example.yaml` to `config.yaml` and configure:

```yaml
repoRoot: "/path/to/your/repository"
baseBranch: "main"
testCommand: "go test ./..."
worktreeDir: ".worktrees"
```

## Building

```bash
make build
```

## Running

```bash
make run
```

## Testing

```bash
make test
```

## Project Status

Currently in development. See [docs/concept.md](docs/concept.md) for vision and roadmap.

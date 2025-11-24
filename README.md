# orchestrAIgent

An MCP (Model Context Protocol) server that manages isolated git worktrees for AI coding agents.

## Overview

This server allows AI agents (Copilot, Claude, GPT, Gemini, etc.) to work in isolated git worktrees, preventing conflicts and enabling safe parallel development.

## Architecture

See [docs/architecture.md](docs/architecture.md) for detailed architecture documentation.

## Features

- **Create Worktree**: Create isolated git worktree and branch for an agent
  ```
  Example: "Create a worktree for session abc123"
  ```

- **Remove Session**: Remove agent worktree and branch
  ```
  Example: "Remove session abc123"
  ```

- **Get Session Info**: Retrieve agent status, worktree path, and branch information
  ```
  Example: "Get info for session abc123"
  ```

- **List Sessions**: Retrieve agent status, worktree path, and branch information
  ```
  Example: "List all sessions"
  ```

## Configure with MCP Clients

### For Claude Code

1. Download the correct build for your platform from the [releases page](https://github.com/tzDel/orchestrAIgent/releases)

2. Add the MCP server:
   ```shell
   claude mcp add --scope project --transport stdio orchestrAIgent -- <path-to-orchestrAIgent-file>
   ```

3. Verify the installation:
   ```shell
   claude mcp list
   ```

## Project Status

Currently in development. See [docs/concept.md](docs/concept.md) for vision and roadmap.

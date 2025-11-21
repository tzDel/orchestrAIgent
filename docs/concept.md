# Concept

## Why

Orchestrating multiple AI agents in parallel is difficult, especially when data compliance limits which AI tools developers can use. Modern AI coding agents (Copilot, Claude, GPT, Gemini, etc.) often work directly in your main repository with broad filesystem access, making it:

- **Hard to isolate**: Agents can interfere with each other when working on the same codebase
- **Risky for compliance**: Giving agents unrestricted filesystem access violates data governance policies
- **Difficult to parallelize**: Running multiple agents simultaneously without isolation causes conflicts
- **Challenging to control**: No standardized way to spawn, monitor, and manage agent lifecycles

This project provides a small local MCP server that:
- **MVP**: Gives each agent its own git worktree, limiting their filesystem access to an isolated workspace
- **Future**: Orchestrates agent processes, spawning and monitoring multiple AI agents in parallel, each with a dedicated worktree and restricted tool set for compliance-safe operation

## Vision

### Immediate Use: MCP-Compatible CLI Tools

The MCP server is immediately usable with existing MCP-compatible tools like **GitHub Copilot CLI**, **Claude Code CLI**, or any MCP client:

**Workflow with Copilot CLI:**
1. Developer runs: `copilot "add user authentication to the API"`
2. Copilot calls `create_worktree("copilot-auth")` via MCP → gets isolated worktree path
3. Copilot works in that worktree, makes changes, commits
4. Copilot calls `run_tests("copilot-auth")` → sees test results
5. Developer manually reviews: `cd .worktrees/agent-copilot-auth && git diff`
6. Developer manually merges: `git checkout main && git merge agent-copilot-auth`

Or the developer can ask Copilot to call `merge_to_main("copilot-auth")` after reviewing the changes.

### Future Vision: Agent Orchestration

The long-term goal is to enable the MCP server to **spawn and manage AI agent processes** directly, orchestrating their work in isolated worktrees with full lifecycle control.

#### Workflow with CLI Tool
1. **Developer runs**: `orchestraigent spawn copilot "Add user authentication to the API"`
2. **MCP server**:
   - Creates dedicated worktree `.worktrees/agent-copilot-auth`
   - Spawns Copilot process configured for that worktree
   - Returns agent ID and PID
3. **Developer polls**: `orchestraigent status copilot-auth` → sees "testing (5 files changed)"
4. **Agent works**: Autonomously makes changes, commits, runs tests in isolation
5. **Developer checks**: `orchestraigent status copilot-auth` → sees "ready_for_review"
6. **Developer reviews**: `cd .worktrees/agent-copilot-auth && git diff && git log`
7. **Developer merges**: `git checkout main && git merge agent-copilot-auth`
8. **Developer cleans up**: `orchestraigent cleanup copilot-auth` → terminates process, removes worktree

The CLI tool acts as an **MCP client**, wrapping MCP protocol calls in user-friendly commands for agent orchestration.

#### Workflow with IDE Plugin
1. **Developer clicks**: "New Agent Task" button in IDE sidebar
2. **Developer enters**: Task description: "Add user authentication to the API"
3. **IDE calls**: `spawn_agent({type: "copilot", task: "Add user auth", agentId: "copilot-auth"})` via MCP
4. **MCP server**:
   - Creates dedicated worktree `.worktrees/agent-copilot-auth`
   - Spawns Copilot process configured for that worktree
   - Monitors agent progress and status
5. **IDE shows**: "🤖 Copilot working on: Add user auth" with live status updates
6. **Agent works**: Autonomously makes changes, commits, runs tests in isolation
7. **IDE notifies**: "✅ Copilot finished: Add user auth" when ready for review
8. **Developer reviews**: Switches to agent worktree in editor, views diffs natively
9. **Developer clicks**: "Merge to Main" → IDE calls `merge_to_main("copilot-auth")` via MCP
10. **Cleanup**: MCP terminates agent process and removes worktree

The IDE plugin acts as an **orchestration layer**, providing rich UI on top of the same MCP protocol used by the CLI. The MCP server handles both worktree lifecycle AND agent process management, providing full isolation and control.

---

## Capabilities

### MVP

- Single-repo support (configured main repo and base branch)
- Create an isolated git worktree and branch per agent
- Track minimal agent state (worktree path, branch, basic status)
- Automatic worktree cleanup (remove worktree/branch when agent work is discarded)
- Richer git info (diffs, commits, changed files) for manual review

### Stretch Goals

#### Phase 1: Orchestration & IDE Integration
- **Agent Process Management** (spawn, monitor, terminate agent instances)
  - `spawn_agent()` - Launch Copilot/Claude/GPT process for a worktree
  - Process monitoring with status updates
  - Support for multiple agent types (Copilot, Claude, GPT, Gemini, etc.)
- **IDE Plugin Integration** (VSCode/IntelliJ extensions)
  - Orchestration UI: "New Agent Task" button
  - Live agent status sidebar
  - Worktree switching and review UI
  - One-click merge with conflict resolution
- **Merge to main** (merge agent branch back to base branch - most valuable with IDE plugin)
- Agent metadata tracking (task description, status, progress)

#### Phase 2: Enhanced Features
- **CLI Tool** (`orchestraigent` command) for orchestration and debugging
  - `orchestraigent spawn <type> <task>` - Spawn agent process with task description
  - `orchestraigent status <agentId>` - Check agent status and progress
  - `orchestraigent list` - List all active agents
  - `orchestraigent logs <agentId>` - View agent process logs
  - `orchestraigent cleanup <agentId>` - Terminate agent and remove worktree
  - `orchestraigent kill <agentId>` - Force-kill stuck agent process
- Better review/merge experience (summaries, conflict info, merge policies)
- Persistence (DB/files), logging, metrics
- Multi-repo support with repository selector

---

## Tech Stack

### Core Language and Runtime

- **Go (Golang)** as the primary implementation language
  - **Rationale**: Multi-client compatibility is the priority
    - Single binary distribution (no runtime dependencies)
    - Works equally well with VSCode (TypeScript), IntelliJ (Java/Kotlin), and CLI tools
    - Client-agnostic: MCP protocol abstracts language differences
    - Cross-platform builds (Windows/Mac/Linux) without complexity
    - Fast git subprocess operations with low overhead
    - Excellent process management primitives for spawning/monitoring agents
- Go standard library for core behavior:
  - Process execution (`os/exec`) – running `git`, test commands, and agent processes
  - Process management (`os.Process`, signals) – monitoring and terminating agents
  - Filesystem operations (`os`, `path/filepath`) – managing worktree paths
  - JSON handling (`encoding/json`) – config, responses, MCP payloads
  - Basic networking / IPC as needed for MCP transport

### MCP Integration

- Custom MCP tooling layer built on:
  - Go stdlib (`encoding/json`, I/O primitives, possibly `net/http` if the MCP transport uses HTTP)
- Responsibilities:
  - Define and expose MCP tools for:
    - **MVP Tools**:
      - Creating worktrees (returns absolute paths for IDE to open)
      - Running tests (returns structured results for IDE display)
      - Querying agent status and file changes (for IDE sidebar/diff views)
      - Cleaning up worktrees (removes worktree and branch)
    - **Stretch Tools (Phase 1)**:
      - Spawning agents (`spawn_agent()` - launches and configures agent process)
      - Monitoring agents (status updates, progress tracking)
      - Merging branches (human-triggered via IDE plugin UI)
      - Terminating agents (kills agent process and cleans up)
  - Map MCP requests/responses to internal Go functions
- Design note: All destructive operations (merge, cleanup, agent spawn/terminate) are intended to be triggered by human actions in the IDE plugin, not autonomous agent calls

### Configuration and Logging

- Configuration:
  - Simple file-based config (e.g., YAML/JSON) for:
    - **MVP**: `repoRoot`, `baseBranch`, `testCommand`
    - **Phase 1 additions**: Agent configurations
      - Agent executable paths (`copilotPath`, `claudePath`, etc.)
      - Agent-specific settings (context window, model versions, etc.)
      - Default prompts and constraints for spawned agents
  - Optionally read environment variables for overrides
- Logging:
  - Initially: Go's standard `log` package
  - Later: Structured logger with levels (DEBUG, INFO, WARN, ERROR)
  - Agent process logs captured and queryable via MCP

### Git Integration

- Shelling out to `git` via `os/exec`:
  - `git worktree add/remove`
  - `git status`, `git diff`, `git commit`, `git merge`, etc.
- No external git library required for the MVP; relying on the git CLI keeps behavior consistent with normal developer workflows

### Optional / Stretch Libraries

- Persistence (for stretch goals):
  - SQLite driver (e.g., `modernc.org/sqlite` or `github.com/mattn/go-sqlite3`)
  - Or simple JSON/YAML file storage using stdlib
- **CLI wrapper** (Phase 2 - `orchestraigent` command):
  - `github.com/spf13/cobra` for structured commands:
    - `orchestraigent spawn <type> <task>` - Spawn agent with task
    - `orchestraigent status <agentId>` - Check agent status
    - `orchestraigent list` - List active agents
    - `orchestraigent logs <agentId>` - View agent logs
    - `orchestraigent cleanup <agentId>` - Cleanup agent
    - `orchestraigent kill <agentId>` - Force-kill agent
  - CLI acts as MCP client, wrapping MCP protocol calls in user-friendly commands
- Optional REST/debug interface (if desired later):
  - Based on `net/http` from stdlib
  - Lightweight router (e.g., `github.com/go-chi/chi/v5`) can be added later for nicer routing, but is not required for the MVP
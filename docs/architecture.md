# Architecture

## System Overview

```mermaid
graph TB
    subgraph "MCP Clients"
        CLI["Direct MCP Usage<br/>(Scripts, CLI tools)"]
        IDE["IDE Plugins<br/>(VSCode, IntelliJ)<br/><b>Orchestration Layer</b>"]
    end

    subgraph "MCP Server (Go)"
        MCP[MCP Protocol Handler]
        AM[Agent Manager<br/><i>spawns & monitors agents</i>]
        WM[Worktree Manager]
        TM[Test Manager]
        GM[Git Manager]
        SM[State Manager]
    end

    subgraph "Agent Processes"
        A1[Copilot Instance 1]
        A2[Claude Instance 1]
        AN[Agent Instance N]
    end

    subgraph "Git Repository"
        MAIN[Main Branch]
        WT1[Agent Worktree 1]
        WT2[Agent Worktree 2]
        WTN[Agent Worktree N]
    end

    CLI -->|MCP Protocol| MCP
    IDE -->|MCP Protocol<br/><b>spawn_agent</b>| MCP
    MCP --> AM
    MCP --> WM
    MCP --> TM
    MCP --> SM
    AM -.->|spawns & monitors| A1
    AM -.->|spawns & monitors| A2
    AM -.->|spawns & monitors| AN
    WM --> GM
    TM --> GM
    A1 -->|works in| WT1
    A2 -->|works in| WT2
    AN -->|works in| WTN
    GM -->|git commands| MAIN
    GM -->|create/manage| WT1
    GM -->|create/manage| WT2
    GM -->|create/manage| WTN

    style MCP fill:#4CAF50
    style AM fill:#FF5722
    style IDE fill:#FF9800
    style A1 fill:#E91E63
    style A2 fill:#E91E63
    style AN fill:#E91E63
```

**Flow:** Client calls MCP tool → Server orchestrates use case → Domain logic + Infrastructure execute → Return result

## Workflows

**MVP (Current):**
1. MCP client calls `create_worktree(sessionId)` → Server creates worktree + branch
2. Agent works in isolated worktree
3. Developer reviews manually: `cd .worktrees/agent-{sessionId} && git diff`
4. Developer merges manually
5. Cleanup with `remove_session(sessionId)`

**Future (Orchestration):**
1. CLI/IDE calls `spawn_agent()` → Server creates worktree + spawns agent process
2. Agent works autonomously, server monitors status
3. Developer polls status via CLI/IDE
4. Developer reviews in IDE, merges via UI
5. Server terminates agent process and cleans up

## Design Principles

- **Protocol-First:** All functionality via MCP tools, usable by any client
- **Human-in-the-Loop:** Critical operations require human approval
- **Process Isolation:** Each agent gets dedicated worktree + process
- **Progressive Enhancement:** MVP → Full orchestration

---

## Clean Architecture

### Layer Organization

Dependencies flow **inward** toward the domain:

```
Adapters (MCP handlers, CLI)
    ↓ depends on
Application (Use Cases: CreateWorktree, RunTests, MergeToMain)
    ↓ depends on
Domain (Entities: Agent, Worktree | Interfaces: GitOperations, AgentRepository)
    ↑ implemented by
Infrastructure (GitClient, InMemoryRepository, ProcessManager)
```

**Rules:**
- Domain has zero external dependencies
- Application depends only on domain interfaces
- Infrastructure implements domain interfaces
- Adapters translate external protocols to use cases

### Directory Structure

```
orchestrAIgent/
├── cmd/
│   └── server/main.go              # MCP server entrypoint
├── internal/
│   ├── domain/                     # Pure business logic (zero deps)
│   │   ├── agent.go, worktree.go
│   │   ├── value_objects.go        # AgentID, WorktreePath, etc.
│   │   └── ports.go                # Interfaces
│   ├── application/                # Use cases
│   │   ├── create_worktree.go
│   │   ├── run_tests.go
│   │   └── merge_to_main.go
│   ├── infrastructure/             # Implements interfaces
│   │   ├── git/git_client.go
│   │   └── persistence/in_memory_repository.go
│   └── adapters/                   # External protocols
│       └── mcp/
│           ├── server.go
│           └── tools/              # MCP tool handlers
├── test/integration/
├── config/config.example.yaml
└── docs/
```

# orchestrAIgent - AI Agent Instructions

**Service:** MCP Server (Model Context Protocol) <br>
**Capabilities:** Git worktree isolation, agent lifecycle management, test execution, conflict-free merging <br>
**Tech Stack:** Go, Clean Architecture, Git CLI, MCP SDK, YAML/JSON config

---

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is an MCP (Model Context Protocol) server that manages isolated git worktrees for AI coding agents. It allows each agent (Copilot, Claude, GPT, Gemini, etc.) to work in their own worktree, run tests in isolation, and merge approved changes back to the main branch.

### Core Concepts

1. **Agent Isolation**: Each agent gets its own git worktree and branch (pattern: `agent-{agentId}`)
2. **Worktree Management**: Manages creation, testing, and merging of agent-specific worktrees
3. **Test Execution**: Runs configurable test commands in isolated agent worktrees
4. **Merge Control**: Merges agent branches back to base branch with conflict handling

---

# 1) Guardrails and behavioral guidelines

- **Test before commit:** Always run `go test ./...` before committing changes; all tests must pass
- **Strict layer boundaries:** Domain layer has zero external dependencies; application depends only on domain interfaces; infrastructure implements domain interfaces
- **Test coverage:** All new features require tests; follow TDD (Red-Green-Refactor) cycle
- **Non-deterministic solutions PROHIBITED:** NO timeouts, delays, sleep, retry loops, or polling in application logic or test code (operational safeguards for CLI commands are acceptable)
- **Code cleanup policy:** Remove unused code completely; no backwards-compatibility hacks (rename unused vars, re-export types, `// removed` comments)
- **Readability first:** Use descriptive variable names (not abbreviations); prefer self-documenting code; add comments only to explain WHY (intent/rationale), not WHAT
- **When in doubt:** Ask questions using AskUserQuestion tool; clarify architectural decisions before implementing

## Working Directory (CRITICAL)

**Your workspace is the active working directory. Do not leave it.**

- Inspect `<env>` to determine your working directory and the current branch
- If you are in a worktree (branch: `orchestrAIgent/*`): edit all files here
- If you are in the main repository (branch: `main`): make changes here directly
- Do NOT infer or rely on parent paths from the directory structure
- Do NOT switch directories with `cd` to other locations unless explicitly required

❌ WRONG: `cd /inferred/path && command`
✅ RIGHT: `command` (executed in the current directory)

---

# 2) Repository architecture summary

## Package/Directory Structure

```
internal/
├── domain/               # Core business logic (zero external dependencies)
│   ├── agent.go         # Agent aggregate root with lifecycle management
│   ├── agent_id.go      # AgentID value object with validation
│   └── ports.go         # Domain interfaces (GitOperations, AgentRepository)
│
├── application/          # Use cases and orchestration
│   └── create_worktree.go  # CreateWorktreeUseCase implementation
│
├── adapters/             # Inbound adapters (external protocols → application)
│   └── mcp/
│       ├── server.go         # MCP server adapter
│       ├── result.go         # MCP result helpers
│       └── server_test.go    # MCP server tests
│
└── infrastructure/       # Outbound adapters (application → external services)
    ├── git/
    │   └── git_client.go      # GitClient implementing GitOperations
    └── persistence/
        └── in_memory_repository.go  # InMemoryAgentRepository

cmd/
└── server/
    └── main.go          # Entry point and test harness
```

**Layer Dependencies** (strictly enforced):
- Domain: No dependencies (pure business logic)
- Application: Depends only on domain interfaces
- Adapters: Translate external protocols to application use cases (inbound)
- Infrastructure: Implements domain interfaces, depends on external libraries (outbound)
- Flow: Adapters/Infrastructure → Application → Domain

## External Dependencies

- `github.com/modelcontextprotocol/go-sdk` - MCP server implementation and protocol handling
- Git CLI - Worktree and branch operations via shell commands

**Important:** Use `go mod tidy` to manage dependencies; never specify versions manually in imports

## Clean Architecture Pattern

**Flow:** MCP Client → Adapters (MCP Server) → Application (Use Cases) → Domain (Business Logic) → Infrastructure (Git, Persistence)

**Key Points:**
- **Adapters** (inbound) translate external protocols (MCP, REST) to use case calls
- **Use Cases** (application layer) orchestrate domain logic and coordinate infrastructure
- **Domain** contains pure business logic with no external dependencies
- **Infrastructure** (outbound) implements domain interfaces (GitOperations, AgentRepository)
- All dependencies point inward toward the domain layer
- Configuration loaded at startup from YAML/JSON files

## Key Design Decisions

1. **In-Memory State**: Current implementation uses in-memory storage; server restart loses all state
2. **Single Repository**: One configured repository per server instance
3. **Local Merges**: Merges stay local; manual `git push` required to sync with remote
4. **Simple Status Model**: Three agent status values (`created`, `merged`, `failed`)
5. **Clean Architecture**: Strict layer separation (domain → application → infrastructure)
6. **Value Objects**: AgentID enforces validation rules at domain level

---

# 3) Coding guidelines and conventions

## Naming Conventions

Use descriptive names. Never use abbreviations!

```go
// ✅ GOOD (Descriptive names)
agentRepository := NewInMemoryRepository()
worktreePath := filepath.Join(baseDir, agent.ID())
testCommand := config.TestCommand

// ❌ BAD (Abbreviations)
repo := NewInMemoryRepository()
path := filepath.Join(baseDir, agent.ID())
cmd := config.TestCommand
```

**Exception:** Well-established conventions in limited scopes:
- `i` in loops
- `err` for errors
- `ctx` for context.Context

## Error Handling

```go
// ✅ GOOD (Contextual error handling)
if err := gitOps.CreateWorktree(path); err != nil {
    return fmt.Errorf("failed to create worktree for agent %s: %w", agentID, err)
}

// ❌ BAD (Empty or generic error handling)
if err := gitOps.CreateWorktree(path); err != nil {
    return err
}
```

- NEVER use empty catch blocks
- Always wrap errors with context using `fmt.Errorf` with `%w`
- Provide actionable information in error messages

## Comment Style

- Do not use comments to narrate what changed or what is new
- Prefer self-documenting code; only add comments when strictly necessary to explain WHY (intent/rationale), not WHAT
- Keep any necessary comments concise and local to the logic they justify

## Testing Patterns

**Key Rules:**
- Use descriptive test names that clearly describe the scenario being tested
- Every test function MUST contain explicit comment blocks: `// arrange`, `// act`, `// assert`
- Follow table-driven test patterns for multiple scenarios
- Use test helpers to reduce duplication in setup/teardown
- Format: `Test{Component}_{Scenario}_{ExpectedBehavior}`

### Unit Tests

**Test Naming Convention:**
- `TestAgentID_CreateWithInvalidFormat_ReturnsError`
- `TestCreateWorktreeUseCase_WhenAgentAlreadyExists_ReturnsError`
- `TestGitClient_CreateWorktree_CreatesDirectoryAndBranch`

**Test Structure:**
```go
func TestAgentID_CreateWithInvalidFormat_ReturnsError(t *testing.T) {
    // arrange
    invalidID := "invalid id with spaces"

    // act
    _, err := domain.NewAgentID(invalidID)

    // assert
    if err == nil {
        t.Error("expected error for invalid agent ID format")
    }
}
```

### Table-Driven Tests

```go
func TestAgentID_Validation(t *testing.T) {
    tests := []struct {
        name    string
        agentID string
        wantErr bool
    }{
        {name: "valid agent ID", agentID: "copilot", wantErr: false},
        {name: "invalid with spaces", agentID: "invalid id", wantErr: true},
        {name: "invalid empty", agentID: "", wantErr: true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // arrange
            // (test case from table)

            // act
            _, err := domain.NewAgentID(tt.agentID)

            // assert
            if (err != nil) != tt.wantErr {
                t.Errorf("NewAgentID() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

### TDD Workflow (MANDATORY)

Always write tests first, before implementing features:
1. **Red**: Write a failing test that describes the desired behavior
2. **Green**: Write minimal code to make the test pass
3. **Refactor**: Improve the implementation while keeping tests green

**CRITICAL Rules:**
- Test failures are NEVER unrelated - fix immediately
- NEVER skip tests (no `.skip()` in Go, no `t.Skip()` unless external dependency unavailable)
- Fix performance test failures (they indicate real issues)
- After every code change, rerun the full validation suite and report "tests green" before handing work back

---

# 4) Build, test, and tooling conventions

## Build Commands

```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Run tests with coverage
go test -cover ./...

# Build the server
go build -o bin/orchestrAIgent ./cmd/server
```

more commands can be found in [Makefile](Makefile)

## Testing Setup

- **Go testing** for all unit and integration tests
- Test files must be named `*_test.go`
- Place tests in the same package as the code being tested
- Use table-driven tests for multiple scenarios

## Local Development

```bash
# Build and run the MCP server
go run ./cmd/server/main.go

# Run with configuration file (when implemented)
go run ./cmd/server/main.go -config config.yaml
```

**Configuration:** Configuration will be loaded from YAML/JSON files specifying repository path, base branch, and test command

---

# 5) Security, secrets, and data handling

- **No secrets in code:** All configuration (repository paths, branch names) stored in config files, never hardcoded
- **No credentials in commits:** Git operations use local system credentials; never store authentication tokens in code or config
- **Local operations only:** All git operations are local; no automatic push to remote (prevents accidental exposure)

---

# 6) Documentation and onboarding

## Key Reference Files

- **Project overview:** `README.md`
- **Project vision:** `docs/concept.md`
- **System architecture:** `docs/architecture.md`
- **Implementation plans:** `docs/plans/`
- **Domain interfaces:** `internal/domain/ports.go`
- **Use cases:** `internal/application/`
- **MCP server adapter:** `internal/adapters/mcp/server.go`
- **Git operations:** `internal/infrastructure/git/git_client.go`
- **Entry point:** `cmd/server/main.go`

## Project Structure Overview

- `internal/domain/` - Pure business logic with no external dependencies
- `internal/application/` - Use cases that orchestrate domain logic
- `internal/adapters/` - Inbound adapters translating external protocols to use cases
- `internal/infrastructure/` - Outbound adapters implementing domain interfaces
- `cmd/server/` - Application entry point and test harness

## Specification Writing Guidelines

### Technical Specs (MANDATORY)

When creating specs for implementation agents:
- **Focus**: Technical implementation details, architecture, code examples
- **Requirements**: Clear dependencies, APIs, integration points
- **Structure**: Components → Implementation → Configuration → Phases
- **Omit**: Resource constraints, obvious details, verbose explanations
- **Include**: Platform-specific APIs, code snippets, data flows, dependencies

## Plan Files

- Store all plan MD files in the `./docs/plans/` directory, not at the repository root
- This keeps the root clean and organizes planning documents
- If you create plans research the codebase or requested details first before making a plan for the implementation
- Don't make plans for making plans, rather do the planning ahead and then implement